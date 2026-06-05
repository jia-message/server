package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"jia/server/internal/database"
	"jia/server/internal/models"
)

type KeyRepository struct{}

func NewKeyRepository() *KeyRepository {
	return &KeyRepository{}
}

func (r *KeyRepository) SaveDeviceKey(dk *models.DeviceKey) error {
	// First check if user already has this device registered, if so we update it.
	var existing models.DeviceKey
	err := database.DB.Where("user_id = ? AND device_id = ?", dk.UserID, dk.DeviceID).First(&existing).Error
	if err == nil {
		dk.ID = existing.ID
		return database.DB.Save(dk).Error
	}
	return database.DB.Create(dk).Error
}

func (r *KeyRepository) GetDeviceKeysByUserID(userID uuid.UUID) ([]models.DeviceKey, error) {
	var keys []models.DeviceKey
	err := database.DB.Where("user_id = ?", userID).Find(&keys).Error
	return keys, err
}

func (r *KeyRepository) GetDeviceKey(userID uuid.UUID, deviceID string) (*models.DeviceKey, error) {
	var dk models.DeviceKey
	err := database.DB.Where("user_id = ? AND device_id = ?", userID, deviceID).First(&dk).Error
	if err != nil {
		return nil, err
	}
	return &dk, nil
}

func (r *KeyRepository) SaveOneTimePrekeys(prekeys []models.OneTimePrekey) error {
	if len(prekeys) == 0 {
		return nil
	}
	return database.DB.Create(&prekeys).Error
}

// GetUnusedOneTimePrekey fetches one unused prekey for a device and marks it as used atomically inside a transaction.
func (r *KeyRepository) GetUnusedOneTimePrekey(deviceKeyID uuid.UUID) (*models.OneTimePrekey, error) {
	var prekey models.OneTimePrekey
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("device_key_id = ? AND is_used = ?", deviceKeyID, false).Order("created_at ASC").First(&prekey).Error
		if err != nil {
			return err
		}
		// Mark it as used
		prekey.IsUsed = true
		return tx.Save(&prekey).Error
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No prekeys left
		}
		return nil, err
	}
	return &prekey, nil
}

func (r *KeyRepository) CountUnusedPrekeys(deviceKeyID uuid.UUID) (int64, error) {
	var count int64
	err := database.DB.Model(&models.OneTimePrekey{}).Where("device_key_id = ? AND is_used = ?", deviceKeyID, false).Count(&count).Error
	return count, err
}
