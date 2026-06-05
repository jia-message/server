package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Message struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID uuid.UUID      `gorm:"type:uuid;not null;index:idx_conv_created,priority:1" json:"conversation_id"`
	SenderID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"sender_id"`
	ReplyToID      *uuid.UUID     `gorm:"type:uuid" json:"reply_to_id,omitempty"`
	ContentType    string         `gorm:"type:varchar(10);not null;default:'text'" json:"content_type"`
	Body           string         `gorm:"type:text;not null" json:"body"` // Always base64 ciphertext
	IsEdited       bool           `gorm:"not null;default:false" json:"is_edited"`
	EditedAt       *time.Time     `json:"edited_at,omitempty"`
	CreatedAt      time.Time      `gorm:"index:idx_conv_created,priority:2" json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Sender      User         `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"-"`
	ReplyTo     *Message     `gorm:"foreignKey:ReplyToID" json:"reply_to,omitempty"`
	Attachments []Attachment `gorm:"foreignKey:MessageID" json:"attachments,omitempty"`
	Reactions   []Reaction   `gorm:"foreignKey:MessageID" json:"reactions,omitempty"`
}
