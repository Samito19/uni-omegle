package main

import (
	"log"
	"net/http"
)

func main() {
	room := newRoom()
	go room.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(room, w, r)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Failed to listen on port 8080")
	}
}
