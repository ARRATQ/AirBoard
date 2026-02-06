package chat

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"airboard/models"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allo cross-origin for development
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Hub *Hub

	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	Send chan []byte

	// DB connection for persisting messages
	DB *gorm.DB

	// Authenticated User ID
	UserID uint
}

// WSMessage represents the structure of messages sent over WebSocket
type WSMessage struct {
	Type      string      `json:"type"`                   // "chat_message", "typing", ...
	Payload   interface{} `json:"payload"`                // The actual data
	Recipient *uint       `json:"recipient_id,omitempty"` // For private messages
	GroupID   *uint       `json:"group_id,omitempty"`     // For group messages
}

// IncomingMessage represents what we expect to receive from the client
type IncomingMessage struct {
	Content     string `json:"content"`
	RecipientID *uint  `json:"recipient_id,omitempty"`
	GroupID     *uint  `json:"group_id,omitempty"`
}

// readPump pumps messages from the websocket connection to the hub.
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] error: %v", err)
			}
			break
		}

		// Process message
		c.handleMessage(message)
	}
}

func (c *Client) handleMessage(msg []byte) {
	var incoming IncomingMessage
	if err := json.Unmarshal(msg, &incoming); err != nil {
		log.Printf("[WS] Invalid JSON format: %v", err)
		return
	}

	if incoming.Content == "" {
		return
	}

	// Persist to DB
	chatMsg := models.ChatMessage{
		Content:     incoming.Content,
		SenderID:    c.UserID,
		RecipientID: incoming.RecipientID,
		GroupID:     incoming.GroupID,
		Type:        "text",
		CreatedAt:   time.Now(),
	}

	// Transaction to ensure data consistency
	if err := c.DB.Create(&chatMsg).Error; err != nil {
		log.Printf("[WS] DB Save Error: %v", err)
		return
	}

	// Load sender for response
	// c.DB.Preload("Sender").First(&chatMsg, chatMsg.ID) // Optional optimization

	// Prepare broadcast message
	responseMsg := WSMessage{
		Type:    "chat_message",
		Payload: chatMsg,
	}

	jsonResponse, _ := json.Marshal(responseMsg)

	// Route the message
	if incoming.RecipientID != nil {
		// DM: Send to recipient AND sender (so it appears in their own chat window immediately via WS)
		c.Hub.SendToUser(*incoming.RecipientID, jsonResponse)
		c.Hub.SendToUser(c.UserID, jsonResponse)
	} else if incoming.GroupID != nil {
		// Group Chat: Logic to find users in group and send to all
		// TODO: Implement group broadcast optimization
		// For now, let's assume we implement a general BroadcastToGroup function in Hub
		// or just rely on the frontend fetching or using a simpler broadcast loop here.
		// For MVP, if we don't have Hub.SendToGroup, we might loop.
		// But better practice: fetch group members.
		var groupMembers []models.User
		c.DB.Joins("JOIN user_groups on user_groups.user_id = users.id").
			Where("user_groups.group_id = ?", *incoming.GroupID).
			Find(&groupMembers)

		for _, member := range groupMembers {
			c.Hub.SendToUser(member.ID, jsonResponse)
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
