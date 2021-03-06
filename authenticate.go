package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"log"
	"time"
)

// PuKey is the corresponding public key
var PuKey *rsa.PublicKey

func init() {
	pubStr, err := ioutil.ReadFile("/etc/pki/public.pem")
	if err != nil {
		panic("Couldn't open public key file")
	}
	block, _ := pem.Decode([]byte(pubStr))
	if block == nil {
		panic("Couldn't decode public key from bytearray.")
	}
	puKey, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		log.Print(err.Error())
		panic("Couldn't parse public key.")
	}
	PuKey = puKey.(*rsa.PublicKey)
}

// Verify verifies the signature on the provided data.
func Verify(signed []byte, container interface{}) error {
	// If the nonce has expired or is wrong, just end it
	if NonceIsOlderThan(30 * time.Minute) {
		UpdateNonce()
		return errors.New("nonce was too old")
	}
	// Grab our data
	type signedObj struct {
		Payload []byte
		Nonce   string
		Sig     string
	}
	var data signedObj
	if err := json.Unmarshal(signed, &data); err != nil {
		log.Print(err)
		return errors.New("couldn't unmarshal signed object")
	}

	// Verify the nonce
	nonce, _ := base64.StdEncoding.DecodeString(data.Nonce)
	if !VerifyNonce(nonce) {
		return errors.New("nonce verification failed")
	}

	// Create hash
	//	Nonce first, then data
	h := sha512.New()
	h.Write(nonce)

	// Things are looking okay, let's grab the data
	if data.Payload != nil {
		var payload []byte

		//if _, err := base64.StdEncoding.Decode(payload, data.Payload); err != nil {
		//	log.Print("Couldn't decode payload")
		//	log.Print(data.Payload)
		//	log.Print(err)
		//}

		if string(payload) != string(data.Payload) {
			if err := json.Unmarshal(data.Payload, &container); err != nil {
				log.Print(err)
				return errors.New("couldn't parse payload")
			}
			h.Write(data.Payload)
		}
	}

	// Get that hash
	hash := h.Sum(nil)

	// Verify signature
	sig, _ := base64.StdEncoding.DecodeString(data.Sig)
	err := rsa.VerifyPKCS1v15(PuKey, crypto.SHA512, hash, sig)

	// Return whether it was valid
	return err
}
