package ws

import (
	"encoding/json"
	"log"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"jia/server/internal/services"
	"jia/server/internal/utils"
)

type WSHandler struct {
	hub         *Hub
	msgService  *services.MessageService
	convService *services.ConversationService
}

func NewWSHandler(
	hub *Hub,
	msgService *services.MessageService,
	convService *services.ConversationService,
) *WSHandler {
	wsh := &WSHandler{
		hub:         hub,
		msgService:  msgService,
		convService: convService,
	}

	// Register incoming message routing
	WSIncomingMessageCallback = wsh.handleIncoming
	return wsh
}

func (h *WSHandler) Upgrade(c fiber.Ctx) error {
	// 1. Authenticate via token query parameter (since websockets often cannot send custom headers)
	tokenStr := c.Query("token")
	if tokenStr == "" {
		// Fallback to Authorization header
		authHeader := c.Get("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenStr = authHeader[7:]
		}
	}

	if tokenStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing token",
		})
	}

	claims, err := utils.ValidateAccessToken(tokenStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	// Store user claims in locals for the upgrade
	c.Locals("user_id", claims.UserID)

	// Upgrade the connection
	return websocket.New(func(conn *websocket.Conn) {
		userID := conn.Locals("user_id").(uuid.UUID)

		client := &Client{
			Hub:    h.hub,
			Conn:   conn,
			UserID: userID,
			Send:   make(chan []byte, 256),
		}

		h.hub.register <- client

		// Start reader and writer routines
		go client.WritePump()
		client.ReadPump()
	})(c)
}

func (h *WSHandler) handleIncoming(client *Client, raw []byte) {
	var msg wsMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("WS JSON parsing error from %s: %v", client.UserID, err)
		return
	}

	switch msg.Type {
	case "message.send":
		convID, err := uuid.Parse(msg.ConversationID)
		if err != nil {
			return
		}
		// Send message
		_, err = h.msgService.SendMessage(client.UserID, convID, msg.Body, "text", nil, nil)
		if err != nil {
			log.Printf("WS SendMessage error: %v", err)
		}

	case "message.typing":
		convID, err := uuid.Parse(msg.ConversationID)
		if err != nil {
			return
		}
		// Broadcast typing to other participants
		event := map[string]interface{}{
			"type": "user.typing",
			"data": map[string]string{
				"user_id":         client.UserID.String(),
				"conversation_id": convID.String(),
			},
		}
		bytes, _ := json.Marshal(event)
		h.hub.BroadcastToConversation(convID, bytes, client.UserID)

	case "message.read":
		convID, err := uuid.Parse(msg.ConversationID)
		if err != nil {
			return
		}
		msgID, err := uuid.Parse(msg.MessageID)
		if err != nil {
			return
		}
		// Mark read cursor
		_ = h.convService.MarkAsRead(client.UserID, convID, msgID)

	case "reaction.add":
		msgID, err := uuid.Parse(msg.MessageID)
		if err != nil {
			return
		}
		_, _ = h.msgService.AddReaction(client.UserID, msgID, msg.Emoji)

	case "reaction.remove":
		msgID, err := uuid.Parse(msg.MessageID)
		if err != nil {
			return
		}
		_ = h.msgService.RemoveReaction(client.UserID, msgID, msg.Emoji)

	default:
		log.Printf("Unknown WS message type: %s", msg.Type)
	}
}
