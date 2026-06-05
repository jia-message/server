package models

import (
	"time"

	"github.com/google/uuid"
)

type OneTimePrekey struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DeviceKeyID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_device_key_pub" json:"device_key_id"`
	PublicKey   string    `gorm:"type:text;not null;uniqueIndex:idx_device_key_pub" json:"public_key"` // base64
	IsUsed      bool      `gorm:"not null;default:false;index" json:"is_used"`
	CreatedAt   time.Time `json:"created_at"`
}
