package services

import (
	"errors"

	"jia/server/internal/models"
	"jia/server/internal/repositories"
	"jia/server/internal/utils"
)

type SetupService struct {
	settingsRepo *repositories.SettingsRepository
	userRepo     *repositories.UserRepository
}

func NewSetupService(
	settingsRepo *repositories.SettingsRepository,
	userRepo *repositories.UserRepository,
) *SetupService {
	return &SetupService{
		settingsRepo: settingsRepo,
		userRepo:     userRepo,
	}
}

func (s *SetupService) IsSetupCompleted() bool {
	return s.settingsRepo.IsSetupCompleted()
}

type SetupRequest struct {
	Admin struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
		Password    string `json:"password"`
	} `json:"admin"`
	Server struct {
		Name             string `json:"name"`
		RegistrationMode string `json:"registration_mode"` // "open", "invite", "closed"
	} `json:"server"`
	S3 struct {
		Endpoint  string `json:"endpoint"`
		Bucket    string `json:"bucket"`
		AccessKey string `json:"access_key"`
		SecretKey string `json:"secret_key"`
		UseSSL    bool   `json:"use_ssl"`
	} `json:"s3"`
}

func (s *SetupService) Setup(req SetupRequest) error {
	if s.IsSetupCompleted() {
		return errors.New("setup has already been completed")
	}

	// Validate inputs
	if req.Admin.Username == "" || req.Admin.Password == "" || req.Admin.Email == "" {
		return errors.New("admin username, password and email are required")
	}
	if req.Server.Name == "" {
		req.Server.Name = "Jia"
	}
	if req.Server.RegistrationMode == "" {
		req.Server.RegistrationMode = "invite"
	}

	// Create admin user
	passHash, err := utils.HashPassword(req.Admin.Password)
	if err != nil {
		return err
	}

	adminUser := &models.User{
		Username:    req.Admin.Username,
		DisplayName: req.Admin.DisplayName,
		Email:       req.Admin.Email,
		PasswordHash: passHash,
		IsAdmin:     true,
	}

	if err := s.userRepo.Create(adminUser); err != nil {
		return err
	}

	// Save settings
	if err := s.settingsRepo.Set("server.name", req.Server.Name); err != nil {
		return err
	}
	if err := s.settingsRepo.Set("registration.mode", req.Server.RegistrationMode); err != nil {
		return err
	}
	if err := s.settingsRepo.Set("registration.max_users", 0); err != nil {
		return err
	}

	// Save S3 settings
	if err := s.settingsRepo.Set("s3.endpoint", req.S3.Endpoint); err != nil {
		return err
	}
	if err := s.settingsRepo.Set("s3.bucket", req.S3.Bucket); err != nil {
		return err
	}
	if err := s.settingsRepo.Set("s3.access_key", req.S3.AccessKey); err != nil {
		return err
	}
	if err := s.settingsRepo.Set("s3.secret_key", req.S3.SecretKey); err != nil {
		return err
	}
	if err := s.settingsRepo.Set("s3.use_ssl", req.S3.UseSSL); err != nil {
		return err
	}

	// Default file limit
	if err := s.settingsRepo.Set("upload.max_file_size", 52428800); err != nil {
		return err
	}

	// Finally, flag setup as complete
	if err := s.settingsRepo.Set("server.setup_completed", true); err != nil {
		return err
	}

	// Dynamic triggers to reload services
	InitializeDynamicServices()

	return nil
}

// Global hook to dynamically reload services after setting changes
var InitializeDynamicServices = func() {}
