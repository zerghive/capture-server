package ca

import (
	"bytes"
	"crypto/x509"
	"encoding/gob"
	"errors"
	"time"
)

type AuthToken struct {
	OrgId    string
	FromDate time.Time
	ToDate   time.Time
}

type signedObj struct {
	Data      []byte
	Signature []byte
}

func getVerifiedAuthToken(data []byte) (token *AuthToken, err error) {

	token = nil
	signedObj := new(signedObj)
	dec := gob.NewDecoder(bytes.NewReader(data))
	if err = dec.Decode(signedObj); err != nil {
		return
	}

	if err = RootCA.Leaf.CheckSignature(x509.SHA256WithRSA, signedObj.Data, signedObj.Signature); err != nil {
		return
	}

	token = new(AuthToken)
	dec = gob.NewDecoder(bytes.NewReader(signedObj.Data))
	if err = dec.Decode(token); err != nil {
		return
	}

	return
}

func Auth(data []byte) (*AuthToken, error) {

	token, err := getVerifiedAuthToken(data)
	if err != nil {
		return nil, err
	}

	if token.OrgId == "" {
		return nil, errors.New("Empty organization ID.")
	}

	now := time.Now()
	if now.Before(token.FromDate) || now.After(token.ToDate) {
		return nil, errors.New("Invalid token time interval.")
	}

	return token, nil
}
