package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"jia/server/internal/dto"
	"jia/server/internal/repositories"
)

type UserHandler struct {
	userRepo *repositories.UserRepository
}

func NewUserHandler(userRepo *repositories.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

func (h *UserHandler) GetMe(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return c.JSON(fiber.Map{
		"id":           user.ID,
		"username":     user.Username,
		"display_name": user.DisplayName,
		"email":        user.Email,
		"status":       user.Status,
		"avatar_url":   user.AvatarURL,
		"is_admin":     user.IsAdmin,
		"created_at":   user.CreatedAt,
	})
}

func (h *UserHandler) UpdateMe(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	var req dto.UpdateProfileRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}

	if err := h.userRepo.Update(user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update profile",
		})
	}

	return c.JSON(fiber.Map{
		"message": "profile updated successfully",
		"user": fiber.Map{
			"id":           user.ID,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"status":       user.Status,
			"avatar_url":   user.AvatarURL,
		},
	})
}

func (h *UserHandler) Search(c fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "query parameter 'q' is required",
		})
	}

	users, err := h.userRepo.Search(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	results := make([]dto.UserPublicResponse, len(users))
	for i, u := range users {
		var lastSeen string
		if u.LastSeenAt != nil {
			lastSeen = u.LastSeenAt.String()
		}
		results[i] = dto.UserPublicResponse{
			ID:          u.ID.String(),
			Username:    u.Username,
			DisplayName: u.DisplayName,
			AvatarURL:   u.AvatarURL,
			Status:      u.Status,
			LastSeenAt:  lastSeen,
		}
	}

	return c.JSON(results)
}

func (h *UserHandler) GetByID(c fiber.Ctx) error {
	targetIDStr := c.Params("id")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id",
		})
	}

	u, err := h.userRepo.GetByID(targetID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	var lastSeen string
	if u.LastSeenAt != nil {
		lastSeen = u.LastSeenAt.String()
	}

	return c.JSON(dto.UserPublicResponse{
		ID:          u.ID.String(),
		Username:    u.Username,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		Status:      u.Status,
		LastSeenAt:  lastSeen,
	})
}
