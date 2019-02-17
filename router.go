package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

// NewRouter creates a new router instance
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	// First catch all OPTIONS requests
	router.Methods("OPTIONS").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Content-Length, X-Requested-With")
			w.WriteHeader(http.StatusOK)
		})

	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc

		// Wrap the handler in a logger
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}
