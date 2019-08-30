package main

// Input is the information we expect from the client to create a new post.
type Input struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Markdown string `json:"markdown"`
}

// SignedInput is an Input/Signature/Nonce triple.
type SignedInput struct {
	In   Input
	Sig  []byte
	Nnce Nonce
}

// SignedPostDeleteRequest is a post ID/sig/Nonce triple.
type SignedPostDeleteRequest struct {
	ID   int
	Sig  []byte
	Nnce Nonce
}

// SignedImageDeleteRequest takes a filename and signature and nonce.
type SignedImageDeleteRequest struct {
	Filename string
	Sig      []byte
	Nnce     Nonce
}
