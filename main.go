// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// TODO: This should all move to either config file or command line options
const (
	TOKENSERVER_API_ROOT           = "/tokenserver"
	TOKENSERVER_API_LISTEN_ADDRESS = "0.0.0.0"
	TOKENSERVER_API_LISTEN_PORT    = 8123
	TOKENSERVER_PERSONA_VERIFIER   = "https://verifier.accounts.firefox.com/v2"
	TOKENSERVER_PERSONA_AUDIENCE   = "https://tokenserver.sateh.com"
	TOKENSERVER_ALLOW_NEW_USERS    = true
	TOKENSERVER_TOKEN_DURATION     = 300
	TOKENSERVER_SHARED_SECRET      = "cheesebaconeggs"
	TOKENSERVER_STORAGESERVER      = "http://127.0.0.1:8124/storage"
	TOKENSERVER_DATABASE           = "postgres://tokenserver:tokenserver@localhost/tokenserver"
)

type TokenServerResponse struct {
	Id          string `json:"id"`           // Signed authorization token
	Key         string `json:"key"`          // Secret derived from the shared secret
	Uid         string `json:"uid"`          // The user id for this service
	ApiEndpoint string `json:"api_endpoint"` // The root URL for the user of this service
	Duration    int64  `json:"duration"`     // the validity duration of the issued token, in seconds
}

var clientIdValidator = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,32}$`)

func handleStuff(w http.ResponseWriter, r *http.Request) {
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

	verifier, err := NewVerifier(TOKENSERVER_PERSONA_VERIFIER, TOKENSERVER_PERSONA_AUDIENCE)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Verifying %s", assertion)

	personaResponse, err := verifier.VerifyAssertion(assertion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Done verifying: %+v", personaResponse)

	if personaResponse.Status != "okay" {
		log.Printf("%+v", personaResponse)
		http.Error(w, "Invalid BrowserID assertion", http.StatusUnauthorized)
		return
	}

	// Grab some things we need from the assertion

	generation := 1 // TODO: This needs to be parsed from the assertion

	// Load the user. Create if new and if signups are allowed.

	db, err := NewDatabaseSession(TOKENSERVER_DATABASE)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer db.Close()

	var user *User

	user, err = db.GetUser(personaResponse.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user == nil {
		if TOKENSERVER_ALLOW_NEW_USERS {
			user, err = db.AllocateUser(personaResponse.Email, generation, clientState)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
	}

	log.Printf("User is %s", user.Email)

	// Deal with generation

	newGeneration := 0
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
		user, err = db.UpdateUser(user.Email, newGeneration, newClientState)
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

	expires := time.Now().Unix() + TOKENSERVER_TOKEN_DURATION

	tokenSecret, derivedSecret, err := GenerateSecret(user.Uid, user.NodeId, expires,
		TOKENSERVER_SHARED_SECRET)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// All done, build a response

	tokenServerResponse := &TokenServerResponse{
		Id:          tokenSecret,
		Key:         derivedSecret,
		Uid:         user.Uid,
		ApiEndpoint: fmt.Sprintf("%s/storage/%s", TOKENSERVER_STORAGESERVER, user.NodeId),
		Duration:    TOKENSERVER_TOKEN_DURATION,
	}

	data, err := json.Marshal(tokenServerResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}

func main() {
	http.HandleFunc(TOKENSERVER_API_ROOT+"/1.0/sync/1.5", handleStuff)
	addr := fmt.Sprintf("%s:%d", TOKENSERVER_API_LISTEN_ADDRESS, TOKENSERVER_API_LISTEN_PORT)
	log.Printf("Starting tokenserver server on http://%s%s", addr, TOKENSERVER_API_ROOT)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
