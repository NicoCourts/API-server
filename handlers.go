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
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var origin string

func init() {
	origin = "*"
}

// Index just welcomes you
func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the NicoCourts.com API!")
	fmt.Fprintln(w, "Visit <a href='https://api.nicocourts.com/posts'>this link</a> for the post list.")
}

// PostIndex returns a JSON list of all posts
func PostIndex(w http.ResponseWriter, r *http.Request) {

	// Responsibly declare our content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(RepoGetVisiblePosts()); err != nil {
		panic(err)
	}
}

// AllPostIndex returns a JSON list of all posts (including invisible ones)
func AllPostIndex(w http.ResponseWriter, r *http.Request) {
	// Don't allow people to flood our API with data
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1000000))

	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	type Nothing struct{}
	var nada Nothing

	if err := Verify(body, &nada); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Print("Unauthorized Access Attempt")
		log.Print(err)
		return
	}

	// Responsibly declare our content type
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

	p := RepoGetPost(postID)
	if (p != Post{}) {
		// Responsibly declare our content type
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if err := json.NewEncoder(w).Encode(p); err != nil {
			panic("Error with JSON encoding")
		}

		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
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

	p := RepoCreatePost(post)
	if err := json.NewEncoder(w).Encode(p); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	w.WriteHeader(http.StatusCreated)
	// Responsibly declare our content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)
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
	// Responsibly declare our content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)
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
	// Responsibly declare our content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)
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

	img := form.Value["img"][0]
	filename := form.Value["Filename"][0]
	Sig := form.Value["Sig"][0]
	Nonce := form.Value["Nonce"][0]

	type Nothing struct{}
	type SignNothing struct {
		Payload interface{}
		Sig     string
		Nonce   string
	}
	var nada Nothing
	stuff := SignNothing{nil, Sig, Nonce}
	blob, _ := json.Marshal(stuff)

	if err := Verify(blob, &nada); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Print("Unauthorized Access Attempt")
		log.Print(err)
		return
	}

	// Get new filename
	h := md5.New()
	imgBytes := []byte(img)
	h.Sum(imgBytes)
	checksum := h.Sum(nil)

	name := hex.EncodeToString(checksum) + filepath.Ext(filename)
	log.Print("Name: " + name)

	// Write the file to disk
	f, err := os.OpenFile("/etc/img/"+name, os.O_WRONLY|os.O_CREATE, 0666)
	//f, err := os.OpenFile("/home/nico/"+name, os.O_WRONLY|os.O_CREATE, 0666)
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
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)
}

// GetImageList returns a list of currently-available images along with some metadata.
func GetImageList(w http.ResponseWriter, r *http.Request) {
	// Responsibly declare our content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(RepoGetImageList()); err != nil {
		panic(err)
	}
}

// ReadNonce prints out the current nonce to use for authentication
func ReadNonce(w http.ResponseWriter, r *http.Request) {
	// Responsibly declare our content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)

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

// Handlers that will do the work for our wedding

// GetRSVP looks up an RSVP given a reservation code and returns
//	the current information we have on it.
func GetRSVP(w http.ResponseWriter, r *http.Request) {
	// Responsibly declare our content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)

	vars := mux.Vars(r)
	rescode := vars["rescode"]

	rsvp := RepoGetRSVP(rescode)
	if (rsvp != Rsvp{}) {
		if err := json.NewEncoder(w).Encode(rsvp); err != nil {
			panic("Error with JSON encoding")
		}

		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

// UpdateRSVP updates the current RSVP with new information.
func UpdateRSVP(w http.ResponseWriter, r *http.Request) {
	// Responsibly declare our content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)

	vars := mux.Vars(r)
	rescode := vars["rescode"]

	// Get data from the database
	currRSVP := RepoGetRSVP(rescode)
	if (currRSVP == Rsvp{}) {
		w.WriteHeader(http.StatusBadRequest)
		log.Print("RSVP not found!")
		return
	}

	// Get POST variables
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err)
		return
	}

	attending := r.FormValue("attending")
	inv := currRSVP.NumInvited
	monconfirm := r.FormValue("monconfirm")
	sunconfirm := r.FormValue("sunconfirm")

	// Parse values and make sure it's a valid request.
	// Don't allow people to reserve more than their allotted spots
	mon, err := strconv.Atoi(monconfirm)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err)
		return
	}
	sun, err := strconv.Atoi(sunconfirm)
	if err != nil || inv < mon || inv < sun {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err)
		return
	}

	err = RepoUpdateRSVP(rescode, attending, mon, sun)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		log.Print(err)
	} else {
		w.WriteHeader(http.StatusNoContent)
		log.Print("Oh no.")
	}
}

// CreateRSVP creates one!
func CreateRSVP(w http.ResponseWriter, r *http.Request) {
	// Get POST variables
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print(err)
		return
	}
	name := r.FormValue("name")
	numinvited := r.FormValue("numinvited")
	rescode := r.FormValue("rescode")

	inv, err := strconv.Atoi(numinvited)
	if err == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rsvp := RepoCreateRSVP(rescode, name, inv)
	if (rsvp != Rsvp{}) {
		// Responsibly declare our content type
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if err := json.NewEncoder(w).Encode(rsvp); err != nil {
			panic("Error with JSON encoding")
		}

		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

// ListRSVP does stuff
func ListRSVP(w http.ResponseWriter, r *http.Request) {
	// Responsibly declare our content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(RepoGetRSVPs()); err != nil {
		panic(err)
	}
}
