package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/bocanada/rest-ws/helpers"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	clients    []*Client
	register   chan *Client
	unregister chan *Client
	mutex      *sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make([]*Client, 0),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		mutex:      &sync.Mutex{},
	}
}

func (hub *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("HandleWebSocket: ", err)
		helpers.NewResponseError(err).Send(w, http.StatusBadRequest)
		return
	}
	client := NewClient(hub, socket)
	hub.register <- client
	go client.Write()
}

func (hub *Hub) onConnect(client *Client) {
	log.Println("Client connected: ", client.socket.RemoteAddr())
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	client.id = client.socket.RemoteAddr().String()
	hub.clients = append(hub.clients, client)
	log.Println("Assigned id: ", client.id)

}

func (hub *Hub) onDisconnect(client *Client) {
	log.Println("Client disconnected: ", client.socket.RemoteAddr(), "with id: ", client.id)
	client.socket.Close()
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	i := -1
	for j, c := range hub.clients {
		if c.id == client.id {
			i = j
			break
		}
	}
	copy(hub.clients[i:], hub.clients[i+1:])
	hub.clients[len(hub.clients)-1] = nil
	hub.clients = hub.clients[:len(hub.clients)-1]
}

func (hub *Hub) Run() {
	for {
		select {
		case client := <-hub.register:
			hub.onConnect(client)
		case client := <-hub.unregister:
			hub.onDisconnect(client)
		}
	}
}

func (hub *Hub) Broadcast(message any, ignore *Client) {
	data, _ := json.Marshal(message)
	for _, c := range hub.clients {
		if ignore != nil && c.id == ignore.id {
			continue
		}
		c.outbound <- data
	}
}