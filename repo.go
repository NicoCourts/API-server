/*
 *	repo.go provides an interface for a data repo.
 *		At the moment this will hold some dummy data while the rest of
 *		the server is developed.
 */
package main

import (
	"fmt"
	"html"
	"strconv"
)

var currentID int
var posts Posts

// Dummy data
func init() {
	RepoCreatePost(Post{
		Title:   "Write presentation",
		Visible: true,
		Body:    html.EscapeString(`<html><a href="test">Test</a></html>`),
	})
	RepoCreatePost(Post{
		Title:   "Host meetup",
		Visible: true,
	})
}

// RepoCreatePost adds a new post to our data store.
func RepoCreatePost(p Post) Post {
	id := GetNextID()
	p.ID = id
	posts = append(posts, p)
	return p
}

// GetNextID returns the ID for a post added at this moment.
func GetNextID() string {
	// For now just increment
	currentID++
	return strconv.Itoa(currentID)
}

// RepoGetPost returns the post for the given ID (if one exists). If
//	not, return a blank post.
func RepoGetPost(id string) Post {
	for _, p := range posts {
		if p.ID == id {
			return p
		}
	}
	// No matching post found
	return Post{}
}

// RepoDestroyPost deletes (disables, actually) a post.
func RepoDestroyPost(id string) error {
	for i, p := range posts {
		if p.ID == id {
			posts[i].Visible = false
			return nil
		}
	}
	return fmt.Errorf("Could not find Post with ID of %s to delete", id)
}
