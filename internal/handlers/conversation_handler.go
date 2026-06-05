package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"jia/server/internal/dto"
	"jia/server/internal/services"
)

type ConversationHandler struct {
	convService *services.ConversationService
}

func NewConversationHandler(convService *services.ConversationService) *ConversationHandler {
	return &ConversationHandler{
		convService: convService,
	}
}

func (h *ConversationHandler) List(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	conversations, err := h.convService.ListConversations(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(conversations)
}

func (h *ConversationHandler) Create(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req dto.CreateConversationRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	// Make sure creator is in participant IDs
	participantIDs := make([]uuid.UUID, 0)
	creatorFound := false
	for _, pIDStr := range req.ParticipantIDs {
		pID, err := uuid.Parse(pIDStr)
		if err != nil {
			continue
		}
		if pID == userID {
			creatorFound = true
		}
		participantIDs = append(participantIDs, pID)
	}

	if !creatorFound {
		participantIDs = append(participantIDs, userID)
	}

	conv, err := h.convService.CreateConversation(userID, req.Type, req.Name, participantIDs)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(conv)
}

func (h *ConversationHandler) GetDetails(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	convIDStr := c.Params("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid conversation id",
		})
	}

	conv, err := h.convService.GetConversationDetails(userID, convID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(conv)
}

func (h *ConversationHandler) Update(c fiber.Ctx) error {
	// Group settings update
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "group chat configuration update not fully implemented",
	})
}

func (h *ConversationHandler) AddParticipants(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	convIDStr := c.Params("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid conversation id",
		})
	}

	var req dto.AddParticipantsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	var pIDs []uuid.UUID
	for _, idStr := range req.ParticipantIDs {
		pID, err := uuid.Parse(idStr)
		if err == nil {
			pIDs = append(pIDs, pID)
		}
	}

	err = h.convService.AddParticipants(userID, convID, pIDs)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "participants added successfully",
	})
}

func (h *ConversationHandler) RemoveParticipant(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	convIDStr := c.Params("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid conversation id",
		})
	}

	targetUserIDStr := c.Params("userId")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid target user id",
		})
	}

	err = h.convService.RemoveParticipant(userID, convID, targetUserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "participant removed successfully",
	})
}

func (h *ConversationHandler) MarkRead(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	convIDStr := c.Params("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid conversation id",
		})
	}

	var req dto.MarkReadRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	msgID, err := uuid.Parse(req.MessageID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid message id",
		})
	}

	err = h.convService.MarkAsRead(userID, convID, msgID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "marked read successfully",
	})
}

func (h *ConversationHandler) Leave(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	convIDStr := c.Params("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid conversation id",
		})
	}

	err = h.convService.LeaveConversation(userID, convID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "left conversation successfully",
	})
}
