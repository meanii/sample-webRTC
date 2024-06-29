package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// AllRooms is the global hashmap for the server
var AllRooms RoomMap

// upgrader is the websocket upgrader, allowing for all origins
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type broadcastMessage struct {
	Message map[string]interface{}
	RoomId  string
	Client  *websocket.Conn
}

var broadcast = make(chan broadcastMessage)

// broadcaster should run once, not for every connection
func broadcaster() {
	for {
		message := <-broadcast
		log.Printf("Server::signalling::broadcaster:Info: received new Message from RoomId('%s')", message.RoomId)
		for _, client := range AllRooms.Map[message.RoomId] {
			if client.Conn != message.Client {
				log.Printf("Server::signalling::broadcast::WritingMessage: from %s to %s", message.Client.RemoteAddr(), client.Conn.RemoteAddr())
				err := client.Conn.WriteJSON(message.Message)
				if err != nil {
					log.Printf("Server::signalling::broadcaster::Error: encountered error while writing message to %s roomId", message.RoomId)
					client.Conn.Close()
				}
			}
		}
	}
}

// CreateRoomRequestHandler creates a room and returns the roomId
func CreateRoomRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	roomId := AllRooms.CreateRoom()
	type resp struct {
		RoomId string `json:"room_id"`
	}

	log.Printf("Server::signalling::CreateRoomRequestHandler:Info: created new room: %s", roomId)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp{RoomId: roomId})
}

// JoinRoomRequestHandler joins the client to a particular room
func JoinRoomRequestHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("roomID")
	if len(roomID) == 0 {
		log.Printf("Server::signalling::JoinRoomRequestHandler:Error: roomID is missing in URL parameters")
		return
	}

	// upgrading connection from HTTP to WS
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Server::signalling::JoinRoomRequestHandler:Error: HTTP connection couldn't upgrade to WS.\nDetails: %s", err)
		return
	}

	log.Printf("Server::signalling::JoinRoomRequestHandler:Info: successfully upgraded the connection from HTTP to WS for roomID (%s)", roomID)
	AllRooms.InsertIntoRoom(roomID, false, ws)

	// Handle messages from the client
	go func() {
		for {
			var message broadcastMessage
			err := ws.ReadJSON(&message.Message)
			if err != nil {
				log.Printf("Server::signalling::JoinRoomRequestHandler:Error: encountered error while reading the message from WS.\nDetails: %s", err)
				return
			}

			message.Client = ws
			message.RoomId = roomID

			broadcast <- message
		}
	}()
}

func init() {
	go broadcaster() // start the broadcaster once during server initialization
}
