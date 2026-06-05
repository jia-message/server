package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"jia/server/internal/dto"
	"jia/server/internal/services"
)

type KeyHandler struct {
	keyService *services.KeyService
}

func NewKeyHandler(keyService *services.KeyService) *KeyHandler {
	return &KeyHandler{
		keyService: keyService,
	}
}

func (h *KeyHandler) UploadBundle(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req dto.UploadKeysRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	err := h.keyService.RegisterDeviceBundle(
		userID,
		req.DeviceID,
		req.IdentityKey,
		req.SignedPrekey,
		req.PrekeySignature,
		req.OneTimePrekeys,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "device key bundle uploaded successfully",
	})
}

func (h *KeyHandler) GetBundle(c fiber.Ctx) error {
	targetUserIDStr := c.Params("userId")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	bundles, err := h.keyService.FetchKeyBundle(targetUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(bundles)
}

func (h *KeyHandler) ReplenishPrekeys(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req dto.ReplenishKeysRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	if err := h.keyService.ReplenishPrekeys(userID, req.DeviceID, req.OneTimePrekeys); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Get count of remaining unused prekeys
	count, _ := h.keyService.GetUnusedPrekeysCount(userID, req.DeviceID)

	return c.JSON(fiber.Map{
		"message":                 "one-time prekeys replenished successfully",
		"unused_prekeys_remaining": count,
	})
}
