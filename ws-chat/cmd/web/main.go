package main

import (
	"log"
	"net/http"
	"ws-chat/internal/handlers"
)

func main() {

	mux := routes()

	log.Println("Starting channel listener")
	go handlers.ListenToWsChannel()

	log.Println("Starting web server on port 9000")
	_ = http.ListenAndServe(":9000", mux)
}
