package main

import (
	"testing"
)

func Test_NewToken(t *testing.T) {
	payload := TokenPayload{
		Uid:     1234,
		Node:    "http://node.example.com",
		Expires: 123467890,
	}

	token, err := NewToken([]byte("thisisasecret"), payload)
	if err != nil {
		t.Error(err)
	}

	if len(token.Token) == 0 {
		t.Error("token.Token is empty")
	}

	if len(token.DerivedSecret) == 0 {
		t.Error("token.DerivedSecret is empty")
	}
}

func Test_ParseToken(t *testing.T) {
	payload := TokenPayload{
		Uid:     1234,
		Node:    "http://node.example.com",
		Expires: 123467890,
	}

	generatedToken, err := NewToken([]byte("thisisasecret"), payload)
	if err != nil {
		t.Error(err)
	}

	if len(generatedToken.Token) == 0 {
		t.Error("generatedToken.Token is empty")
	}

	if len(generatedToken.DerivedSecret) == 0 {
		t.Error("generatedToken.DerivedSecret is empty")
	}

	//

	parsedToken, err := ParseToken([]byte("thisisasecret"), generatedToken.Token)
	if err != nil {
		t.Error(err)
	}

	if generatedToken.Payload.Salt != parsedToken.Payload.Salt {
		t.Error("Different Payload.Salt")
	}
	if generatedToken.Payload.Uid != parsedToken.Payload.Uid {
		t.Error("Different Payload.Uid")
	}
	if generatedToken.Payload.Node != parsedToken.Payload.Node {
		t.Error("Different Payload.Node")
	}
	if generatedToken.Payload.Expires != parsedToken.Payload.Expires {
		t.Error("Different Payload.Expires")
	}

	if generatedToken.Token != parsedToken.Token {
		t.Error("Different Token %+v vs %+v", generatedToken, parsedToken)
	}

	if generatedToken.DerivedSecret != parsedToken.DerivedSecret {
		t.Error("Different DerivedSecret %+v vs %+v", generatedToken, parsedToken)
	}
}
