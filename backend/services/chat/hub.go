package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// EmailNotifier is an interface for sending chat email notifications.
// Defined here to avoid circular imports with the services package.
type EmailNotifier interface {
	SendChatNotificationEmail(senderID, recipientID uint) error
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	Clients map[*Client]bool

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client

	// Inbound messages from the clients.
	Broadcast chan []byte

	// Map of connected users: userID -> []*Client
	// A user can have multiple connections (multiple tabs/devices)
	UserClients map[uint][]*Client

	mu sync.RWMutex

	// Email notification for offline users
	EmailService EmailNotifier

	// Rate limiting: map["senderID:recipientID"] -> lastEmailSentAt
	emailRateLimit map[string]time.Time
	rateMu         sync.Mutex
}

const chatEmailCooldown = 5 * time.Minute

func NewHub() *Hub {
	return &Hub{
		Broadcast:      make(chan []byte),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		Clients:        make(map[*Client]bool),
		UserClients:    make(map[uint][]*Client),
		emailRateLimit: make(map[string]time.Time),
	}
}

func (h *Hub) Run() {
	rateLimitCleanup := time.NewTicker(10 * time.Minute)
	defer rateLimitCleanup.Stop()

	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.UserClients[client.UserID] = append(h.UserClients[client.UserID], client)
			h.mu.Unlock()

			// Notify others that user is online (Optional optimization: only if first connection)
			h.BroadcastStatus(client.UserID, true)

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				h.mu.Lock()
				delete(h.Clients, client)
				close(client.Send)

				// Remove from UserClients
				userConnections := h.UserClients[client.UserID]
				for i, c := range userConnections {
					if c == client {
						h.UserClients[client.UserID] = append(userConnections[:i], userConnections[i+1:]...)
						break
					}
				}

				// Check if user is completely offline
				isOffline := len(h.UserClients[client.UserID]) == 0
				if isOffline {
					delete(h.UserClients, client.UserID)
				}
				h.mu.Unlock()

				if isOffline {
					h.BroadcastStatus(client.UserID, false)
				}
			}

		case message := <-h.Broadcast:
			// Ensure message is valid JSON and extract recipient if needed
			// For now, simple broadcast to all (for testing) or specific logic can be added here
			// Real routing logic is typically handled in client.go readPump before sending to Broadcast
			// OR Broadcast channel receives a struct with recipients.

			// For simplicity in this step, we just broadcast raw bytes to everyone
			// (Refinement in client.go will target specific users)
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}

		case <-rateLimitCleanup.C:
			h.rateMu.Lock()
			for key, lastSent := range h.emailRateLimit {
				if time.Since(lastSent) > chatEmailCooldown*2 {
					delete(h.emailRateLimit, key)
				}
			}
			h.rateMu.Unlock()
		}
	}
}

// SendToUser sends a message to a specific user (all their connections)
func (h *Hub) SendToUser(userID uint, message []byte) {
	h.mu.RLock()
	clients, ok := h.UserClients[userID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			// If send buffer is full, we might close connection or skip
			// Ideally we don't close here to avoid race conditions with Run loop
			// creating a resilient system requires more complex handling,
			// but for now we skip.
		}
	}
}

// IsUserOnline checks if a user has any active WebSocket connections
func (h *Hub) IsUserOnline(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.UserClients[userID]
	return ok && len(clients) > 0
}

// NotifyOfflineUser sends an email notification if the recipient is offline
// and the rate limit allows it.
func (h *Hub) NotifyOfflineUser(senderID, recipientID uint) {
	if h.EmailService == nil {
		return
	}

	if h.IsUserOnline(recipientID) {
		return
	}

	// Rate limiting per sender-recipient pair
	key := fmt.Sprintf("%d:%d", senderID, recipientID)
	h.rateMu.Lock()
	lastSent, exists := h.emailRateLimit[key]
	if exists && time.Since(lastSent) < chatEmailCooldown {
		h.rateMu.Unlock()
		log.Printf("[Chat Email] Rate limited: %s (last sent %v ago)", key, time.Since(lastSent).Round(time.Second))
		return
	}
	h.emailRateLimit[key] = time.Now()
	h.rateMu.Unlock()

	go func() {
		if err := h.EmailService.SendChatNotificationEmail(senderID, recipientID); err != nil {
			log.Printf("[Chat Email] Error sending notification: %v", err)
		}
	}()
}

// BroadcastStatus sends a status update (online/offline) to all clients
func (h *Hub) BroadcastStatus(userID uint, isOnline bool) {
	statusMsg := map[string]interface{}{
		"type": "user_status",
		"payload": map[string]interface{}{
			"user_id":   userID,
			"is_online": isOnline,
		},
	}

	jsonMsg, err := json.Marshal(statusMsg)
	if err != nil {
		log.Printf("[Hub] Error marshaling status: %v", err)
		return
	}

	// Broadcast to everyone
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.Clients {
		// Don't send status update to the user themselves (optional)
		if client.UserID != userID {
			select {
			case client.Send <- jsonMsg:
			default:
			}
		}
	}
}
