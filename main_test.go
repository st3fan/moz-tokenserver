package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	setupHandlers()
}

func Test_handleStuff(t *testing.T) {
	// Grab an assertion from the mockmyid api
	assertion, err := RequestAssertion("test@mockmyid.com", TOKENSERVER_PERSONA_AUDIENCE)
	if err != nil {
		t.Error("Could not request assertion", err)
	}
	if len(assertion) == 0 {
		t.Error("Could not create assertion (it is zero length or not returned)")
	}

	// Exchange the assertion for an access token
	request, _ := http.NewRequest("GET", TOKENSERVER_API_ROOT+"/1.0/sync/1.5", nil)
	request.Header.Set("Authorization", "BrowserID "+assertion)
	response := httptest.NewRecorder()

	handleStuff(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", response.Code, response.Body)
	}
}
