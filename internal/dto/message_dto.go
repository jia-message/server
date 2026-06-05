package dto

type SendMessageRequest struct {
	Body             string     `json:"body"` // ciphertext
	ReplyToID        *string    `json:"reply_to_id,omitempty"`
	ContentType      string     `json:"content_type"`                 // "text" or "image" or "file"
	FileName         *string    `json:"file_name,omitempty"`          // for file attachments
	FileSize         *int64     `json:"file_size,omitempty"`          // for file attachments
	MimeType         *string    `json:"mime_type,omitempty"`          // for file attachments
	EncryptionKeyEnc *string    `json:"encryption_key_enc,omitempty"` // for file attachments
	StorageKey       *string    `json:"storage_key,omitempty"`        // if uploaded directly
}

type EditMessageRequest struct {
	Body string `json:"body"`
}

type ReactMessageRequest struct {
	Emoji string `json:"emoji"`
}
