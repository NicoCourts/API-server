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
		"Posts List",
		"GET",
		"/posts",
		PostIndex,
	},
	Route{
		"Post Show",
		"GET",
		"/posts/{postID}",
		PostShow,
	},
	Route{
		"Post Create",
		"POST",
		"/posts",
		PostCreate,
	},
	Route{
		"Post Delete",
		"DELETE",
		"/posts/{postID}",
		PostDelete,
	},
}
