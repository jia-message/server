package services

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
	"jia/server/internal/models"
	"jia/server/internal/repositories"
)

type AdminService struct {
	userRepo     *repositories.UserRepository
	settingsRepo *repositories.SettingsRepository
	inviteRepo   *repositories.InviteRepository
	messageRepo  *repositories.MessageRepository
}

func NewAdminService(
	userRepo *repositories.UserRepository,
	settingsRepo *repositories.SettingsRepository,
	inviteRepo *repositories.InviteRepository,
	messageRepo *repositories.MessageRepository,
) *AdminService {
	return &AdminService{
		userRepo:     userRepo,
		settingsRepo: settingsRepo,
		inviteRepo:   inviteRepo,
		messageRepo:  messageRepo,
	}
}

func (s *AdminService) GetStats() (map[string]interface{}, error) {
	userCount, err := s.userRepo.Count()
	if err != nil {
		return nil, err
	}

	msgCount, err := s.messageRepo.Count()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_users":    userCount,
		"total_messages": msgCount,
		"uptime_seconds": time.Since(ServerStartTime).Seconds(),
	}

	return stats, nil
}

var ServerStartTime = time.Now()

func (s *AdminService) GetSettings() (map[string]interface{}, error) {
	return s.settingsRepo.GetAll()
}

func (s *AdminService) UpdateSettings(settings map[string]interface{}) error {
	for k, v := range settings {
		if k == "server.setup_completed" {
			continue // Cannot update setup completed state through normal settings API
		}
		if err := s.settingsRepo.Set(k, v); err != nil {
			return err
		}
	}

	// Dynamic triggers to reload services
	InitializeDynamicServices()

	return nil
}

func (s *AdminService) ListUsers(page, pageSize int) ([]models.User, int64, error) {
	return s.userRepo.List(page, pageSize)
}

func (s *AdminService) UpdateUser(userID uuid.UUID, isAdmin bool) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	user.IsAdmin = isAdmin
	return s.userRepo.Update(user)
}

func (s *AdminService) DeleteUser(userID uuid.UUID) error {
	return s.userRepo.Delete(userID)
}

func (s *AdminService) ListInvites() ([]models.InviteCode, error) {
	return s.inviteRepo.List()
}

func (s *AdminService) CreateInvite(createdBy uuid.UUID, maxUses int, expiresAt *time.Time) (*models.InviteCode, error) {
	code := generateRandomInviteCode()
	invite := &models.InviteCode{
		Code:      code,
		CreatedBy: createdBy,
		MaxUses:   maxUses,
		UseCount:  0,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := s.inviteRepo.Create(invite); err != nil {
		return nil, err
	}

	return invite, nil
}

func (s *AdminService) RevokeInvite(inviteID uuid.UUID) error {
	return s.inviteRepo.Delete(inviteID)
}

const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomInviteCode() string {
	b := make([]byte, 8)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	return string(b)
}
