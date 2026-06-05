package middleware

import (
	"github.com/gofiber/fiber/v3"
	"jia/server/internal/services"
)

func SetupCompletedRequired(setupService *services.SetupService) fiber.Handler {
	return func(c fiber.Ctx) error {
		if !setupService.IsSetupCompleted() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "server setup is not completed. please navigate to /api/setup to configure the instance.",
			})
		}
		return c.Next()
	}
}

func SetupIncompleteRequired(setupService *services.SetupService) fiber.Handler {
	return func(c fiber.Ctx) error {
		if setupService.IsSetupCompleted() {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "setup has already been completed",
			})
		}
		return c.Next()
	}
}
