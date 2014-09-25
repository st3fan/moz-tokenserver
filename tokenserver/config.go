// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package tokenserver

const (
	DEFAULT_PERSONA_VERIFIER   = "https://verifier.accounts.firefox.com/v2"
	DEFAULT_PERSONA_AUDIENCE   = "http://127.0.0.1"
	DEFAULT_ALLOW_NEW_USERS    = true
	DEFAULT_TOKEN_DURATION     = 300
	DEFAULT_SHARED_SECRET      = "ThisIsAnImportantSecretThatYouShouldChange"
	DEFAULT_STORAGESERVER_NODE = "http://127.0.0.1:8124/storage"
	DEFAULT_DATABASE_PATH      = "/tmp/tokenserver.db"
)

type Config struct {
	PersonaVerifier   string
	PersonaAudience   string
	AllowNewUsers     bool
	TokenDuration     int64
	SharedSecret      string
	StorageServerNode string
	DatabasePath      string
}

func DefaultConfig() Config {
	return Config{
		PersonaVerifier:   DEFAULT_PERSONA_VERIFIER,
		PersonaAudience:   DEFAULT_PERSONA_AUDIENCE,
		AllowNewUsers:     DEFAULT_ALLOW_NEW_USERS,
		TokenDuration:     DEFAULT_TOKEN_DURATION,
		SharedSecret:      DEFAULT_SHARED_SECRET,
		StorageServerNode: DEFAULT_STORAGESERVER_NODE,
		DatabasePath:      DEFAULT_DATABASE_PATH,
	}
}
