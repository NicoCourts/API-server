package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/posts/", PostIndex)
	router.HandleFunc("/posts/{postId}", PostShow)

	log.Fatal(http.ListenAndServe(":8080", router))
}

// Index just welcomes you
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

// PostIndex returns a JSON list of all posts
func PostIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "This will eventually return a list of posts.")
}

// PostShow returns the details of a specific post
func PostShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["postId"]
	fmt.Fprintln(w, "This would show post:", postID)
}
