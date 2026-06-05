package models

import (
	"time"

	"github.com/google/uuid"
)

type Reaction struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MessageID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_msg_user_emoji" json:"message_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_msg_user_emoji;index" json:"user_id"`
	Emoji     string    `gorm:"type:varchar(32);not null;uniqueIndex:idx_msg_user_emoji" json:"emoji"`
	CreatedAt time.Time `json:"created_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
