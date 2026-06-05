package models

import (
	"time"

	"github.com/google/uuid"
)

type DeviceKey struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	DeviceID         string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"device_id"`
	IdentityKey      string    `gorm:"type:text;not null" json:"identity_key"` // Per-device X25519 public key (base64)
	SignedPrekey     string    `gorm:"type:text;not null" json:"signed_prekey"`
	PrekeySignature  string    `gorm:"type:text;not null" json:"prekey_signature"`
	CreatedAt        time.Time `json:"created_at"`

	// Relations
	OneTimePrekeys []OneTimePrekey `gorm:"foreignKey:DeviceKeyID" json:"-"`
}
