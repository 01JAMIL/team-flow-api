package websocket

import "github.com/gorilla/websocket"

type Client struct {
	conn *websocket.Conn
	send chan []byte // The Hub will put messages into this channel, and writePump will consume them.
}

type Hub struct {
	clients map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
	}
}

func (h *Hub) AddClient(client *Client) {
	h.clients[client] = true
}

func (h *Hub) RemoveClient(client *Client) {
	delete(h.clients, client)
	close(client.send)
}

func (h *Hub) BroadcastMessage(message []byte) {
	for client := range h.clients {
		client.send <- message // The Hub sends the message to the client's channel.
	}
}

func (client *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.RemoveClient(client)
		client.conn.Close()
	}()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			break
		}

		hub.BroadcastMessage(message)
	}
}

func (client *Client) WritePump() {
	defer client.conn.Close()

	for message := range client.send {
		err := client.conn.WriteMessage(
			websocket.TextMessage,
			message,
		)

		if err != nil {
			break
		}
	}
}
