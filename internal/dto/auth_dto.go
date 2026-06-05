package dto

type RegisterRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	InviteCode  string `json:"invite_code,omitempty"`
}

type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email"`
	Password        string `json:"password"`
	DeviceName      string `json:"device_name"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	DeviceName   string `json:"device_name"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}
