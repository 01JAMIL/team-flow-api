package websocket

import (
	"gin-api-1/internal/auth"
	"gin-api-1/internal/messages"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// GET A NORMAL HTTP REQUEST AND UPGRADES IT INTO A WEBSOCKET CONNECTION
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var hub = NewHub()

func Handler(c *gin.Context, service messages.Service) {
	loggedUser := c.MustGet("user").(auth.UserResponse)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		conn:    conn,
		send:    make(chan []byte),
		userID:  loggedUser.ID,
		service: service,
	}

	hub.AddClient(client)
	go client.WritePump()

	client.ReadPump(hub)
}
