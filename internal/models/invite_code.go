package models

import (
	"time"

	"github.com/google/uuid"
)

type InviteCode struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Code      string     `gorm:"type:varchar(32);uniqueIndex;not null" json:"code"`
	CreatedBy uuid.UUID  `gorm:"type:uuid;not null;index" json:"created_by"`
	MaxUses   int        `gorm:"type:integer;not null;default:1" json:"max_uses"` // 0 = unlimited
	UseCount  int        `gorm:"type:integer;not null;default:0" json:"use_count"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`

	// Relations
	Creator User `gorm:"foreignKey:CreatedBy" json:"-"`
}
