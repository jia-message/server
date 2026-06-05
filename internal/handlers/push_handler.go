package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"jia/server/internal/services"
)

type PushHandler struct {
	pushService *services.PushService
}

func NewPushHandler(pushService *services.PushService) *PushHandler {
	return &PushHandler{
		pushService: pushService,
	}
}

type SubscribeRequest struct {
	Platform    string  `json:"platform"` // "fcm" or "apns"
	DeviceToken string  `json:"device_token"`
	DeviceName  *string `json:"device_name,omitempty"`
}

func (h *PushHandler) Subscribe(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req SubscribeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	if req.Platform != "fcm" && req.Platform != "apns" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "platform must be either 'fcm' or 'apns'",
		})
	}
	if req.DeviceToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "device_token is required",
		})
	}

	name := "Unknown Device"
	if req.DeviceName != nil {
		name = *req.DeviceName
	}

	err := h.pushService.RegisterToken(userID, req.Platform, req.DeviceToken, name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "push token registered successfully",
	})
}

type UnsubscribeRequest struct {
	DeviceToken string `json:"device_token"`
}

func (h *PushHandler) Unsubscribe(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req UnsubscribeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	if req.DeviceToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "device_token is required",
		})
	}

	err := h.pushService.UnregisterToken(userID, req.DeviceToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "push token unregistered successfully",
	})
}
