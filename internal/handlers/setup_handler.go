package handlers

import (
	"github.com/gofiber/fiber/v3"
	"jia/server/internal/services"
)

type SetupHandler struct {
	setupService *services.SetupService
}

func NewSetupHandler(setupService *services.SetupService) *SetupHandler {
	return &SetupHandler{
		setupService: setupService,
	}
}

func (h *SetupHandler) GetStatus(c fiber.Ctx) error {
	completed := h.setupService.IsSetupCompleted()
	mode := "open"
	if completed {
		mode = h.setupService.GetRegistrationMode()
	}
	return c.JSON(fiber.Map{
		"setup_completed": completed,
		"registration_mode": mode,
	})
}

func (h *SetupHandler) Setup(c fiber.Ctx) error {
	var req services.SetupRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload: " + err.Error(),
		})
	}

	if err := h.setupService.Setup(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "server setup completed successfully",
	})
}
