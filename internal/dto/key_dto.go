package dto

type UploadKeysRequest struct {
	DeviceID        string   `json:"device_id"`
	IdentityKey     string   `json:"identity_key"`
	SignedPrekey    string   `json:"signed_prekey"`
	PrekeySignature string   `json:"prekey_signature"`
	OneTimePrekeys  []string `json:"one_time_prekeys"`
}

type ReplenishKeysRequest struct {
	DeviceID       string   `json:"device_id"`
	OneTimePrekeys []string `json:"one_time_prekeys"`
}
