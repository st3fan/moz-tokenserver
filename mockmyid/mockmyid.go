// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package mockmyid

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

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
