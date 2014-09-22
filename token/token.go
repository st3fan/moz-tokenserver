// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/

package token

import (
	"bytes"
	"code.google.com/p/go.crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
)

var TokenSignatureMismatchErr = errors.New("TokenSignatureMismatchErr")
var TokenPayloadDecodingErr = errors.New("TokenPayloadDecodingErr")

const (
	HKDF_INFO_SIGNING = "services.mozilla.com/tokenlib/v1/signing"
	HKDF_INFO_DERIVE  = "services.mozilla.com/tokenlib/v1/derive/"
)

type TokenPayload struct {
	Salt    string `json: "salt"`
	Uid     uint64 `json: "uid"`
	Node    string `json: "node"`
	Expires int64  `json: "expires"`
}

type Token struct {
	Payload       TokenPayload
	Token         string
	DerivedSecret string
}

func randomHexString(length int) (string, error) {
	data := make([]byte, length)
	_, err := rand.Read(data)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

func generateToken(secret []byte, payload TokenPayload) (string, error) {
	secretHkdf := hkdf.New(sha256.New, []byte(secret), nil, []byte(HKDF_INFO_SIGNING))

	signatureSecret := make([]byte, sha256.Size)
	_, err := io.ReadFull(secretHkdf, signatureSecret)
	if err != nil {
		return "", err
	}

	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Calculate and encode the token secret

	mac := hmac.New(sha256.New, signatureSecret)
	mac.Write(encodedPayload)
	payloadSignature := mac.Sum(nil)

	tokenSecret := append(encodedPayload, payloadSignature...)

	return base64.URLEncoding.EncodeToString(tokenSecret), nil
}

func generateDerivedSecret(secret []byte, salt string, encodedTokenSecret string) (string, error) {
	derivedHkdf := hkdf.New(sha256.New, []byte(secret), []byte(salt), []byte(HKDF_INFO_DERIVE+encodedTokenSecret))

	derivedSecret := make([]byte, sha256.Size)
	_, err := io.ReadFull(derivedHkdf, derivedSecret)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(derivedSecret), nil
}

func NewToken(secret []byte, payload TokenPayload) (Token, error) {
	if len(payload.Salt) == 0 {
		var err error
		if payload.Salt, err = randomHexString(3); err != nil {
			return Token{}, err
		}
	}

	token := Token{
		Token:         "",
		DerivedSecret: "",
		Payload:       payload,
	}

	var err error
	if token.Token, err = generateToken(secret, payload); err != nil {
		return Token{}, err
	}

	if token.DerivedSecret, err = generateDerivedSecret(secret, payload.Salt, token.Token); err != nil {
		return Token{}, err
	}

	return token, nil
}

func splitToken(token string) ([]byte, []byte, error) {
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, nil, err
	}

	return decoded[0 : len(decoded)-sha256.Size], decoded[len(decoded)-sha256.Size : len(decoded)], nil
}

func calculateSignatureSecret(secret []byte) ([]byte, error) {
	signatureSecretHkdf := hkdf.New(sha256.New, []byte(secret), nil, []byte(HKDF_INFO_SIGNING))

	signatureSecret := make([]byte, sha256.Size)
	if _, err := io.ReadFull(signatureSecretHkdf, signatureSecret); err != nil {
		return nil, err
	}

	return signatureSecret, nil
}

func calculatePayloadSignature(encodedPayload, signatureSecret []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, signatureSecret)
	mac.Write(encodedPayload)
	payloadSignature := mac.Sum(nil)
	return payloadSignature, nil
}

func ParseToken(secret []byte, tokenSecret string) (Token, error) {
	encodedPayload, signature, err := splitToken(tokenSecret)
	if err != nil {
		return Token{}, err
	}

	signatureSecret, err := calculateSignatureSecret(secret)
	if err != nil {
		return Token{}, err
	}

	// Check the signature on the payload
	expectedSignature, err := calculatePayloadSignature(encodedPayload, signatureSecret)
	if err != nil {
		return Token{}, err
	}
	if !bytes.Equal(signature, expectedSignature) {
		return Token{}, TokenSignatureMismatchErr
	}

	token := Token{
		Token: tokenSecret,
	}

	if err = json.Unmarshal(encodedPayload, &token.Payload); err != nil {
		return Token{}, TokenPayloadDecodingErr
	}

	if token.DerivedSecret, err = generateDerivedSecret(secret, token.Payload.Salt, token.Token); err != nil {
		return Token{}, err
	}

	return token, nil
}
