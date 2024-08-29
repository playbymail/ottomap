// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package jot provides a simple web token package.
// Do not use this package for anything other than testing.
package jot

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"time"
)

type JOT struct {
	KeyID     string    `json:"kid,omitempty"` // key identifier, implies algorithm
	Subject   int       `json:"sub"`           // this is our user id
	ExpiresAt time.Time `json:"exp"`           // token is valid only before this time
	Signature string    `json:"sig,omitempty"` // signature
}

type Signer_t interface {
	KeyID() string
	Sign(data []byte) ([]byte, error)
}

func CreateToken(subject int, expiresAt time.Time, signer Signer_t) ([]byte, error) {
	jot := JOT{
		Subject:   subject,
		ExpiresAt: expiresAt,
	}
	return jot.Sign(signer)
}

func (j *JOT) Sign(signer Signer_t) ([]byte, error) {
	j.KeyID = signer.KeyID()
	j.Signature = ""

	data, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	signature, err := signer.Sign(data)
	if err != nil {
		return nil, err
	}
	j.Signature = hex.EncodeToString(signature)

	return json.Marshal(j)
}

func (j *JOT) Verify(signer Signer_t) bool {
	if j.KeyID != signer.KeyID() {
		return false
	} else if !time.Now().Before(j.ExpiresAt) {
		return false
	}
	originalSignature := j.Signature
	_, err := j.Sign(signer)
	if err != nil {
		j.Signature = originalSignature
		return false
	}
	newSignature := j.Signature
	j.Signature = originalSignature
	return newSignature == originalSignature
}

func (j *JOT) String() string {
	src, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	dst := make([]byte, base64.RawURLEncoding.EncodedLen(len(src)))
	base64.RawURLEncoding.Encode(dst, src)
	return string(dst)
}

// decode_bytes is a helper for base-64 decoding.
func decode_bytes(src []byte) ([]byte, error) {
	dst := make([]byte, base64.RawURLEncoding.DecodedLen(len(src)))
	_, err := base64.RawURLEncoding.Decode(dst, src)
	if err != nil {
		return nil, err
	}
	return dst, nil
}

// decode_str is a helper for base-64 decoding.
func decode_str(src string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(src)
}

// encode_bytes is a helper for base-64 encoding
func encode_bytes(src []byte) []byte {
	dst := make([]byte, base64.RawURLEncoding.EncodedLen(len(src)))
	base64.RawURLEncoding.Encode(dst, src)
	return dst
}

// encode_str is a helper for base-64 encoding
func encode_str(src string) []byte {
	dst := make([]byte, base64.RawURLEncoding.EncodedLen(len(src)))
	base64.RawURLEncoding.Encode(dst, []byte(src))
	return dst
}
