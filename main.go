package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var devmode bool

func main() {
	// Extract command line arguments
	devmode = false
	if len(os.Args) >= 2 {
		if len(os.Args) == 2 && os.Args[1] == "--dev" {
			devmode = true
		} else {
			fmt.Println("Usage:\n --dev:\t enables development parameters")
			os.Exit(1)
		}
	}

	router := NewRouter()
	log.Fatal(http.ListenAndServe(":8080", router))
}
