package dto

import "time"

type CreateInviteRequest struct {
	MaxUses   int        `json:"max_uses"` // 0 = unlimited
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type UpdateUserRoleRequest struct {
	IsAdmin bool `json:"is_admin"`
}
