package websocket

import (
	"context"
	"encoding/json"
	"gin-api-1/internal/messages"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn    *websocket.Conn
	send    chan []byte // The Hub will put messages into this channel, and writePump will consume them.
	userID  string
	service messages.Service
}

type Hub struct {
	clients map[string]*Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
	}
}

func (h *Hub) AddClient(client *Client) {
	h.clients[client.userID] = client
}

func (h *Hub) RemoveClient(client *Client) {
	delete(h.clients, client.userID)
	close(client.send)
}

func (h *Hub) BroadcastMessage(userID string, message []byte) {
	client, exists := h.clients[userID]

	if !exists {
		return
	}

	client.send <- message
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

		var payload messages.CreateMessagePayload
		err = json.Unmarshal(message, &payload)
		if err != nil {
			client.send <- []byte(`{"error":"Invalid JSON payload"}`)
			continue
		}

		msg, err := client.service.CreateMessage(
			context.Background(),
			client.userID,
			payload,
		)

		messageResponse := messages.MessageResponse{
			ID:         msg.ID.String(),
			SenderID:   msg.SenderID.String(),
			ReceiverID: msg.ReceiverID.String(),
			Content:    msg.Content,
			CreatedAt:  msg.CreatedAt.Time,
		}

		data, err := json.Marshal(messageResponse)
		if err != nil {
			continue
		}

		hub.BroadcastMessage(payload.ReceiverID, data)
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
