// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

// TODO: The MockMyID code below should probably move to a mockmyid package in the moz-mockmyid-api project

type MockMyIDResponse struct {
	Assertion string `json:"assertion"`
}

func RequestAssertion(email, audience string) (string, error) {

	u, err := url.Parse("https://mockmyid-api.sateh.com/assertion")
	if err != nil {
		return "", err
	}

	parameters := url.Values{}
	parameters.Add("email", email)
	parameters.Add("audience", audience)
	u.RawQuery = parameters.Encode()

	res, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	response := &MockMyIDResponse{}
	if err = json.Unmarshal(body, response); err != nil {
		return "", err
	}

	return response.Assertion, nil
}

func Test_NewVerifier(t *testing.T) {
	_, err := NewVerifier(TOKENSERVER_PERSONA_VERIFIER, TOKENSERVER_PERSONA_AUDIENCE)
	if err != nil {
		t.Error("Could not create a verifier")
	}
}

func Test_Verify(t *testing.T) {
	// Grab an assertion from the mockmyid api
	assertion, err := RequestAssertion("test@mockmyid.com", TOKENSERVER_PERSONA_AUDIENCE)
	if err != nil {
		t.Error("Could not request assertion", err)
	}
	if len(assertion) == 0 {
		t.Error("Could not create assertion (it is zero length or not returned)")
	}

	// Run it through the verifier
	verifier, err := NewVerifier(TOKENSERVER_PERSONA_VERIFIER, TOKENSERVER_PERSONA_AUDIENCE)
	if err != nil {
		t.Error("Could not create a verifier")
	}
	response, err := verifier.VerifyAssertion(assertion)
	if err != nil {
		t.Error("Could not verify assertion")
	}
	if response.Status != "okay" {
		t.Errorf("Failed to verify assertion: %s / %s", response.Status, response.Reason)
	}
}
