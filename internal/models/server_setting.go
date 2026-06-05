package models

import (
	"time"
)

type ServerSetting struct {
	Key       string    `gorm:"type:varchar(100);primaryKey" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"value"` // JSON string
	UpdatedAt time.Time `json:"updated_at"`
}
