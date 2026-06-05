package dto

type CreateConversationRequest struct {
	Type           string   `json:"type"` // "dm" or "group"
	Name           *string  `json:"name,omitempty"`
	ParticipantIDs []string `json:"participant_ids"`
}

type UpdateConversationRequest struct {
	Name      *string `json:"name,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

type AddParticipantsRequest struct {
	ParticipantIDs []string `json:"participant_ids"`
}

type MarkReadRequest struct {
	MessageID string `json:"message_id"`
}
