package main

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestPostCreate(t *testing.T) {
	// Get Nonce
	res, err := http.Get("http://127.0.0.1:8080/nonce/")
	if err != nil {
		t.Error("Cannot get nonce")
	}
	var nonce Nonce
	var body []byte

	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error("read nonce body")
	}
	err = json.Unmarshal(body, &nonce)
	if err != nil {
		t.Error("Unmarshal nonce")
	}

	// Create dummy Input
	input := Input{
		Title:   "The Inserted Post's;",
		Body:    "This will be the body but I don't want to have to worry about html at the moment",
		IsShort: false,
	}

	// Compute Hash
	h := sha512.New()
	h.Write(nonce.Value)
	h.Write([]byte(input.Title))
	h.Write([]byte(input.Body))
	hash := h.Sum(nil)

	// Sign the hash
	prStr, err := ioutil.ReadFile("/etc/pki/test-private.pem")
	if err != nil {
		t.Error("Couldn't load private key.")
	}
	block, _ := pem.Decode([]byte(prStr))
	prKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Error("Private key")
	}

	sig, err := rsa.SignPKCS1v15(rand.Reader, prKey, crypto.SHA512, hash)
	if err != nil {
		t.Error("sign")
	}

	// Create SignedInput
	signed := SignedInput{
		In:   input,
		Sig:  sig,
		Nnce: nonce,
	}

	// (finally!) make our request!
	j, err := json.Marshal(signed)
	if err != nil {
		t.Error("marshal request")
	}
	req, err := http.NewRequest("POST", "http://localhost:8080/post/", bytes.NewReader(j))
	var c http.Client
	res, err = c.Do(req)

}
