package handlers

import (
	"github.com/gofiber/fiber/v3"
	"jia/server/internal/dto"
	"jia/server/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	if req.Username == "" || req.Password == "" || req.Email == "" || req.DisplayName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username, password, email, and display name are required",
		})
	}

	user, err := h.authService.Register(req.Username, req.DisplayName, req.Email, req.Password, req.InviteCode)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "registration successful",
		"user": fiber.Map{
			"id":           user.ID,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"email":        user.Email,
		},
	})
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	if req.UsernameOrEmail == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username or email and password are required",
		})
	}

	deviceName := req.DeviceName
	if deviceName == "" {
		deviceName = c.Get("User-Agent", "Unknown Device")
	}
	ipAddress := c.IP()

	accessToken, refreshToken, err := h.authService.Login(req.UsernameOrEmail, req.Password, deviceName, ipAddress)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    900, // 15 mins in seconds
	})
}

func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var req dto.RefreshRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "refresh token is required",
		})
	}

	deviceName := req.DeviceName
	if deviceName == "" {
		deviceName = c.Get("User-Agent", "Unknown Device")
	}
	ipAddress := c.IP()

	accessToken, newRefreshToken, err := h.authService.Refresh(req.RefreshToken, deviceName, ipAddress)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    900,
	})
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var req dto.LogoutRequest
	if err := c.Bind().JSON(&req); err != nil {
		// Try parsing from query or headers if JSON body isn't sent
		req.RefreshToken = c.FormValue("refresh_token")
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "refresh token is required",
		})
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to logout",
		})
	}

	return c.JSON(fiber.Map{
		"message": "logged out successfully",
	})
}
