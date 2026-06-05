package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
	"jia/server/internal/models"
	"jia/server/internal/repositories"
	"jia/server/internal/services"
)

type Hub struct {
	// Active connections by User ID
	clients    map[uuid.UUID]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex

	convRepo *repositories.ConversationRepository
}

func NewHub(convRepo *repositories.ConversationRepository) *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		convRepo:   convRepo,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
				// Broadcast online state if it's their first device
				go h.broadcastUserStatus(client.UserID, true)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			log.Printf("Client registered: %s (User: %s)", client.Conn.RemoteAddr(), client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if connections, ok := h.clients[client.UserID]; ok {
				if _, exists := connections[client]; exists {
					delete(connections, client)
					client.Close()
					if len(connections) == 0 {
						delete(h.clients, client.UserID)
						// Broadcast offline state if last device disconnected
						go h.broadcastUserStatus(client.UserID, false)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("Client unregistered: (User: %s)", client.UserID)
		}
	}
}

func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns, exists := h.clients[userID]
	return exists && len(conns) > 0
}

func (h *Hub) BroadcastToUser(userID uuid.UUID, message []byte) {
	h.mu.RLock()
	conns, exists := h.clients[userID]
	if !exists {
		h.mu.RUnlock()
		return
	}

	// Make a copy to avoid holding lock during network send
	clientList := make([]*Client, 0, len(conns))
	for c := range conns {
		clientList = append(clientList, c)
	}
	h.mu.RUnlock()

	for _, c := range clientList {
		c.Send <- message
	}
}

func (h *Hub) BroadcastToConversation(convID uuid.UUID, message []byte, excludeUserID uuid.UUID) {
	// Fetch participants
	participants, err := h.convRepo.GetByID(convID)
	if err != nil {
		return
	}

	for _, p := range participants.Participants {
		if p.LeftAt != nil {
			continue
		}
		if p.UserID == excludeUserID {
			continue
		}
		h.BroadcastToUser(p.UserID, message)
	}
}

func (h *Hub) broadcastUserStatus(userID uuid.UUID, online bool) {
	event := map[string]interface{}{
		"type": "user.online",
		"data": map[string]interface{}{
			"user_id": userID.String(),
			"online":  online,
		},
	}
	bytes, err := json.Marshal(event)
	if err != nil {
		return
	}

	// We can broadcast user status to all users who have chats with this user.
	// For simplicity in this self-hosted structure, we broadcast to all active connections.
	h.mu.RLock()
	var allClients []*Client
	for _, conns := range h.clients {
		for c := range conns {
			allClients = append(allClients, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range allClients {
		if c.UserID != userID {
			c.Send <- bytes
		}
	}
}

// BindMessageServiceHooks wires message service triggers directly to Hub broadcasts
func (h *Hub) BindMessageServiceHooks(msgService *services.MessageService, pushService *services.PushService) {
	msgService.OnMessageCreated = func(msg *models.Message) {
		event := map[string]interface{}{
			"type": "message.new",
			"data": msg,
		}
		bytes, _ := json.Marshal(event)

		// Broadcast to all participants except sender
		h.BroadcastToConversation(msg.ConversationID, bytes, msg.SenderID)

		// Send push notifications to offline participants
		participants, err := h.convRepo.GetByID(msg.ConversationID)
		if err == nil {
			for _, p := range participants.Participants {
				if p.LeftAt == nil && p.UserID != msg.SenderID {
					if !h.IsUserOnline(p.UserID) {
						pushService.SendNotification(
							p.UserID,
							msg.Sender.DisplayName,
							msg.ConversationID,
							msg.ID,
							msg.ContentType,
						)
					}
				}
			}
		}
	}

	msgService.OnMessageUpdated = func(msg *models.Message) {
		event := map[string]interface{}{
			"type": "message.updated",
			"data": msg,
		}
		bytes, _ := json.Marshal(event)
		h.BroadcastToConversation(msg.ConversationID, bytes, uuid.Nil)
	}

	msgService.OnMessageDeleted = func(msgID uuid.UUID, convID uuid.UUID) {
		event := map[string]interface{}{
			"type": "message.deleted",
			"data": map[string]string{
				"id":              msgID.String(),
				"conversation_id": convID.String(),
			},
		}
		bytes, _ := json.Marshal(event)
		h.BroadcastToConversation(convID, bytes, uuid.Nil)
	}

	msgService.OnReactionAdded = func(r *models.Reaction) {
		// Fetch conversation containing this message to know who to notify
		msgRepo := repositories.NewMessageRepository()
		msg, err := msgRepo.GetByID(r.MessageID)
		if err != nil {
			return
		}

		event := map[string]interface{}{
			"type": "reaction.new",
			"data": map[string]interface{}{
				"message_id": r.MessageID.String(),
				"user_id":    r.UserID.String(),
				"emoji":      r.Emoji,
				"user":       r.User,
			},
		}
		bytes, _ := json.Marshal(event)
		h.BroadcastToConversation(msg.ConversationID, bytes, uuid.Nil)
	}

	msgService.OnReactionRemoved = func(msgID uuid.UUID, userID uuid.UUID, emoji string) {
		msgRepo := repositories.NewMessageRepository()
		msg, err := msgRepo.GetByID(msgID)
		if err != nil {
			return
		}

		event := map[string]interface{}{
			"type": "reaction.removed",
			"data": map[string]interface{}{
				"message_id": msgID.String(),
				"user_id":    userID.String(),
				"emoji":      emoji,
			},
		}
		bytes, _ := json.Marshal(event)
		h.BroadcastToConversation(msg.ConversationID, bytes, uuid.Nil)
	}
}
