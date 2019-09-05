/*repo.go provides an interface for a data repo.
 *		At the moment this will hold some dummy data while the rest of
 *		the server is developed.
 */
package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/OneOfOne/xxhash"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var currentID int

func init() {
	// Mess with the DB here

}

// RepoCreatePost adds a new post to our data store.
func RepoCreatePost(post Post) Post {
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

// RepoUpdatePost updates the title and body in the database
func RepoUpdatePost(postID string, post Input) error {
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
	var result Post
	id, _ := strconv.Atoi(postID)
	if err := c.Find(bson.M{"id": id}).One(&result); err != nil {
		log.Print("Couldn't extract post from DB")
		log.Print(err)
	}
	var e Post

	if result == e {
		return fmt.Errorf("Could not find Post with ID of %s to update", postID)
	}

	// Update Values
	if err := c.Update(result,
		bson.M{"$set": bson.M{
			"body":     post.Body,
			"markdown": post.Markdown,
			"title":    post.Title,
			"updated":  time.Now(),
		}}); err != nil {
		log.Print("Could not update post")
		return err
	}

	return nil
}

// RepoURLTitleExists checks if a urltitle is already in use and returns a boolean to that effect.
func RepoURLTitleExists(urlTitle string) bool {
	// Create channel and mutex
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex

	// Prepare mutex to hold connection open until we're done with it.
	mux.Lock()
	defer mux.Unlock()

	// Open the connection and catch the incoming pointer
	go databaseHelper(ch1, &mux)
	c := <-ch1

	count, _ := c.Find(bson.M{"urltitle": urlTitle}).Count()
	return count >= 1
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
	if err := c.Find(bson.M{"urltitle": urltitle}).One(&post); err != nil || !post.Visible {
		log.Print("Post not found!")
		log.Print(err)
		return Post{}
	}

	return post
}

// RepoTogglePost toggles visibility of a post.
func RepoTogglePost(postID string) error {
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
	if err := c.Find(bson.M{"id": id}).One(&post); err != nil {
		log.Print(err)
	}
	var e Post

	if post == e {
		return fmt.Errorf("Could not find Post with ID of %s to delete", postID)
	}

	// Toggle visibility
	if err := c.Update(post, bson.M{"$set": bson.M{"visible": !post.Visible}}); err != nil {
		log.Print(err)
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
	s := "mongodb:27017" //prod
	//s := "localhost:27017" //dev

	session, err := mgo.Dial(s)
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
	s := "mongodb:27017"
	if devmode {
		s = "localhost:27017" //dev
	}
	session, err := mgo.Dial(s) //production

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

/*// RepoCreateRSVP adds a new RSVP to our data store.
func RepoCreateRSVP(rescode string, name string, inv int) Rsvp {
	// Create channel and mutex
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex

	// Prepare mutex to hold connection open until we're done with it.
	mux.Lock()
	defer mux.Unlock()

	// Open the connection and catch the incoming pointer
	go databaseHelperRSVP(ch1, &mux)
	c := <-ch1

	// Insert a "random" ID
	h := xxhash.New32()
	h.Write([]byte(name))
	h.Write([]byte(time.Now().String()))

	// Create rsvp
	rsvp := Rsvp{
		ID:         h.Sum32(),
		Name:       name,
		ShortCode:  rescode,
		Attending:  false,
		NumInvited: inv,
		MonConfirm: 0,
		SunConfirm: 0,
	}

	// Insert rsvp
	err := c.Insert(rsvp)
	if err != nil {
		log.Fatal(err)
	}

	return rsvp
}*/

// RepoUpdateRSVP updates an RSVP
func RepoUpdateRSVP(rescode string, attending string, mon int, sun int) error {
	// Create channel and mutex
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex

	// Prepare mutex to hold connection open until we're done with it.
	mux.Lock()
	defer mux.Unlock()

	// Open the connection and catch the incoming pointer
	go databaseHelperRSVP(ch1, &mux)
	c := <-ch1

	// Update values
	att := (attending == "true")
	err := c.Update(bson.M{"shortcode": rescode}, bson.M{"$set": bson.M{
		"attending":  att,
		"updated":    true,
		"monconfirm": mon,
		"sunconfirm": sun,
	}})

	if err != nil {
		log.Print("Update failed")
		log.Print("I tried to look up " + rescode)
	}

	return err
}

/*// RepoGetRSVPs does its job
func RepoGetRSVPs() []Rsvp {
	// Create channel and mutex
	ch1 := make(chan *mgo.Collection)
	var mux sync.Mutex

	// Prepare mutex to hold connection open until we're done with it.
	mux.Lock()
	defer mux.Unlock()

	// Open the connection and catch the incoming pointer
	go databaseHelperRSVP(ch1, &mux)
	c := <-ch1

	var rsvps []Rsvp
	err := c.Find(bson.M{}).All(&rsvps)
	if err != nil {
		log.Fatal(err)
	}

	return rsvps
}*/
