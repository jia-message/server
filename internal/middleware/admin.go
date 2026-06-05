package middleware

import (
	"github.com/gofiber/fiber/v3"
)

func AdminRequired(c fiber.Ctx) error {
	isAdmin, ok := c.Locals("is_admin").(bool)
	if !ok || !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "forbidden: administrator privileges required",
		})
	}
	return c.Next()
}
