package main

import (
	"fmt"
	"log"
	"net/http"
)

func handlePing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "pong")
}

func main() {
	// File server for the website's statis files
	fs := http.FileServer(http.Dir("./website/static"))
	if fs == nil {
		log.Fatal("Failed to create file server")
	}
	http.Handle("/", fs)

	http.HandleFunc("/ping", handlePing)

	fmt.Println("Server is listening at port 3000...")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
