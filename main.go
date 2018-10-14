package main

import (
	"os"
	//"log"
	"net/http"

	"github.com/gorilla/handlers"
	"google.golang.org/appengine"
)

func main() {
	router := NewRouter()

	//log.Fatal(http.ListenAndServe(":80", router))
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stderr, router))
	appengine.Main()
}
