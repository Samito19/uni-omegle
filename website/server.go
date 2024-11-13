package main

import (
	"fmt"
	"log"
	"net/http"
)

func handlePing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "pong")
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(fmt.Sprintf("%v %v %v %v", r.RemoteAddr, r.Method, r.URL, r.UserAgent()))
		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("Server is listening at port 3000...")
	mux := http.NewServeMux()

	// File server for the website's statis files
	fs := http.FileServer(http.Dir("./website/static"))
	if fs == nil {
		log.Fatal("Failed to create file server")
	}

	mux.Handle("/", LoggingMiddleware(fs))
	mux.HandleFunc("/ping", handlePing)

	log.Fatal(http.ListenAndServe(":3000", LoggingMiddleware(mux)))
}
