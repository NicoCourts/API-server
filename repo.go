/* repo.go provides an interface for a data repo.
 *		At the moment this will hold some dummy data while the rest of
 *		the server is developed.
 */
package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/OneOfOne/xxhash"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var currentID int
var posts Posts

// RepoCreatePost adds a new post to our data store.
func RepoCreatePost(post Post) Post {
	// Open a database connection
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex
	mux.Lock()
	defer mux.Unlock()
	go databaseHelper(ch1, &mux)
	// Catch the incoming pointer
	c := <-ch1

	// Do the thing
	id := GetNextID(post.URLTitle, post.Date)
	post.ID = id

	err := c.Insert(post)
	if err != nil {
		log.Fatal(err)
	}

	return post
}

// GetNextID returns the ID for a post added at this moment.
func GetNextID(title string, date time.Time) uint32 {
	// hash the (url-safe) title and the time together using a fast hash
	h := xxhash.New32()
	h.Write([]byte(title))
	h.Write([]byte(date.String()))

	return h.Sum32()
}

// RepoGetPost returns the post for the given ID (if one exists). If
//	not, return a blank post.
func RepoGetPost(urltitle string) Post {
	// Open a database connection
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex
	mux.Lock()
	defer mux.Unlock()
	go databaseHelper(ch1, &mux)
	// Catch the incoming pointer
	c := <-ch1

	var post Post
	err := c.Find(bson.M{"urltitle": urltitle}).One(&post)
	if err != nil {
		log.Fatal(err)
	}

	return post
}

// RepoDestroyPost deletes (disables, actually) a post.
func RepoDestroyPost(urltitle string) error {
	// Open a database connection
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex
	mux.Lock()
	defer mux.Unlock()
	go databaseHelper(ch1, &mux)
	// Catch the incoming pointer
	c := <-ch1

	// Find post, if it exists
	var post Post
	err := c.Find(bson.M{"urltitle": urltitle}).One(&post)
	var e Post
	if post == e {
		return fmt.Errorf("Could not find Post with ID of %s to delete", urltitle)

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
	// Open a database connection
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex
	mux.Lock()
	defer mux.Unlock()
	go databaseHelper(ch1, &mux)
	// Catch the incoming pointer
	c := <-ch1

	var posts Posts
	err := c.Find(bson.M{"visible": true}).All(&posts)
	if err != nil {
		log.Fatal(err)
	}

	return posts
}

// databaseHelper does the work of opening the database
func databaseHelper(c1 chan *mgo.Collection, mux *sync.Mutex) {
	//Set up DB connection
	session, err := mgo.Dial("localhost:27017")
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
