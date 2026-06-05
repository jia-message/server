package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"jia/server/internal/models"
	"jia/server/internal/repositories"
)

type MessageService struct {
	msgRepo  *repositories.MessageRepository
	convRepo *repositories.ConversationRepository

	// WebSocket broadcast hooks (populated by ws package)
	OnMessageCreated  func(msg *models.Message)
	OnMessageUpdated  func(msg *models.Message)
	OnMessageDeleted  func(msgID uuid.UUID, convID uuid.UUID)
	OnReactionAdded   func(reaction *models.Reaction)
	OnReactionRemoved func(msgID uuid.UUID, userID uuid.UUID, emoji string)
}

func NewMessageService(
	msgRepo *repositories.MessageRepository,
	convRepo *repositories.ConversationRepository,
) *MessageService {
	return &MessageService{
		msgRepo:  msgRepo,
		convRepo: convRepo,
	}
}

func (s *MessageService) SendMessage(
	senderID uuid.UUID,
	convID uuid.UUID,
	body string,
	contentType string,
	replyToID *uuid.UUID,
	attachments []models.Attachment,
) (*models.Message, error) {
	// Verify conversation participant
	_, err := s.convRepo.GetParticipant(convID, senderID)
	if err != nil {
		return nil, errors.New("access denied: not a participant")
	}

	msg := &models.Message{
		ConversationID: convID,
		SenderID:       senderID,
		ReplyToID:      replyToID,
		ContentType:    contentType,
		Body:           body,
		CreatedAt:      time.Now(),
	}

	if err := s.msgRepo.Create(msg); err != nil {
		return nil, err
	}

	// Save attachments if any
	for i := range attachments {
		attachments[i].MessageID = msg.ID
		attachments[i].CreatedAt = time.Now()
		_ = s.msgRepo.SaveAttachment(&attachments[i])
	}

	// Reload to get preloads
	fullMsg, err := s.msgRepo.GetByID(msg.ID)
	if err == nil {
		// Broadcast via websocket
		if s.OnMessageCreated != nil {
			s.OnMessageCreated(fullMsg)
		}
	}

	return fullMsg, err
}

func (s *MessageService) GetHistory(userID uuid.UUID, convID uuid.UUID, before *time.Time, limit int) ([]models.Message, error) {
	// Verify participation
	_, err := s.convRepo.GetParticipant(convID, userID)
	if err != nil {
		return nil, errors.New("access denied")
	}

	if limit <= 0 || limit > 100 {
		limit = 50 // Default history limit
	}

	return s.msgRepo.GetHistory(convID, before, limit)
}

func (s *MessageService) EditMessage(userID uuid.UUID, msgID uuid.UUID, newBody string) (*models.Message, error) {
	msg, err := s.msgRepo.GetByID(msgID)
	if err != nil {
		return nil, err
	}

	if msg.SenderID != userID {
		return nil, errors.New("unauthorized to edit this message")
	}

	msg.Body = newBody
	msg.IsEdited = true
	now := time.Now()
	msg.EditedAt = &now

	if err := s.msgRepo.Update(msg); err != nil {
		return nil, err
	}

	// Reload
	fullMsg, err := s.msgRepo.GetByID(msg.ID)
	if err == nil && s.OnMessageUpdated != nil {
		s.OnMessageUpdated(fullMsg)
	}

	return fullMsg, err
}

func (s *MessageService) DeleteMessage(userID uuid.UUID, msgID uuid.UUID) error {
	msg, err := s.msgRepo.GetByID(msgID)
	if err != nil {
		return err
	}

	// Message can be deleted by sender, or by conversation owner/admin
	if msg.SenderID != userID {
		cp, err := s.convRepo.GetParticipant(msg.ConversationID, userID)
		if err != nil || (cp.Role != "owner" && cp.Role != "admin") {
			return errors.New("unauthorized to delete this message")
		}
	}

	if err := s.msgRepo.Delete(msgID); err != nil {
		return err
	}

	if s.OnMessageDeleted != nil {
		s.OnMessageDeleted(msgID, msg.ConversationID)
	}

	return nil
}

func (s *MessageService) AddReaction(userID uuid.UUID, msgID uuid.UUID, emoji string) (*models.Reaction, error) {
	msg, err := s.msgRepo.GetByID(msgID)
	if err != nil {
		return nil, err
	}

	// Check if user is participant of conversation
	_, err = s.convRepo.GetParticipant(msg.ConversationID, userID)
	if err != nil {
		return nil, errors.New("access denied")
	}

	reaction := &models.Reaction{
		MessageID: msgID,
		UserID:    userID,
		Emoji:     emoji,
		CreatedAt: time.Now(),
	}

	if err := s.msgRepo.AddReaction(reaction); err != nil {
		return nil, err
	}

	userRepo := repositories.NewUserRepository()
	usr, err := userRepo.GetByID(userID)
	if err == nil && usr != nil {
		reaction.User = *usr
	}

	if s.OnReactionAdded != nil {
		s.OnReactionAdded(reaction)
	}

	return reaction, nil
}

func (s *MessageService) RemoveReaction(userID uuid.UUID, msgID uuid.UUID, emoji string) error {
	msg, err := s.msgRepo.GetByID(msgID)
	if err != nil {
		return err
	}

	// Verify participant
	_, err = s.convRepo.GetParticipant(msg.ConversationID, userID)
	if err != nil {
		return errors.New("access denied")
	}

	if err := s.msgRepo.RemoveReaction(msgID, userID, emoji); err != nil {
		return err
	}

	if s.OnReactionRemoved != nil {
		s.OnReactionRemoved(msgID, userID, emoji)
	}

	return nil
}
