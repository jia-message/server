package models

import (
	"time"

	"github.com/google/uuid"
)

type ConversationParticipant struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_conv_user" json:"conversation_id"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_conv_user;index" json:"user_id"`
	Role              string     `gorm:"type:varchar(10);not null;default:'member'" json:"role"` // owner, admin, member
	LastReadMessageID *uuid.UUID `gorm:"type:uuid" json:"last_read_message_id,omitempty"`
	IsMuted           bool       `gorm:"not null;default:false" json:"is_muted"`
	JoinedAt          time.Time  `gorm:"not null" json:"joined_at"`
	LeftAt            *time.Time `json:"left_at,omitempty"`

	// Relations
	User         User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"-"`
}
