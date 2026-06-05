package models

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Type      string    `gorm:"type:varchar(10);not null;index" json:"type"` // "dm" or "group"
	Name      *string   `gorm:"type:varchar(100)" json:"name,omitempty"`
	AvatarURL *string   `gorm:"type:text" json:"avatar_url,omitempty"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null;index" json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Creator      User                      `gorm:"foreignKey:CreatedBy" json:"-"`
	Participants []ConversationParticipant `gorm:"foreignKey:ConversationID" json:"participants,omitempty"`
	Messages     []Message                 `gorm:"foreignKey:ConversationID" json:"-"`
}
