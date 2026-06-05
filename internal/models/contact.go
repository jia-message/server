package models

import (
	"time"

	"github.com/google/uuid"
)

type Contact struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_contact" json:"user_id"`
	ContactUserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_user_contact;index" json:"contact_user_id"`
	Nickname      *string   `gorm:"type:varchar(64)" json:"nickname,omitempty"`
	IsBlocked     bool      `gorm:"not null;default:false" json:"is_blocked"`
	CreatedAt     time.Time `json:"created_at"`

	// Relations
	ContactUser User `gorm:"foreignKey:ContactUserID" json:"contact_user,omitempty"`
}
