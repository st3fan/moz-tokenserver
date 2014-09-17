package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func init() {
}

func Test_SyncTokenHandler(t *testing.T) {
	// Grab an assertion from the mockmyid api
	assertion, err := RequestAssertion("test@mockmyid.com", DEFAULT_PERSONA_AUDIENCE)
	if err != nil {
		t.Error("Could not request assertion", err)
	}
	if len(assertion) == 0 {
		t.Error("Could not create assertion (it is zero length or not returned)")
	}

	// Exchange the assertion for an access token
	request, _ := http.NewRequest("GET", DEFAULT_API_PREFIX+"/1.0/sync/1.5", nil)
	request.Header.Set("Authorization", "BrowserID "+assertion)
	response := httptest.NewRecorder()

	//
	router := mux.NewRouter()
	router.HandleFunc("/version", VersionHandler)
	config := DefaultTokenServerConfig()
	context, err := SetupTokenServerRouter(router.PathPrefix(DEFAULT_API_PREFIX).Subrouter(), config)
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

	if len(tokenServerResponse.Uid) == 0 {
		t.Fatal("Token server did not return Uid")
	}

	if len(tokenServerResponse.ApiEndpoint) == 0 {
		t.Fatal("Token server did not return ApiEndpoint")
	}
	if !strings.HasPrefix(tokenServerResponse.ApiEndpoint, config.StorageServerUrl) {
		t.Fatal("Token server did not return expected ApiEndpoint")
	}

	if tokenServerResponse.Duration == 0 {
		t.Fatal("Token server returned zero Duration")
	}
}
