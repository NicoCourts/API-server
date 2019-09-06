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
		"POST",
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
		"PostUpdate",
		"POST",
		"/post/{postID}",
		PostUpdate,
	},
	Route{
		"ToggleVisibility",
		"POST",
		"/post/toggle/{postID}",
		PostToggle,
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
	Route{
		"UploadImage",
		"POST",
		"/upload/",
		UploadImage,
	},
	Route{
		"ImageList",
		"GET",
		"/images/",
		GetImageList,
	},
	Route{
		"Image",
		"DELETE",
		"/image/{filename}",
		ImageDelete,
	},
	Route{
		"RSSFeed",
		"GET",
		"/rss/",
		GetRSSFeed,
	},
	/*Route{
		"RsvpCreate",
		"POST",
		"/rsvp/new/",
		CreateRSVP,
	},
	Route{
		"RsvpList",
		"GET",
		"/rsvp/list/",
		ListRSVP,
	},
	Route{
		"RsvpUpdate",
		"POST",
		"/rsvp/{rescode}",
		UpdateRSVP,
	},*/
	Route{
		"RsvpFetch",
		"GET",
		"/rsvp/{rescode}",
		GetRSVP,
	},
}
