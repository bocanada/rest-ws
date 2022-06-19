package websocket

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	hub      *Hub
	id       string
	socket   *websocket.Conn
	outbound chan []byte
}

func NewClient(hub *Hub, socket *websocket.Conn) *Client {
	return &Client{
		hub:      hub,
		socket:   socket,
		outbound: make(chan []byte),
	}
}

func (c *Client) Write() {
	defer func() {
		c.socket.WriteMessage(websocket.CloseMessage, []byte{})
		c.hub.unregister <- c
	}()
	for {
		select {
		case message, ok := <-c.outbound:
			if !ok {
				return
			}
			if err := c.socket.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}
