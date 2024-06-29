package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/meanii/webrtc/server"
)

func main() {
	server.AllRooms.Init()

	http.HandleFunc("/create", server.CreateRoomRequestHandler)
	http.HandleFunc("/join", server.JoinRoomRequestHandler)

	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	log.Printf("Server::main: listening on http://0.0.0.0%s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
