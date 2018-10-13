package main

import (
	"net/http"
)

// Route is a template for a specific route
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes is an array of Route
type Routes []Route

// Collection of routes the user can take with our API
var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},
	Route{
		"ListPosts",
		"GET",
		"/posts/",
		PostIndex,
	},
	Route{
		"ListAllPosts",
		"GET",
		"/posts/all/",
		AllPostIndex,
	},
	Route{
		"PostShow",
		"GET",
		"/post/{postID}",
		PostShow,
	},
	Route{
		"PostCreate",
		"POST",
		"/post/",
		PostCreate,
	},
	Route{
		"PostDelete",
		"DELETE",
		"/post/{postID}",
		PostDelete,
	},
	Route{
		"ReadNonce",
		"GET",
		"/nonce/",
		ReadNonce,
	},
	Route{
		"UpdateNonce",
		"GET",
		"/nonce/update/",
		NonceUpdate,
	},
}
