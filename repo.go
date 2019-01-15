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

	/*p := Post{
		ID:       1234,
		Title:    "This is a sample post.",
		URLTitle: "this-is-a-sample-post-1",
		Body: "<p>I am wanting to provide a longer post this time. In particular I want line breaks and " +
			"eventually to constrain the number of words that will appear in the preview.</p><p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec ipsum elit, consectetur eget ex eget, aliquet bibendum quam. Sed accumsan, leo vitae lobortis mollis, metus ante tempor enim, vel mollis lectus nisl eu erat. Mauris lorem ipsum, accumsan sit amet est aliquet, lobortis hendrerit elit. Suspendisse potenti. Fusce ac diam et ante lobortis rhoncus vehicula ac dolor. Phasellus porttitor, arcu at mollis faucibus, dui lacus vestibulum nisi, ut consectetur leo mi laoreet justo. Aenean rhoncus eget mi vitae tincidunt. Duis vitae ex quis massa tincidunt mollis. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Curabitur vel urna eget sapien pharetra commodo. Nullam maximus massa vitae enim gravida maximus. Maecenas tempus tortor fermentum quam viverra, vel iaculis dolor consequat. Vestibulum a ex vitae augue mollis condimentum. Ut finibus leo magna, non aliquet nulla hendrerit a. Vestibulum sagittis ut turpis sed iaculis.</p>",
		Date:    time.Now(),
		Visible: true,
		IsShort: false,
	}

	req, _ := http.NewRequest("GET", "/post/", nil)
	RepoCreatePost(p, req)

	RepoDestroyPost("632867513")
	*/
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
	session, err := mgo.Dial("mongodb:27017") //production
	//session, err := mgo.Dial("localhost:27017") //dev
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
