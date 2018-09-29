package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Index just welcomes you
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the NicoCourts.com blog API!")
}

// PostIndex returns a JSON list of all posts
func PostIndex(w http.ResponseWriter, r *http.Request) {

	// Responsibly declare our content type and return code
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(RepoGetVisiblePosts()); err != nil {
		panic(err)
	}
}

// PostShow returns the details of a specific post
func PostShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["postID"]

	// Responsibly declare our content type and return code
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(RepoGetPost(postID)); err != nil {
		panic(err)
	}
}

// PostCreate inserts a new post into the repo. Requests will be JSON of
//	the form {"title":"t", "body":"b", "isshort":T/F} where t and b are
//	treated as html that has been escaped via html.EscapeString().
func PostCreate(w http.ResponseWriter, r *http.Request) {
	type Input struct {
		Title   string `json:"title"`
		Body    string `json:"body"`
		IsShort bool   `json:"isshort"`
	}

	var input Input

	// Don't allow people to flood our API with data
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1000000))

	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if err := json.Unmarshal(body, &input); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	post := Post{
		Title:   input.Title,
		Body:    input.Body,
		IsShort: input.IsShort,
		Visible: true,
		Date:    time.Now(),
	}

	p := RepoCreatePost(post)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(p); err != nil {
		panic(err)
	}
}

// PostDelete deletes the post with the given ID
func PostDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["postID"]

	if err := RepoDestroyPost(postID); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}
	w.WriteHeader(http.StatusAccepted)

}
