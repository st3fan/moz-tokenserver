// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package main

import (
	"code.google.com/p/go.crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
)

const (
	HKDF_INFO_SIGNING = "services.mozilla.com/tokenlib/v1/signing"
	HKDF_INFO_DERIVE  = "services.mozilla.com/tokenlib/v1/derive/"
)

type TokenPayload struct {
	Salt    string `json: "salt"`
	Uid     string `json: "uid"`
	Node    string `json: "node"`
	Expires int64  `json: "expires"`
}

func randomHexString(length int) (string, error) {
	data := make([]byte, length)
	_, err := rand.Read(data)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

func GenerateSecret(uid string, node string, expires int64, secret string) (string, string, error) {
	secretHkdf := hkdf.New(sha256.New, []byte(secret), nil, []byte(HKDF_INFO_SIGNING))

	signatureSecret := make([]byte, sha256.Size)
	_, err := io.ReadFull(secretHkdf, signatureSecret)
	if err != nil {
		return "", "", err
	}

	salt, err := randomHexString(3)
	if err != nil {
		return "", "", err
	}

	tokenPayload := &TokenPayload{
		Salt:    salt,
		Uid:     uid,
		Node:    node,
		Expires: expires,
	}

	encodedPayload, err := json.Marshal(tokenPayload)
	if err != nil {
		return "", "", err
	}

	// Calculate and encode the token secret

	mac := hmac.New(sha256.New, signatureSecret)
	mac.Write(encodedPayload)
	payloadSignature := mac.Sum(nil)

	tokenSecret := append(encodedPayload, payloadSignature...)

	encodedTokenSecret := base64.URLEncoding.EncodeToString(tokenSecret)

	// Calculate and encode the derived secret

	derivedHkdf := hkdf.New(sha256.New, []byte(secret), []byte(salt),
		[]byte(HKDF_INFO_DERIVE+encodedTokenSecret))

	derivedSecret := make([]byte, sha256.Size)
	_, err = io.ReadFull(derivedHkdf, derivedSecret)
	if err != nil {
		return "", "", err
	}

	encodedDerivedSecret := base64.URLEncoding.EncodeToString(derivedSecret)

	return encodedTokenSecret, encodedDerivedSecret, nil
}
