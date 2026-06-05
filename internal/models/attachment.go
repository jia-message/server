package models

import (
	"time"

	"github.com/google/uuid"
)

type Attachment struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MessageID         uuid.UUID `gorm:"type:uuid;not null;index" json:"message_id"`
	FileName          string    `gorm:"type:varchar(255);not null" json:"file_name"`
	StorageKey        string    `gorm:"type:text;not null" json:"-"`
	MimeType          string    `gorm:"type:varchar(127);not null" json:"mime_type"`
	FileSize          int64     `gorm:"not null" json:"file_size"`
	EncryptionKeyEnc  string    `gorm:"type:text;not null" json:"encryption_key_enc"` // AES key encrypted for recipient(s)
	CreatedAt         time.Time `json:"created_at"`

	// Relations
	Message Message `gorm:"foreignKey:MessageID" json:"-"`
}
