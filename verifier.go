// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Verifier struct {
	verifier string
	audience string
}

type VerifierRequest struct {
	Assertion string `json:"assertion"`
	Audience  string `json:"audience"`
}

type VerifierResponse struct {
	Status   string `json:"status"`
	Email    string `json:"email"`
	Audience string `json:"audience"`
	Expires  int64  `json:"expires"`
	Issuer   string `json:"issuer"`
	Reason   string `json:"reason,omitempty"`
}

func NewVerifier(verifier, audience string) (*Verifier, error) {
	return &Verifier{verifier: verifier, audience: audience}, nil
}

func (v *Verifier) VerifyAssertion(assertion string) (*VerifierResponse, error) {
	verifierRequest := VerifierRequest{
		Audience:  v.audience,
		Assertion: assertion,
	}

	encodedVerifierRequest, err := json.Marshal(verifierRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", v.verifier, bytes.NewBuffer(encodedVerifierRequest))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	verifierResponse := &VerifierResponse{}
	if err = json.Unmarshal(body, verifierResponse); err != nil {
		return nil, err
	}

	return verifierResponse, nil
}
