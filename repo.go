/*repo.go provides an interface for a data repo.
 *		At the moment this will hold some dummy data while the rest of
 *		the server is developed.
 */
package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/OneOfOne/xxhash"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var currentID int

func init() {
	//Code for providing test data

	/*r := Rsvp{
		ID:         12345,
		ShortCode:  "ABCD",
		Attending:  true,
		NumInvited: 4,
		MonConfirm: 2,
		SunConfirm: 3,
	}
	req, _ := http.NewRequest("POST", "/rsvp/", nil)
	RepoCreateRSVP(r, req)*/
}

// RepoCreatePost adds a new post to our data store.
func RepoCreatePost(post Post, r *http.Request) Post {
	// Create channel and mutex
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex

	// Prepare mutex to hold connection open until we're done with it.
	mux.Lock()
	defer mux.Unlock()

	// Open the connection and catch the incoming pointer
	go databaseHelper(ch1, &mux)
	c := <-ch1

	// Get the id to use
	id := getNextID(post.URLTitle, post.Date)
	post.ID = id

	// Insert post
	err := c.Insert(post)
	if err != nil {
		log.Fatal(err)
	}

	return post
}

// getNextID returns the ID for a post added at this moment.
func getNextID(title string, date time.Time) uint32 {
	// hash the (url-safe) title and the time together using a fast hash
	h := xxhash.New32()
	h.Write([]byte(title))
	h.Write([]byte(date.String()))

	return h.Sum32()
}

// RepoGetPost returns the post for the given ID (if one exists). If
//	not, return a blank post.
func RepoGetPost(urltitle string) Post {
	// Create channel and mutex
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex

	// Prepare mutex to hold connection open until we're done with it.
	mux.Lock()
	defer mux.Unlock()

	// Open the connection and catch the incoming pointer
	go databaseHelper(ch1, &mux)
	c := <-ch1

	var post Post
	err := c.Find(bson.M{"urltitle": urltitle}).One(&post)
	if err != nil {
		log.Print("Post not found!")
		return Post{}
	}

	return post
}

// RepoDestroyPost deletes (disables, actually) a post.
func RepoDestroyPost(postID string) error {
	// Create channel and mutex
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex

	// Prepare mutex to hold connection open until we're done with it.
	mux.Lock()
	defer mux.Unlock()

	// Open the connection and catch the incoming pointer
	go databaseHelper(ch1, &mux)
	c := <-ch1

	// Find post, if it exists
	var post Post
	id, _ := strconv.Atoi(postID)
	err := c.Find(bson.M{"id": id}).One(&post)
	var e Post

	if post == e {
		return fmt.Errorf("Could not find Post with ID of %s to delete", postID)
	}

	// Toggle visibility
	err = c.Update(post, bson.M{"$set": bson.M{"visible": false}})
	if err != nil {
		return fmt.Errorf("Could not update post")
	}

	return nil

}

// RepoGetVisiblePosts returns a list of all visible posts (publc)
func RepoGetVisiblePosts() Posts {
	// Create channel and mutex
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex

	// Prepare mutex to hold connection open until we're done with it.
	mux.Lock()
	defer mux.Unlock()

	// Open the connection and catch the incoming pointer
	go databaseHelper(ch1, &mux)
	c := <-ch1

	var posts Posts
	err := c.Find(bson.M{"visible": true}).All(&posts)
	if err != nil {
		log.Fatal(err)
	}

	return posts
}

// RepoAddImage adds a new image to the database
func RepoAddImage(filename string, extension string, shortname string) Image {
	// Create channel and mutex
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex

	// Prepare mutex to hold connection open until we're done with it.
	mux.Lock()
	defer mux.Unlock()

	// Open the connection and catch the incoming pointer
	go databaseHelper(ch1, &mux, "images")
	c := <-ch1

	// Create the Image
	img := Image{
		Filename: filename + extension,
		Title:    shortname,
		AltText:  shortname,
		URL:      "https://nicocourts.com/img/" + filename + extension,
		Date:     time.Now(),
	}

	// Insert post
	err := c.Insert(img)
	if err != nil {
		log.Fatal(err)
	}

	return img
}

// RepoGetImageList returns a list of all available images with urls and friendly names
func RepoGetImageList() Images {
	// Create channel and mutex
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex

	// Prepare mutex to hold connection open until we're done with it.
	mux.Lock()
	defer mux.Unlock()

	// Open the connection and catch the incoming pointer
	go databaseHelper(ch1, &mux, "images")
	c := <-ch1

	var images Images
	if err := c.Find(bson.M{}).All(&images); err != nil {
		log.Fatal(err)
	}

	return images
}

// RepoGetAllPosts returns a list of all visible posts (publc)
func RepoGetAllPosts() Posts {
	// Create channel and mutex
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex

	// Prepare mutex to hold connection open until we're done with it.
	mux.Lock()
	defer mux.Unlock()

	// Open the connection and catch the incoming pointer
	go databaseHelper(ch1, &mux)
	c := <-ch1

	var posts Posts
	err := c.Find(bson.M{}).All(&posts)
	if err != nil {
		log.Fatal(err)
	}

	return posts
}

// databaseHelper does the work of opening the database
func databaseHelper(c1 chan *mgo.Collection, mux *sync.Mutex, table ...string) {
	//Parse optional argument
	if len(table) == 0 {
		table = append(table, "posts")
	}

	//Set up DB connection
	session, err := mgo.Dial("mongodb:27017") //production
	//session, err := mgo.Dial("localhost:27017") //dev
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Choose database and collection
	conn := session.DB("postDB").C(table[0])

	// Send back connection.
	c1 <- conn
	// Wait for the mutex to unlock before quitting
	mux.Lock()
}

/////////////////////////////////////////////////////////////

// databaseHelperRSVP does the work of opening the database
func databaseHelperRSVP(c1 chan *mgo.Collection, mux *sync.Mutex, table ...string) {
	//Parse optional argument
	if len(table) == 0 {
		table = append(table, "posts")
	}

	//Set up DB connection
	session, err := mgo.Dial("mongodb:27017") //production
	//session, err := mgo.Dial("localhost:27017") //dev
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Choose database and collection
	conn := session.DB("rsvpDB").C(table[0])

	// Send back connection.
	c1 <- conn
	// Wait for the mutex to unlock before quitting
	mux.Lock()
}

// RepoGetRSVP returns the post for the given ID (if one exists). If
//	not, return a blank post.
func RepoGetRSVP(rescode string) Rsvp {
	// Create channel and mutex
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex

	// Prepare mutex to hold connection open until we're done with it.
	mux.Lock()
	defer mux.Unlock()

	// Open the connection and catch the incoming pointer
	go databaseHelperRSVP(ch1, &mux)
	c := <-ch1

	var rsvp Rsvp
	err := c.Find(bson.M{"shortcode": rescode}).One(&rsvp)
	if err != nil {
		log.Print("Post not found!")
		return Rsvp{}
	}

	return rsvp
}

// RepoCreateRSVP adds a new RSVP to our data store.
func RepoCreateRSVP(rsvp Rsvp, r *http.Request) Rsvp {
	// Create channel and mutex
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex

	// Prepare mutex to hold connection open until we're done with it.
	mux.Lock()
	defer mux.Unlock()

	// Open the connection and catch the incoming pointer
	go databaseHelperRSVP(ch1, &mux)
	c := <-ch1

	// Insert rsvp
	err := c.Insert(rsvp)
	if err != nil {
		log.Fatal(err)
	}

	return rsvp
}
