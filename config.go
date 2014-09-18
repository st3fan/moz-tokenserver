// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package main

const (
	DEFAULT_API_PREFIX         = "/tokenserver"
	DEFAULT_API_LISTEN_ADDRESS = "0.0.0.0"
	DEFAULT_API_LISTEN_PORT    = 8123
	DEFAULT_PERSONA_VERIFIER   = "https://verifier.accounts.firefox.com/v2"
	DEFAULT_PERSONA_AUDIENCE   = "https://tokenserver.sateh.com"
	DEFAULT_ALLOW_NEW_USERS    = true
	DEFAULT_TOKEN_DURATION     = 300
	DEFAULT_SHARED_SECRET      = "cheesebaconeggs"
	DEFAULT_STORAGESERVER_NODE = "http://127.0.0.1:8124/storage"
	DEFAULT_DATABASE_URL       = "postgres://tokenserver:tokenserver@localhost/tokenserver"
)

type TokenServerConfig struct {
	PersonaVerifier   string
	PersonaAudience   string
	AllowNewUsers     bool
	TokenDuration     int64
	SharedSecret      string
	StorageServerNode string
	DatabaseUrl       string
}

func DefaultTokenServerConfig() TokenServerConfig {
	return TokenServerConfig{
		PersonaVerifier:   DEFAULT_PERSONA_VERIFIER,
		PersonaAudience:   DEFAULT_PERSONA_AUDIENCE,
		AllowNewUsers:     DEFAULT_ALLOW_NEW_USERS,
		TokenDuration:     DEFAULT_TOKEN_DURATION,
		SharedSecret:      DEFAULT_SHARED_SECRET,
		StorageServerNode: DEFAULT_STORAGESERVER_NODE,
		DatabaseUrl:       DEFAULT_DATABASE_URL,
	}
}
