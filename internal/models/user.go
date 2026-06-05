package models

import (
	"time"
)

type User struct {
	Base
	Username           string     `gorm:"type:varchar(32);uniqueIndex;not null" json:"username"`
	DisplayName        string     `gorm:"type:varchar(64);not null" json:"display_name"`
	Email              string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash       string     `gorm:"type:varchar(255);not null" json:"-"`
	AvatarURL          *string    `gorm:"type:text" json:"avatar_url,omitempty"`
	Status             string     `gorm:"type:varchar(140);default:''" json:"status"`
	IsAdmin            bool       `gorm:"not null;default:false" json:"is_admin"`
	IdentityPublicKey  *string    `gorm:"type:text" json:"identity_public_key,omitempty"` // X25519 (base64)
	SignedPrekey       *string    `gorm:"type:text" json:"signed_prekey,omitempty"`       // base64
	SignedPrekeySig    *string    `gorm:"type:text" json:"signed_prekey_signature,omitempty"`
	LastSeenAt         *time.Time `json:"last_seen_at,omitempty"`

	// Relations
	Conversations []ConversationParticipant `gorm:"foreignKey:UserID" json:"-"`
	Messages      []Message                 `gorm:"foreignKey:SenderID" json:"-"`
	Sessions      []Session                 `gorm:"foreignKey:UserID" json:"-"`
	Contacts      []Contact                 `gorm:"foreignKey:UserID" json:"-"`
}
