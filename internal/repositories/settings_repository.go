package repositories

import (
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/gorm"
	"jia/server/internal/database"
	"jia/server/internal/models"
	"jia/server/internal/utils"
)

type SettingsRepository struct{}

func NewSettingsRepository() *SettingsRepository {
	return &SettingsRepository{}
}

// IsSensitiveKey returns true if the key's value should be encrypted at rest
func (r *SettingsRepository) IsSensitiveKey(key string) bool {
	sensitiveKeys := []string{
		"s3.access_key",
		"s3.secret_key",
		"push.fcm_credentials",
		"push.apns_key",
	}
	for _, sk := range sensitiveKeys {
		if strings.EqualFold(key, sk) {
			return true
		}
	}
	return false
}

func (r *SettingsRepository) Get(key string, target interface{}) error {
	var setting models.ServerSetting
	err := database.DB.Where("key = ?", key).First(&setting).Error
	if err != nil {
		return err
	}

	valStr := setting.Value
	if r.IsSensitiveKey(key) {
		decrypted, err := utils.Decrypt(valStr)
		if err != nil {
			return err
		}
		valStr = decrypted
	}

	return json.Unmarshal([]byte(valStr), target)
}

func (r *SettingsRepository) Set(key string, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	valStr := string(bytes)
	if r.IsSensitiveKey(key) {
		encrypted, err := utils.Encrypt(valStr)
		if err != nil {
			return err
		}
		valStr = encrypted
	}

	setting := models.ServerSetting{
		Key:   key,
		Value: valStr,
	}

	return database.DB.Save(&setting).Error
}

func (r *SettingsRepository) IsSetupCompleted() bool {
	var completed bool
	err := r.Get("server.setup_completed", &completed)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
		return false
	}
	return completed
}

func (r *SettingsRepository) GetAll() (map[string]interface{}, error) {
	var dbSettings []models.ServerSetting
	err := database.DB.Find(&dbSettings).Error
	if err != nil {
		return nil, err
	}

	results := make(map[string]interface{})
	for _, s := range dbSettings {
		var val interface{}
		valStr := s.Value

		if r.IsSensitiveKey(s.Key) {
			// Decrypt sensitive keys but mask them for safety in normal settings dump
			decrypted, err := utils.Decrypt(valStr)
			if err == nil {
				valStr = decrypted
			}
			// Mask in JSON dump
			results[s.Key] = "********"
			continue
		}

		if err := json.Unmarshal([]byte(valStr), &val); err == nil {
			results[s.Key] = val
		} else {
			results[s.Key] = valStr
		}
	}
	return results, nil
}
