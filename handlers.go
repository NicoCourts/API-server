package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Index just welcomes you
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the NicoCourts.com blog API!")
	fmt.Fprintln(w, "Visit https://api.nicocourts.com/posts for the post list.")
}

// PostIndex returns a JSON list of all posts
func PostIndex(w http.ResponseWriter, r *http.Request) {

	// Responsibly declare our content type and return code
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// TODO Replace this for production
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "https://nicocourts.com")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(RepoGetVisiblePosts()); err != nil {
		panic(err)
	}
}

// AllPostIndex returns a JSON list of all posts (including invisible ones)
func AllPostIndex(w http.ResponseWriter, r *http.Request) {

	// Responsibly declare our content type and return code
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// TODO Replace this for production
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "https://nicocourts.com")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(RepoGetAllPosts()); err != nil {
		panic(err)
	}
}

// PostShow returns the details of a specific post
func PostShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["postID"]

	// Responsibly declare our content type and return code
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// TODO Replace this for production
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "https://nicocourts.com")

	p := RepoGetPost(postID)
	if (p != Post{}) {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(p); err != nil {
			panic("Error with JSON encoding")
		}
	}
	// Found nothing
	w.WriteHeader(http.StatusNoContent)
}

// PostCreate inserts a new post into the repo. Requests will be JSON of
//	the form {"title":"t", "body":"b", "isshort":T/F} where t and b are
//	treated as html that has been escaped via html.EscapeString().
func PostCreate(w http.ResponseWriter, r *http.Request) {
	var input SignedInput

	// Don't allow people to flood our API with data
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1000000))

	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// TODO Replace this for production
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "https://nicocourts.com")

	if err := json.Unmarshal(body, &input); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)

		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	// Verify nonce and signature
	var in []byte
	in = append(in, []byte(input.In.Title)...)
	in = append(in, []byte(input.In.Body)...)
	if !Verify(in, input.Nnce, input.Sig) {
		w.WriteHeader(http.StatusUnauthorized)
		log.Print("Unauthorized Access Attempt")
		return
	}

	// Make the URLTitle
	urlTitle := strings.Replace(strings.ToLower(input.In.Title), " ", "-", -1)
	re := regexp.MustCompile("[^a-zA-Z0-9-]+")
	urlTitle = re.ReplaceAllString(urlTitle, "")
	if len(urlTitle) > 35 {
		urlTitle = urlTitle[:35]
	}

	// We've confirmed authenticity at this point. Prepare post for insertion.
	post := Post{
		Title:    input.In.Title,
		URLTitle: urlTitle,
		Body:     input.In.Body,
		IsShort:  input.In.IsShort,
		Visible:  true,
		Date:     time.Now(),
	}

	p := RepoCreatePost(post, r)
	if err := json.NewEncoder(w).Encode(p); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	w.WriteHeader(http.StatusCreated)
}

// PostDelete deletes the post with the given ID
func PostDelete(w http.ResponseWriter, r *http.Request) {
	var input SignedDeleteRequest

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
		return
	}

	// Ensure the request is valid
	vars := mux.Vars(r)
	postID := vars["postID"]
	id, _ := strconv.Atoi(postID)
	if id != input.ID {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(fmt.Sprintf("Bad Delete Request: Expected ID %v, got ID %v.", postID, input.ID))
		return
	}

	// Verify nonce and signature
	var in []byte
	in = []byte(postID)

	// TODO Replace this for production
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "https://nicocourts.com")

	if !Verify(in, input.Nnce, input.Sig) {
		w.WriteHeader(http.StatusUnauthorized)
		log.Print("Unauthorized Access Attempt")
		return
	}

	if err := RepoDestroyPost(postID); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}
	w.WriteHeader(http.StatusAccepted)

}

// ReadNonce prints out the current nonce to use for authentication
func ReadNonce(w http.ResponseWriter, r *http.Request) {
	// Responsibly declare our content type and return code
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// TODO Replace this for production
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "https://nicocourts.com")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(CurrentNonce()); err != nil {
		panic(err)
	}
}

// NonceUpdate just churns the nonce. This is useful if the client
//	notices the nonce is near expiring and would rather not risk it.
//	Since generating pseudorandom noise can be expensive, require
//	nonce to be at least 10 minutes old to prevent DDOS attacks.
func NonceUpdate(w http.ResponseWriter, r *http.Request) {
	if NonceIsOlderThan(10 * time.Minute) {
		UpdateNonce()
		w.WriteHeader(http.StatusAccepted)
	} else {
		// Sorry, chum
		w.WriteHeader(http.StatusForbidden)
	}
}
