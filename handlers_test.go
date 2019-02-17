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
	res, err := http.Get("http://localhost:8080/nonce/")
	if err != nil {
		t.Error("Cannot get nonce")
	}
	var nonce Nonce
	var body []byte

	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error("Problem reading nonce body")
	}
	err = json.Unmarshal(body, &nonce)
	if err != nil {
		t.Error("Problem unmarshalling nonce")
	}

	// Create dummy Input
	input := Input{
		Title:   "The Inserted Post's Here;",
		Body:    "This will be the body but I don't want to have to worry about html at the moment",
		IsShort: false,
	}

	// Compute Hash
	h := sha512.New()
	inputBytes, _ := json.Marshal(input)
	h.Write(append(nonce.Value, inputBytes...))
	hash := h.Sum(nil)

	// Sign the hash
	prStr, err := ioutil.ReadFile("private.pem") //production
	//prStr, err := ioutil.ReadFile("/home/nico/omfg_lag/pki/private.pem") //dev
	if err != nil {
		t.Error("Couldn't load private key.")
	}
	block, _ := pem.Decode([]byte(prStr))

	prKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Error("Problem decoding private key")
	}

	sig, err := rsa.SignPKCS1v15(rand.Reader, prKey, crypto.SHA512, hash)
	if err != nil {
		t.Error("Problem with creating a signature.")
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
