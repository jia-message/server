package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	RefreshTokenHash string    `gorm:"type:varchar(255);not null" json:"-"`
	DeviceName       *string   `gorm:"type:varchar(100)" json:"device_name,omitempty"`
	IPAddress        *string   `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	ExpiresAt        time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt        time.Time `json:"created_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"-"`
}
