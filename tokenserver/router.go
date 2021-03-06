// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package tokenserver

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/st3fan/moz-tokenserver/fxa"
	"github.com/st3fan/moz-tokenserver/token"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type TokenServerResponse struct {
	Id          string `json:"id"`           // Signed authorization token
	Key         string `json:"key"`          // Secret derived from the shared secret
	Uid         uint64 `json:"uid"`          // The user id for this service
	ApiEndpoint string `json:"api_endpoint"` // The root URL for the user of this service
	Duration    int64  `json:"duration"`     // the validity duration of the issued token, in seconds
}

var clientIdValidator = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,32}$`)

type tokenServerContext struct {
	config Config
	db     *DatabaseSession
}

func (c *tokenServerContext) SyncTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Make sure we have a BrowserID Authorization header

	authorization := r.Header.Get("Authorization")
	if len(authorization) == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tokens := strings.Split(authorization, " ")
	if len(tokens) != 2 {
		http.Error(w, "Unsupported authorization method", http.StatusUnauthorized)
		return
	}
	if tokens[0] != "BrowserID" {
		http.Error(w, "Unsupported authorization method", http.StatusUnauthorized)
		return
	}

	assertion := tokens[1]

	// Check if the client state is valid

	clientState := r.Header.Get("X-Client-State")
	if len(clientState) != 0 {
		if !clientIdValidator.MatchString(clientState) {
			http.Error(w, "Invalid X-Client-State", http.StatusInternalServerError) // TODO: JSON Error
			return
		}
	}

	// Verify the assertion

	verifier, err := fxa.NewVerifier(c.config.PersonaVerifier, c.config.PersonaAudience)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	personaResponse, err := verifier.VerifyAssertion(assertion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if personaResponse.Status != "okay" {
		http.Error(w, "Invalid BrowserID assertion: "+personaResponse.Reason, http.StatusUnauthorized)
		return
	}

	// Grab the generation from the assertion. This conveniently
	// defaults to 0 if it is not present.

	generation := personaResponse.Claims.Generation

	// Load the user. Create if new and if signups are allowed.

	var user *User
	user, err = c.db.GetUser(personaResponse.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		if c.config.AllowNewUsers {
			user, err = c.db.AllocateUser(personaResponse.Email, generation, clientState)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
	}

	// Deal with generation

	var newGeneration uint64
	newClientState := ""

	if generation > user.Generation {
		newGeneration = generation
	}

	if clientState != user.ClientState {
		// Don't revert from some-client-state to no-client-state
		if len(clientState) == 0 {
			http.Error(w, "invalid-client-state", http.StatusUnauthorized)
			return
		}
		// Don't revert to a previous client-state
		if user.IsOldClientState(clientState) {
			http.Error(w, "invalid-client-state", http.StatusUnauthorized)
			return
		}
		// If the IdP has been sending generation numbers, then don't
		// update client-state without a change in generation number
		if user.Generation > 0 && newGeneration != 0 {
			http.Error(w, "invalid-client-state", http.StatusUnauthorized)
			return
		}
		newClientState = clientState
	}

	if newGeneration != 0 || len(newClientState) != 0 {
		user, err = c.db.UpdateUser(user.Email, newGeneration, newClientState)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Error out if this client is behind some previously-seen
	// client. This is done after the updates because some other, even
	// more up-to-date client may have raced with a concurrent update.

	if user.Generation > generation {
		http.Error(w, "invalid-generation", http.StatusUnauthorized)
		return
	}

	// Finally, create token and secret

	expires := time.Now().Unix() + c.config.TokenDuration

	payload := token.TokenPayload{
		Uid:     user.Uid,
		Node:    c.config.StorageServerNode,
		Expires: expires,
	}

	token, err := token.NewToken([]byte(c.config.SharedSecret), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// All done, build a response

	tokenServerResponse := &TokenServerResponse{
		Id:          token.Token,
		Key:         token.DerivedSecret,
		Uid:         user.Uid,
		ApiEndpoint: fmt.Sprintf("%s/1.5/%d", c.config.StorageServerNode, user.Uid),
		Duration:    c.config.TokenDuration,
	}

	data, err := json.Marshal(tokenServerResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

func SetupRouter(r *mux.Router, config Config) (*tokenServerContext, error) {
	db, err := NewDatabaseSession(config.DatabasePath)
	if err != nil {
		return nil, err
	}

	context := &tokenServerContext{config: config, db: db}
	r.HandleFunc("/1.0/sync/1.5", context.SyncTokenHandler)

	return context, nil
}
