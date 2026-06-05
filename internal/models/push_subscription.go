package models

import (
	"time"

	"github.com/google/uuid"
)

type PushSubscription struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Platform    string    `gorm:"type:varchar(10);not null" json:"platform"` // "fcm" or "apns"
	DeviceToken string    `gorm:"type:text;uniqueIndex;not null" json:"device_token"`
	DeviceName  *string   `gorm:"type:varchar(100)" json:"device_name,omitempty"`
	IsActive    bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
