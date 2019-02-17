package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var origin string

func init() {
	// TODO Replace this for production
	origin = "*"
	// origin = "https://nicocourts.com"
}

// Index just welcomes you
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the NicoCourts.com blog API!")
	fmt.Fprintln(w, "Visit https://api.nicocourts.com/posts for the post list.")
}

// PostIndex returns a JSON list of all posts
func PostIndex(w http.ResponseWriter, r *http.Request) {

	// Responsibly declare our content type and return code
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(RepoGetVisiblePosts()); err != nil {
		panic(err)
	}
}

// AllPostIndex returns a JSON list of all posts (including invisible ones)
func AllPostIndex(w http.ResponseWriter, r *http.Request) {

	// Responsibly declare our content type and return code
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)
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
	w.Header().Set("Access-Control-Allow-Origin", origin)

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
	// Don't allow people to flood our API with data
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1000000))

	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	// Responsibly declare our content type and return code
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)

	// Refactoring to maximize code reuse
	var input Input
	if err := Verify(body, &input); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Print("Unauthorized Access Attempt")
		return
	}

	// Make the URLTitle
	urlTitle := strings.Replace(strings.ToLower(input.Title), " ", "-", -1)
	re := regexp.MustCompile("[^a-zA-Z0-9-]+")
	urlTitle = re.ReplaceAllString(urlTitle, "")
	if len(urlTitle) > 35 {
		urlTitle = urlTitle[:35]
	}

	// We've confirmed authenticity at this point. Prepare post for insertion.
	post := Post{
		Title:    input.Title,
		URLTitle: urlTitle,
		Body:     input.Body,
		IsShort:  input.IsShort,
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
	// Don't allow people to flood our API with data
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1000000))

	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	// Responsibly declare our content type and return code
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)

	type DelPost struct {
		ToRemove string
	}
	var in DelPost
	if err := Verify(body, &in); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Print("Unauthorized Access Attempt")
		return
	}

	if err := RepoDestroyPost(in.ToRemove); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}
	w.WriteHeader(http.StatusAccepted)
}

// ImageDelete deletes the post with the given ID
func ImageDelete(w http.ResponseWriter, r *http.Request) {
	// Don't allow people to flood our API with data
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1000000))

	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	// Responsibly declare our content type and return code
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)

	type PicData struct {
		Filename string
	}
	var data PicData
	if err := Verify(body, &data); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Print("Unauthorized Access Attempt")
		return
	}

	// Everything is kosher -- delete the file.
	//f err := os.Remove("/etc/img/" + filename); err != nil {
	if err := os.Remove("/home/nico/" + data.Filename); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// UploadImage takes in some multipart form info representing an image
//	and returns metadata for the resource if the upload is successful.
func UploadImage(w http.ResponseWriter, r *http.Request) {
	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		log.Print("Can't read media headers")
		log.Print(err)
		return
	}
	// Don't allow people to flood our API with data
	rdr := multipart.NewReader(r.Body, params["boundary"])
	form, err := rdr.ReadForm(50000000)
	if err != nil {
		panic(err)
	}

	var img multipart.File
	var filename string
	for k, v := range form.File {
		img, _ = v[0].Open()
		filename = k
	}

	Sig := form.Value["Sig"][0]
	Nonce := form.Value["Nonce"][0]

	// Responsibly declare our content type and return code
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	type Nothing struct{}
	type SignNothing struct {
		Payload interface{}
		Sig     string
		Nonce   string
	}
	var nada Nothing
	blob, _ := json.Marshal(SignNothing{nil, Sig, Nonce})

	if err := Verify(blob, &nada); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Print("Unauthorized Access Attempt")
		log.Print(err)
		return
	}

	// Get new filename
	h := md5.New()
	var imgBytes []byte
	img.Read(imgBytes)
	h.Sum(imgBytes)
	checksum := h.Sum(nil)

	name := hex.EncodeToString(checksum) + filepath.Ext(filename)

	// Write the file to disk
	//f, err := os.OpenFile("/etc/img/"+name, os.O_WRONLY|os.O_CREATE, 0666)
	f, err := os.OpenFile("/home/nico/"+name, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	defer f.Close()
	f.Write(imgBytes)

	image := RepoAddImage(hex.EncodeToString(checksum), filepath.Ext(filename), (filename)[0:len(filename)-len(filepath.Ext(filename))])
	err = json.NewEncoder(w).Encode(image)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Great!
	w.WriteHeader(http.StatusCreated)
}

// GetImageList returns a list of currently-available images along with some metadata.
func GetImageList(w http.ResponseWriter, r *http.Request) {
	// Responsibly declare our content type and return code
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(RepoGetImageList()); err != nil {
		panic(err)
	}
}

// ReadNonce prints out the current nonce to use for authentication
func ReadNonce(w http.ResponseWriter, r *http.Request) {
	// Responsibly declare our content type and return code
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// TODO Replace this for production
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "https://nicocourts.com")
	w.WriteHeader(http.StatusOK)

	nonce := CurrentNonce()

	if err := json.NewEncoder(w).Encode(nonce); err != nil {
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
