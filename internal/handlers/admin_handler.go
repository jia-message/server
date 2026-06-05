package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"jia/server/internal/dto"
	"jia/server/internal/services"
)

type AdminHandler struct {
	adminService *services.AdminService
}

func NewAdminHandler(adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

func (h *AdminHandler) GetStats(c fiber.Ctx) error {
	stats, err := h.adminService.GetStats()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(stats)
}

func (h *AdminHandler) GetSettings(c fiber.Ctx) error {
	settings, err := h.adminService.GetSettings()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(settings)
}

func (h *AdminHandler) UpdateSettings(c fiber.Ctx) error {
	var settings map[string]interface{}
	if err := c.Bind().JSON(&settings); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid settings payload",
		})
	}

	if err := h.adminService.UpdateSettings(settings); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "settings updated successfully",
	})
}

func (h *AdminHandler) ListUsers(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := h.adminService.ListUsers(page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Format output to hide private fields
	userList := make([]fiber.Map, len(users))
	for i, u := range users {
		userList[i] = fiber.Map{
			"id":           u.ID,
			"username":     u.Username,
			"display_name": u.DisplayName,
			"email":        u.Email,
			"is_admin":     u.IsAdmin,
			"status":       u.Status,
			"created_at":   u.CreatedAt,
			"last_seen_at": u.LastSeenAt,
		}
	}

	return c.JSON(fiber.Map{
		"users":      userList,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"last_page":  (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

func (h *AdminHandler) UpdateUserRole(c fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	var req dto.UpdateUserRoleRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	// Check that admin does not strip their own admin privileges
	callerID := c.Locals("user_id").(uuid.UUID)
	if callerID == userID && !req.IsAdmin {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "you cannot revoke admin privileges from yourself",
		})
	}

	if err := h.adminService.UpdateUser(userID, req.IsAdmin); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "user role updated successfully",
	})
}

func (h *AdminHandler) DeleteUser(c fiber.Ctx) error {
	userIDStr := c.Params("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	callerID := c.Locals("user_id").(uuid.UUID)
	if callerID == userID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "you cannot delete your own admin account",
		})
	}

	if err := h.adminService.DeleteUser(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "user account deleted successfully",
	})
}

func (h *AdminHandler) ListInvites(c fiber.Ctx) error {
	invites, err := h.adminService.ListInvites()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(invites)
}

func (h *AdminHandler) CreateInvite(c fiber.Ctx) error {
	var req dto.CreateInviteRequest
	if err := c.Bind().JSON(&req); err != nil {
		// Bind defaults
		req.MaxUses = 1
	}

	createdBy := c.Locals("user_id").(uuid.UUID)

	var expires *time.Time
	if req.ExpiresAt != nil {
		expires = req.ExpiresAt
	}

	invite, err := h.adminService.CreateInvite(createdBy, req.MaxUses, expires)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(invite)
}

func (h *AdminHandler) RevokeInvite(c fiber.Ctx) error {
	inviteIDStr := c.Params("id")
	inviteID, err := uuid.Parse(inviteIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid invite id",
		})
	}

	if err := h.adminService.RevokeInvite(inviteID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "invite code revoked successfully",
	})
}
