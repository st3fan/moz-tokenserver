// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package tokenserver

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/st3fan/moz-tokenserver/mockmyid"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_SyncTokenHandler(t *testing.T) {
	// Grab an assertion from the mockmyid api
	assertion, err := mockmyid.RequestAssertion("test@mockmyid.com", DEFAULT_PERSONA_AUDIENCE)
	if err != nil {
		t.Error("Could not request assertion", err)
	}
	if len(assertion) == 0 {
		t.Error("Could not create assertion (it is zero length or not returned)")
	}

	// Exchange the assertion for an access token
	request, _ := http.NewRequest("GET", "/tokenserver/1.0/sync/1.5", nil)
	request.Header.Set("Authorization", "BrowserID "+assertion)
	response := httptest.NewRecorder()

	//
	router := mux.NewRouter()
	config := DefaultTokenServerConfig()
	context, err := SetupTokenServerRouter(router.PathPrefix("/tokenserver").Subrouter(), config)
	if err != nil {
		panic("Cannot setup router")
	}
	http.Handle("/", router)

	//
	context.SyncTokenHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", response.Code, response.Body)
	}

	tokenServerResponse := &TokenServerResponse{}
	if err = json.Unmarshal(response.Body.Bytes(), tokenServerResponse); err != nil {
		t.Fatal("Can't unmarshal token server response", err)
	}

	if len(tokenServerResponse.Id) == 0 {
		t.Fatal("Token server did not return Id")
	}

	if len(tokenServerResponse.Key) == 0 {
		t.Fatal("Token server did not return Key")
	}

	if tokenServerResponse.Uid == 0 {
		t.Fatal("Token server did not return Uid")
	}

	if len(tokenServerResponse.ApiEndpoint) == 0 {
		t.Fatal("Token server did not return ApiEndpoint")
	}
	if !strings.HasPrefix(tokenServerResponse.ApiEndpoint, config.StorageServerNode) {
		t.Fatal("Token server did not return expected ApiEndpoint")
	}

	if tokenServerResponse.Duration == 0 {
		t.Fatal("Token server returned zero Duration")
	}
}
