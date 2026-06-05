package dto

type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
	Status      *string `json:"status,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

type UserPublicResponse struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	Status      string  `json:"status"`
	LastSeenAt  string  `json:"last_seen_at,omitempty"`
}
