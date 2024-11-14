package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096

	// Constants for implementing the perfect negotiation scheme
	polite   = 0
	impolite = 1
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	WriteBufferSize: 4096,
	ReadBufferSize:  4096,
}

type PeerRole = int

type Client struct {
	room *Room
	conn *websocket.Conn
	role PeerRole
	send chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.room.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.room.broadcast <- message
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			log.Println("Received : " + string(message))
			if !ok {
				// c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				log.Println("Error: Failed to receive message, closing connection...")
				// return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println("Error: Failed to open NextWriter")
				return
			}

			_, err = w.Write(message)
			if err != nil {
				log.Println("Error: Failed to write to clients " + err.Error())
			}

			// Flush queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			log.Println("Error: Failed to close c.Writer")
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("Error: Failed to ping the room.")
				return
			}
		}
	}
}

var clientsQueue []*Client

func serveWs(w http.ResponseWriter, r *http.Request) {
	log.Printf("Queue: %v\n", clientsQueue)
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		log.Println("Error: Error: Failed to upgrade connection to websocket")
		return
	}

	client := &Client{room: nil, conn: conn, send: make(chan []byte, 256)}
	clientsQueue = append(clientsQueue, client)

	if len(clientsQueue) > 1 {
		log.Println("Matching in process!")
		room := newRoom()
		go room.run()

		// Assign roles
		clientsQueue[0].role = impolite
		clientsQueue[1].role = polite
		clientsQueue[0].send <- []byte("{\"peerType\": \"impolite\"}")
		clientsQueue[1].send <- []byte("{\"peerType\": \"polite\"}")

		// Assign new room
		clientsQueue[0].room = room
		clientsQueue[1].room = room

		//Register clients to a room
		clientsQueue[0].room.register <- clientsQueue[0]
		clientsQueue[1].room.register <- clientsQueue[1]

		for i := 0; i < 2; i++ {
			go clientsQueue[i].writePump()
			go clientsQueue[i].readPump()
		}

		// Flush the queue
		if len(clientsQueue) > 2 {
			clientsQueue = clientsQueue[2:]
		} else {
			clientsQueue = clientsQueue[:0]
		}
	}
	log.Printf("Queue: %v\n", clientsQueue)
}
