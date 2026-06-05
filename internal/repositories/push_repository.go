package repositories

import (
	"github.com/google/uuid"
	"jia/server/internal/database"
	"jia/server/internal/models"
)

type PushRepository struct{}

func NewPushRepository() *PushRepository {
	return &PushRepository{}
}

func (r *PushRepository) Subscribe(sub *models.PushSubscription) error {
	var existing models.PushSubscription
	err := database.DB.Where("device_token = ?", sub.DeviceToken).First(&existing).Error
	if err == nil {
		sub.ID = existing.ID
		return database.DB.Save(sub).Error
	}
	return database.DB.Create(sub).Error
}

func (r *PushRepository) Unsubscribe(userID uuid.UUID, deviceToken string) error {
	return database.DB.Where("user_id = ? AND device_token = ?", userID, deviceToken).
		Delete(&models.PushSubscription{}).Error
}

func (r *PushRepository) GetSubscriptionsByUserID(userID uuid.UUID) ([]models.PushSubscription, error) {
	var subs []models.PushSubscription
	err := database.DB.Where("user_id = ? AND is_active = ?", userID, true).Find(&subs).Error
	return subs, err
}

func (r *PushRepository) DeactivateToken(token string) error {
	return database.DB.Model(&models.PushSubscription{}).
		Where("device_token = ?", token).
		Update("is_active", false).Error
}
