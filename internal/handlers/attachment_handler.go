package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"jia/server/internal/repositories"
	"jia/server/internal/services"
)

type AttachmentHandler struct {
	msgRepo        *repositories.MessageRepository
	convRepo       *repositories.ConversationRepository
	storageService *services.StorageService
}

func NewAttachmentHandler(
	msgRepo *repositories.MessageRepository,
	convRepo *repositories.ConversationRepository,
	storageService *services.StorageService,
) *AttachmentHandler {
	return &AttachmentHandler{
		msgRepo:        msgRepo,
		convRepo:       convRepo,
		storageService: storageService,
	}
}

func (h *AttachmentHandler) GetPresignedURL(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	attIDStr := c.Params("id")
	attID, err := uuid.Parse(attIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid attachment id",
		})
	}

	// Fetch attachment details
	att, err := h.msgRepo.GetAttachmentByID(attID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "attachment not found",
		})
	}

	// Fetch parent message
	msg, err := h.msgRepo.GetByID(att.MessageID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "associated message not found",
		})
	}

	// Verify caller participates in conversation
	_, err = h.convRepo.GetParticipant(msg.ConversationID, userID)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access denied: you are not a participant in this conversation",
		})
	}

	// Generate presigned GET URL (expires in 1 hour)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	urlStr, err := h.storageService.GetPresignedURL(ctx, att.StorageKey, time.Hour)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate download URL: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"url":                urlStr,
		"expires_in_seconds": 3600,
		"encryption_key_enc": att.EncryptionKeyEnc,
	})
}
