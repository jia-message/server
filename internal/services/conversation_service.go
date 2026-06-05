package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"jia/server/internal/models"
	"jia/server/internal/repositories"
)

type ConversationService struct {
	convRepo *repositories.ConversationRepository
	userRepo *repositories.UserRepository
}

func NewConversationService(
	convRepo *repositories.ConversationRepository,
	userRepo *repositories.UserRepository,
) *ConversationService {
	return &ConversationService{
		convRepo: convRepo,
		userRepo: userRepo,
	}
}

func (s *ConversationService) CreateConversation(creatorID uuid.UUID, convType string, name *string, participantIDs []uuid.UUID) (*models.Conversation, error) {
	if convType != "dm" && convType != "group" {
		return nil, errors.New("invalid conversation type")
	}

	// For DM, check if it already exists between the two users
	if convType == "dm" {
		if len(participantIDs) != 2 {
			return nil, errors.New("a direct message requires exactly two participants")
		}
		var otherID uuid.UUID
		if participantIDs[0] == creatorID {
			otherID = participantIDs[1]
		} else {
			otherID = participantIDs[0]
		}

		existing, err := s.convRepo.FindDM(creatorID, otherID)
		if err == nil {
			return existing, nil
		}
	}

	// Create conversation
	conv := &models.Conversation{
		Type:      convType,
		Name:      name,
		CreatedBy: creatorID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.convRepo.Create(conv, creatorID, participantIDs)
	if err != nil {
		return nil, err
	}

	// Fetch conversation with participants preloaded
	return s.convRepo.GetByID(conv.ID)
}

func (s *ConversationService) GetConversationDetails(userID, convID uuid.UUID) (*models.Conversation, error) {
	// Verify user is a participant
	_, err := s.convRepo.GetParticipant(convID, userID)
	if err != nil {
		return nil, errors.New("access denied: you are not a participant in this conversation")
	}

	return s.convRepo.GetByID(convID)
}

func (s *ConversationService) ListConversations(userID uuid.UUID) ([]models.Conversation, error) {
	return s.convRepo.ListByUserID(userID)
}

func (s *ConversationService) AddParticipants(userID, convID uuid.UUID, participantIDs []uuid.UUID) error {
	// Only admins or owner can add participants to groups
	cp, err := s.convRepo.GetParticipant(convID, userID)
	if err != nil {
		return errors.New("access denied")
	}

	if cp.Role != "owner" && cp.Role != "admin" {
		return errors.New("unauthorized to add participants")
	}

	return s.convRepo.AddParticipants(convID, participantIDs)
}

func (s *ConversationService) RemoveParticipant(userID, convID, targetUserID uuid.UUID) error {
	cp, err := s.convRepo.GetParticipant(convID, userID)
	if err != nil {
		return errors.New("access denied")
	}

	if cp.Role != "owner" && cp.Role != "admin" {
		return errors.New("unauthorized to remove participants")
	}

	return s.convRepo.RemoveParticipant(convID, targetUserID)
}

func (s *ConversationService) MarkAsRead(userID, convID uuid.UUID, messageID uuid.UUID) error {
	_, err := s.convRepo.GetParticipant(convID, userID)
	if err != nil {
		return errors.New("access denied")
	}

	return s.convRepo.UpdateParticipantReadCursor(convID, userID, messageID)
}

func (s *ConversationService) LeaveConversation(userID, convID uuid.UUID) error {
	_, err := s.convRepo.GetParticipant(convID, userID)
	if err != nil {
		return errors.New("not a participant in this conversation")
	}

	return s.convRepo.RemoveParticipant(convID, userID)
}
