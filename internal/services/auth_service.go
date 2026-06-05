package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"
	"jia/server/internal/models"
	"jia/server/internal/repositories"
	"jia/server/internal/utils"
)

type AuthService struct {
	userRepo     *repositories.UserRepository
	sessionRepo  *repositories.SessionRepository
	settingsRepo *repositories.SettingsRepository
	inviteRepo   *repositories.InviteRepository
}

func NewAuthService(
	userRepo *repositories.UserRepository,
	sessionRepo *repositories.SessionRepository,
	settingsRepo *repositories.SettingsRepository,
	inviteRepo *repositories.InviteRepository,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		settingsRepo: settingsRepo,
		inviteRepo:   inviteRepo,
	}
}

func (s *AuthService) Register(username, displayName, email, password, inviteCode string) (*models.User, error) {
	// 1. Check registration mode
	var regMode string
	err := s.settingsRepo.Get("registration.mode", &regMode)
	if err != nil {
		regMode = "invite" // Default fallback
	}

	if regMode == "closed" {
		return nil, errors.New("registration is closed on this server")
	}

	// 2. Validate invite code if in invite mode
	if regMode == "invite" {
		if inviteCode == "" {
			return nil, errors.New("invite code is required for registration")
		}
		inv, err := s.inviteRepo.GetByCode(inviteCode)
		if err != nil {
			return nil, errors.New("invalid invite code")
		}

		if inv.ExpiresAt != nil && inv.ExpiresAt.Before(time.Now()) {
			return nil, errors.New("invite code has expired")
		}

		if inv.MaxUses > 0 && inv.UseCount >= inv.MaxUses {
			return nil, errors.New("invite code has reached its maximum usage limit")
		}

		// Increment use count
		if err := s.inviteRepo.IncrementUseCount(inviteCode); err != nil {
			return nil, err
		}
	}

	// 3. Verify maximum users limit
	var maxUsers int
	if err := s.settingsRepo.Get("registration.max_users", &maxUsers); err == nil && maxUsers > 0 {
		currentCount, err := s.userRepo.Count()
		if err == nil && currentCount >= int64(maxUsers) {
			return nil, errors.New("server has reached its maximum registered user limit")
		}
	}

	// 4. Create user
	passHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Verify uniqueness
	if _, err := s.userRepo.GetByUsername(username); err == nil {
		return nil, errors.New("username is already taken")
	}
	if _, err := s.userRepo.GetByEmail(email); err == nil {
		return nil, errors.New("email is already registered")
	}

	user := &models.User{
		Username:     username,
		DisplayName:  displayName,
		Email:        email,
		PasswordHash: passHash,
		IsAdmin:      false, // default to non-admin
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(usernameOrEmail, password, deviceName, ipAddress string) (string, string, error) {
	var user *models.User
	var err error

	// Try fetching by username first, then by email
	user, err = s.userRepo.GetByUsername(usernameOrEmail)
	if err != nil {
		user, err = s.userRepo.GetByEmail(usernameOrEmail)
		if err != nil {
			return "", "", errors.New("invalid username or password")
		}
	}

	// Verify password
	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		return "", "", errors.New("invalid username or password")
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.IsAdmin)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// Save session
	tokenHash := hashToken(refreshToken)
	session := &models.Session{
		UserID:           user.ID,
		RefreshTokenHash: tokenHash,
		DeviceName:       &deviceName,
		IPAddress:        &ipAddress,
		ExpiresAt:        time.Now().Add(30 * 24 * time.Hour), // 30 days
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return "", "", err
	}

	// Update user's last seen at
	now := time.Now()
	user.LastSeenAt = &now
	_ = s.userRepo.Update(user)

	return accessToken, refreshToken, nil
}

func (s *AuthService) Refresh(refreshToken, deviceName, ipAddress string) (string, string, error) {
	tokenHash := hashToken(refreshToken)
	session, err := s.sessionRepo.GetByRefreshToken(tokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errors.New("invalid or expired refresh token")
		}
		return "", "", err
	}

	user, err := s.userRepo.GetByID(session.UserID)
	if err != nil {
		return "", "", err
	}

	// Generate new tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.IsAdmin)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// Delete old session
	_ = s.sessionRepo.Delete(session.ID)

	// Create new session
	newSession := &models.Session{
		UserID:           user.ID,
		RefreshTokenHash: hashToken(newRefreshToken),
		DeviceName:       &deviceName,
		IPAddress:        &ipAddress,
		ExpiresAt:        time.Now().Add(30 * 24 * time.Hour),
	}

	if err := s.sessionRepo.Create(newSession); err != nil {
		return "", "", err
	}

	// Update last seen
	now := time.Now()
	user.LastSeenAt = &now
	_ = s.userRepo.Update(user)

	return accessToken, newRefreshToken, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	tokenHash := hashToken(refreshToken)
	session, err := s.sessionRepo.GetByRefreshToken(tokenHash)
	if err != nil {
		return nil // Already logged out or invalid
	}
	return s.sessionRepo.Delete(session.ID)
}

func hashToken(token string) string {
	hasher := sha256.New()
	hasher.Write([]byte(token))
	return hex.EncodeToString(hasher.Sum(nil))
}
