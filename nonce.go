package main

import (
	"bytes"
	"crypto/rand"
	"log"
	"time"
)

// Nonce is very aptly named. The Created value allows the Value
//	to be regularly cycled to avoid any precomputation.
type Nonce struct {
	Value   []byte    `json:"value"`
	Created time.Time `json:"created"`
}

var theNonce Nonce

func init() {
	UpdateNonce()
}

// NonceIsOlderThan returns a bool that reflects whether the nonce is older
//	than the provided time in nanoseconds
func NonceIsOlderThan(d time.Duration) bool {
	return time.Since(theNonce.Created) > d
}

// CurrentNonce checks first to make sure the nonce isn't too old. If it is,
//	it updates the value BEFORE returning the new one. The entire Nonce
//	object is returned to give the client a heads up if it is about to expire.
func CurrentNonce() Nonce {
	// Don't use nonces that are more than half an hour old
	if NonceIsOlderThan(30 * time.Minute) {
		UpdateNonce()
	}

	// Return that bad boy
	return theNonce
}

// VerifyNonce just checks whether the given nonce is the
//	current nonce and returns a bool to that effect. Regardless
//	of result the value of the current nonce is changed.
func VerifyNonce(in []byte) bool {
	// Update the nonce after checking for equality
	defer UpdateNonce()
	return bytes.Equal(in, theNonce.Value)
}

// UpdateNonce creates a new value and updates the field
func UpdateNonce() {
	// we will use a 64-byte (512-bit) pseudorandom nonce
	val := make([]byte, 64)
	_, err := rand.Read(val)
	if err != nil {
		log.Fatal(err)
	}

	// update the nonce with our new value
	theNonce = Nonce{
		Value:   val,
		Created: time.Now(),
	}

}
