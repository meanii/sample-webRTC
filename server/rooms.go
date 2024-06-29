package server

import (
	"errors"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/oklog/ulid/v2"
)

// Participant describes a single entity the the hashmap
type Participant struct {
	Host bool
	Conn *websocket.Conn
}

// RoomMap is the main hashmap [roomId string] -> [[]Participant]
type RoomMap struct {
	Mutext sync.RWMutex
	Map    map[string][]Participant
}

// Init initialises the RoomsMap struct
func (r *RoomMap) Init() {
	r.Map = make(map[string][]Participant)
}

// Get will return the array of participants of the room
func (r *RoomMap) Get(roomId string) []Participant {
	r.Mutext.RLock()
	defer r.Mutext.RUnlock()

	return r.Map[roomId]
}

// CreateRoom generate a unqiue roomID and return it
func (r *RoomMap) CreateRoom() string {
	r.Mutext.Lock()
	defer r.Mutext.Unlock()

	roomId := ulid.Make().String()
	log.Printf("Server::rooms::CreateRoom:Info: creating new room: %s", roomId)
	r.Map[roomId] = []Participant{}

	return roomId
}

// InsertIntoRoom will create a participant and add into hashmap
func (r *RoomMap) InsertIntoRoom(roomID string, host bool, conn *websocket.Conn) error {
	r.Mutext.Lock()
	defer r.Mutext.Unlock()

	room, ok := r.Map[roomID]
	if !ok {
		return errors.New("the roomId isn't exists yet")
	}
	log.Printf("Server::InsertIntoRoom:Info: roomId:%s\n", roomID)
	participant := Participant{Host: host, Conn: conn}
	rooms := append(room, participant)
	r.Map[roomID] = rooms
	return nil
}

// DeleteRoom delete participant from the roomId
func (r *RoomMap) DeleteRoom(roomID string) {
	r.Mutext.Lock()
	defer r.Mutext.Unlock()

	delete(r.Map, roomID)
}
