package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"jia/server/internal/dto"
	"jia/server/internal/models"
	"jia/server/internal/services"
)

type MessageHandler struct {
	msgService     *services.MessageService
	storageService *services.StorageService
}

func NewMessageHandler(
	msgService *services.MessageService,
	storageService *services.StorageService,
) *MessageHandler {
	return &MessageHandler{
		msgService:     msgService,
		storageService: storageService,
	}
}

func (h *MessageHandler) GetHistory(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	convIDStr := c.Params("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid conversation id",
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	beforeStr := c.Query("before")

	var before *time.Time
	if beforeStr != "" {
		t, err := time.Parse(time.RFC3339, beforeStr)
		if err == nil {
			before = &t
		}
	}

	history, err := h.msgService.GetHistory(userID, convID, before, limit)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(history)
}

func (h *MessageHandler) Send(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	convIDStr := c.Params("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid conversation id",
		})
	}

	// 1. Try checking for multipart file upload
	fileHeader, err := c.FormFile("file")
	var attachments []models.Attachment

	var body string
	var contentType string = "text"
	var replyToID *uuid.UUID

	if err == nil && fileHeader != nil {
		// File upload flow
		contentType = "file"
		if c.FormValue("content_type") != "" {
			contentType = c.FormValue("content_type") // e.g. "image"
		}

		body = c.FormValue("body")
		replyToIDStr := c.FormValue("reply_to_id")
		if replyToIDStr != "" {
			if rID, err := uuid.Parse(replyToIDStr); err == nil {
				replyToID = &rID
			}
		}

		encKey := c.FormValue("encryption_key_enc")
		if encKey == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "encryption_key_enc is mandatory for file attachments",
			})
		}

		file, err := fileHeader.Open()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "failed to open uploaded file",
			})
		}
		defer file.Close()

		// Generate random S3 key: attachments/{conversation_id}/{message_uuid}/{filename}
		messageUUID, _ := uuid.NewRandom()
		storageKey := "attachments/" + convID.String() + "/" + messageUUID.String() + "/" + fileHeader.Filename

		// Upload to S3
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		_, err = h.storageService.UploadFile(ctx, storageKey, file, fileHeader.Size, fileHeader.Header.Get("Content-Type"))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to upload file: " + err.Error(),
			})
		}

		attachments = append(attachments, models.Attachment{
			FileName:         fileHeader.Filename,
			StorageKey:        storageKey,
			MimeType:         fileHeader.Header.Get("Content-Type"),
			FileSize:         fileHeader.Size,
			EncryptionKeyEnc: encKey,
		})

	} else {
		// Plain DTO JSON flow
		var req dto.SendMessageRequest
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request payload",
			})
		}

		body = req.Body
		contentType = req.ContentType
		if req.ReplyToID != nil {
			if rID, err := uuid.Parse(*req.ReplyToID); err == nil {
				replyToID = &rID
			}
		}

		// If client has already uploaded the file directly to S3 and sent key
		if req.StorageKey != nil && req.EncryptionKeyEnc != nil {
			attachments = append(attachments, models.Attachment{
				FileName:         getOrDefault(req.FileName, "file"),
				StorageKey:        *req.StorageKey,
				MimeType:         getOrDefault(req.MimeType, "application/octet-stream"),
				FileSize:         getOrDefaultInt64(req.FileSize, 0),
				EncryptionKeyEnc: *req.EncryptionKeyEnc,
			})
		}
	}

	msg, err := h.msgService.SendMessage(userID, convID, body, contentType, replyToID, attachments)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(msg)
}

func (h *MessageHandler) Edit(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	msgIDStr := c.Params("id")
	msgID, err := uuid.Parse(msgIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid message id",
		})
	}

	var req dto.EditMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	msg, err := h.msgService.EditMessage(userID, msgID, req.Body)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(msg)
}

func (h *MessageHandler) Delete(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	msgIDStr := c.Params("id")
	msgID, err := uuid.Parse(msgIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid message id",
		})
	}

	if err := h.msgService.DeleteMessage(userID, msgID); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "message soft deleted successfully",
	})
}

func (h *MessageHandler) AddReaction(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	msgIDStr := c.Params("id")
	msgID, err := uuid.Parse(msgIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid message id",
		})
	}

	var req dto.ReactMessageRequest
	if err := c.Bind().JSON(&req); err != nil || req.Emoji == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "emoji character is required",
		})
	}

	reaction, err := h.msgService.AddReaction(userID, msgID, req.Emoji)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(reaction)
}

func (h *MessageHandler) RemoveReaction(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	msgIDStr := c.Params("id")
	msgID, err := uuid.Parse(msgIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid message id",
		})
	}

	emoji := c.Params("emoji")
	if emoji == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "emoji parameter is required",
		})
	}

	err = h.msgService.RemoveReaction(userID, msgID, emoji)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "reaction removed successfully",
	})
}

func getOrDefault(ptr *string, fallback string) string {
	if ptr != nil {
		return *ptr
	}
	return fallback
}

func getOrDefaultInt64(ptr *int64, fallback int64) int64 {
	if ptr != nil {
		return *ptr
	}
	return fallback
}
