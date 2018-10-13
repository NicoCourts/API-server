package main

import (
	"log"
	"net/http"

	"google.golang.org/appengine"
)

func main() {
	appengine.Main()

	router := NewRouter()

	log.Fatal(http.ListenAndServe(":8080", router))
}
