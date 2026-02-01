package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"server/internal/server"
	"server/internal/server/clients"
)

var (
	port = flag.Int("port", 8080, "port to run server on")
)

func main() {
	flag.Parse()

	// Define the game hub
	hub := server.NewHub()

	// define the handler for the websocket connections
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewWebSocketClient, w, r)
	})

	go hub.Run()
	addr := fmt.Sprintf(":%d", *port)

	log.Printf("Server started on %s", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
