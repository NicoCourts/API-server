/*repo.go provides an interface for a data repo.
 *		At the moment this will hold some dummy data while the rest of
 *		the server is developed.
 */
package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/OneOfOne/xxhash"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var currentID int

func init() {
	/* p := Post{
		ID:       1234,
		Title:    "There",
		URLTitle: "there",
		Body:     "This is the body",
		Date:     time.Now(),
		Visible:  true,
		IsShort:  true,
	}

	req, _ := http.NewRequest("GET", "/post/", nil)
	RepoCreatePost(p, req) */
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
func RepoGetPost(postID string) Post {
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
	err := c.Find(bson.M{"_id": []byte(postID)}).One(&post)
	if err != nil {
		log.Fatal("Problem with reading post!")
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
	err := c.Find(bson.M{"_id": postID}).One(&post)
	var e Post

	if post == e {
		return fmt.Errorf("Could not find Post with ID of %s to delete", postID)

	}

	// Toggle visibility
	err = c.Update(post, bson.M{"visible": false})
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
func databaseHelper(c1 chan *mgo.Collection, mux *sync.Mutex) {
	//Set up DB connection
	session, err := mgo.Dial("mongodb:27017")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Choose database and collection
	conn := session.DB("postDB").C("posts")

	// Send back connection.
	c1 <- conn
	// Wait for the mutex to unlock before quitting
	mux.Lock()
}
