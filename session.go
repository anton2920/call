package main

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

func GenerateSessionToken() (string, error) {
	var tokenBytes [64]byte

	timeBytes, err := time.Now().MarshalBinary()
	if err != nil {
		return "", WrapErrorWithTrace(err)
	}
	copy(tokenBytes[:len(timeBytes)], timeBytes)

	if _, err := rand.Read(tokenBytes[len(timeBytes):]); err != nil {
		return "", WrapErrorWithTrace(err)
	}

	tokenEncodedBytes := make([]byte, base64.StdEncoding.EncodedLen(len(tokenBytes)))
	base64.StdEncoding.Encode(tokenEncodedBytes, tokenBytes[:])

	return string(tokenEncodedBytes), nil
}
