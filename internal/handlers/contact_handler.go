package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"jia/server/internal/services"
)

type ContactHandler struct {
	contactService *services.ContactService
}

func NewContactHandler(contactService *services.ContactService) *ContactHandler {
	return &ContactHandler{
		contactService: contactService,
	}
}

func (h *ContactHandler) List(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	contacts, err := h.contactService.ListContacts(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(contacts)
}

type AddContactRequest struct {
	Username string  `json:"username"`
	Nickname *string `json:"nickname,omitempty"`
}

func (h *ContactHandler) Add(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req AddContactRequest
	if err := c.Bind().JSON(&req); err != nil || req.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username parameter is required",
		})
	}

	contact, err := h.contactService.AddContact(userID, req.Username, req.Nickname)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(contact)
}

func (h *ContactHandler) Remove(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	contactUserIDStr := c.Params("id")
	contactUserID, err := uuid.Parse(contactUserIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid contact user id",
		})
	}

	err = h.contactService.RemoveContact(userID, contactUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "contact removed successfully",
	})
}

type UpdateContactRequest struct {
	Nickname  *string `json:"nickname,omitempty"`
	IsBlocked bool    `json:"is_blocked"`
}

func (h *ContactHandler) Update(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	contactUserIDStr := c.Params("id")
	contactUserID, err := uuid.Parse(contactUserIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid contact user id",
		})
	}

	var req UpdateContactRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	err = h.contactService.UpdateContact(userID, contactUserID, req.Nickname, req.IsBlocked)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "contact updated successfully",
	})
}
