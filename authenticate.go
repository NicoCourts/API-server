package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"time"
)

// PuKey is the corresponding public key
var PuKey *rsa.PublicKey

func init() {
	pubStr, err := ioutil.ReadFile("/etc/pki/public.pem") //production
	//pubStr, err := ioutil.ReadFile("/home/nico/omfg_lag/pki/public.pem") //dev
	if err != nil {
		panic("Couldn't open public key file")
	}
	block, _ := pem.Decode([]byte(pubStr))
	if block == nil {
		panic("Couldn't decode public key from bytearray.")
	}
	puKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	PuKey = puKey.(*rsa.PublicKey)

	if err != nil {
		panic("Couldn't parse public key.")
	}
}

// Verify verifies the signature on the provided data.
func Verify(in []byte, nonce []byte, sig []byte) bool {
	// If the nonce has expired or is wrong, just end it
	if NonceIsOlderThan(30 * time.Minute) {
		UpdateNonce()
		return false
	}
	if !VerifyNonce(nonce) {
		return false
	}

	// Create hash
	//	Nonce first, then data
	h := sha512.New()
	h.Write(nonce)
	h.Write(in)

	// Get that hash
	hash := h.Sum(nil)

	// Verify signature
	err := rsa.VerifyPKCS1v15(PuKey, crypto.SHA512, hash, sig)

	// Return whether it was valid
	return err == nil
}
