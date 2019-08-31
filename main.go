package main

import (
	"log"
	"net/http"
)

var devmode bool

func main() {
	router := NewRouter()
	log.Fatal(http.ListenAndServe(":8080", router))
}
