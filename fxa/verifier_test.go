// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package fxa

import (
	"encoding/json"
	"github.com/st3fan/moz-tokenserver/fxa"
	"github.com/st3fan/moz-tokenserver/mockmyid"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

// TODO: The MockMyID code below should probably move to a mockmyid package in the moz-mockmyid-api project

func Test_NewVerifier(t *testing.T) {
	_, err := NewVerifier(DEFAULT_PERSONA_VERIFIER, DEFAULT_PERSONA_AUDIENCE)
	if err != nil {
		t.Error("Could not create a verifier")
	}
}

func Test_Verify(t *testing.T) {
	// Grab an assertion from the mockmyid api
	assertion, err := mockmyid.RequestAssertion("test@mockmyid.com", DEFAULT_PERSONA_AUDIENCE)
	if err != nil {
		t.Error("Could not request assertion", err)
	}
	if len(assertion) == 0 {
		t.Error("Could not create assertion (it is zero length or not returned)")
	}

	// Run it through the verifier
	verifier, err := fxa.NewVerifier(DEFAULT_PERSONA_VERIFIER, DEFAULT_PERSONA_AUDIENCE)
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
