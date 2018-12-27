package main

// Input is the information we expect from the client to create a new post.
type Input struct {
	Title   string `json:"title"`
	Body    string `json:"body"`
	IsShort bool   `json:"isshort"`
}

// SignedInput is an Input/Signature/Nonce triple.
type SignedInput struct {
	In   Input
	Sig  []byte
	Nnce Nonce
}

// SignedDeleteRequest is a post ID/sig/Nonce triple.
type SignedDeleteRequest struct {
	ID   int
	Sig  []byte
	Nnce Nonce
}
