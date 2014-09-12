// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package main

import "testing"

func Test_NewVerifier(t *testing.T) {
	_, err := NewVerifier(TOKENSERVER_PERSONA_VERIFIER, TOKENSERVER_PERSONA_AUDIENCE)
	if err != nil {
		t.Error("Could not create a verifier")
	}
}
