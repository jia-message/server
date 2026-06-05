package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"jia/server/internal/database"
	"jia/server/internal/models"
)

type InviteRepository struct{}

func NewInviteRepository() *InviteRepository {
	return &InviteRepository{}
}

func (r *InviteRepository) Create(invite *models.InviteCode) error {
	return database.DB.Create(invite).Error
}

func (r *InviteRepository) GetByCode(code string) (*models.InviteCode, error) {
	var invite models.InviteCode
	err := database.DB.Where("code = ?", code).First(&invite).Error
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *InviteRepository) IncrementUseCount(code string) error {
	return database.DB.Model(&models.InviteCode{}).
		Where("code = ?", code).
		UpdateColumn("use_count", gorm.Expr("use_count + 1")).Error
}

func (r *InviteRepository) List() ([]models.InviteCode, error) {
	var invites []models.InviteCode
	err := database.DB.Order("created_at desc").Find(&invites).Error
	return invites, err
}

func (r *InviteRepository) Delete(id uuid.UUID) error {
	return database.DB.Delete(&models.InviteCode{}, id).Error
}
