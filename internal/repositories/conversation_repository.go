package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"jia/server/internal/database"
	"jia/server/internal/models"
)

type ConversationRepository struct{}

func NewConversationRepository() *ConversationRepository {
	return &ConversationRepository{}
}

func (r *ConversationRepository) Create(conv *models.Conversation, creatorID uuid.UUID, participantIDs []uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Create the conversation
		if err := tx.Create(conv).Error; err != nil {
			return err
		}

		// Add participants
		now := time.Now()
		participants := make([]models.ConversationParticipant, len(participantIDs))
		for i, pID := range participantIDs {
			role := "member"
			if pID == creatorID {
				role = "owner"
			}
			participants[i] = models.ConversationParticipant{
				ConversationID: conv.ID,
				UserID:         pID,
				Role:           role,
				JoinedAt:       now,
			}
		}

		if err := tx.Create(&participants).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *ConversationRepository) GetByID(id uuid.UUID) (*models.Conversation, error) {
	var conv models.Conversation
	err := database.DB.Preload("Participants.User").Where("id = ?", id).First(&conv).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *ConversationRepository) ListByUserID(userID uuid.UUID) ([]models.Conversation, error) {
	var participants []models.ConversationParticipant
	// Find all conversation IDs this user participates in
	err := database.DB.Where("user_id = ? AND left_at IS NULL", userID).Find(&participants).Error
	if err != nil {
		return nil, err
	}

	if len(participants) == 0 {
		return []models.Conversation{}, nil
	}

	convIDs := make([]uuid.UUID, len(participants))
	for i, p := range participants {
		convIDs[i] = p.ConversationID
	}

	var conversations []models.Conversation
	err = database.DB.Preload("Participants.User").Where("id IN ?", convIDs).Order("updated_at DESC").Find(&conversations).Error
	if err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *ConversationRepository) AddParticipants(convID uuid.UUID, participantIDs []uuid.UUID) error {
	now := time.Now()
	var participants []models.ConversationParticipant
	for _, pID := range participantIDs {
		participants = append(participants, models.ConversationParticipant{
			ConversationID: convID,
			UserID:         pID,
			Role:           "member",
			JoinedAt:       now,
		})
	}
	return database.DB.Create(&participants).Error
}

func (r *ConversationRepository) RemoveParticipant(convID uuid.UUID, userID uuid.UUID) error {
	return database.DB.Model(&models.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", convID, userID).
		Update("left_at", time.Now()).Error
}

func (r *ConversationRepository) UpdateParticipantReadCursor(convID uuid.UUID, userID uuid.UUID, messageID uuid.UUID) error {
	return database.DB.Model(&models.ConversationParticipant{}).
		Where("conversation_id = ? AND user_id = ?", convID, userID).
		Update("last_read_message_id", messageID).Error
}

func (r *ConversationRepository) GetParticipant(convID uuid.UUID, userID uuid.UUID) (*models.ConversationParticipant, error) {
	var cp models.ConversationParticipant
	err := database.DB.Where("conversation_id = ? AND user_id = ? AND left_at IS NULL", convID, userID).First(&cp).Error
	if err != nil {
		return nil, err
	}
	return &cp, nil
}

func (r *ConversationRepository) FindDM(userA, userB uuid.UUID) (*models.Conversation, error) {
	var participantA []models.ConversationParticipant
	err := database.DB.Where("user_id = ? AND left_at IS NULL", userA).Find(&participantA).Error
	if err != nil {
		return nil, err
	}

	if len(participantA) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	convIDs := make([]uuid.UUID, len(participantA))
	for i, p := range participantA {
		convIDs[i] = p.ConversationID
	}

	var conv models.Conversation
	err = database.DB.Preload("Participants.User").
		Joins("JOIN conversation_participants cp ON cp.conversation_id = conversations.id").
		Where("conversations.id IN ? AND conversations.type = 'dm' AND cp.user_id = ? AND cp.left_at IS NULL", convIDs, userB).
		First(&conv).Error

	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (r *ConversationRepository) Delete(id uuid.UUID) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Soft delete conversation messages
		if err := tx.Where("conversation_id = ?", id).Delete(&models.Message{}).Error; err != nil {
			return err
		}
		// Hard delete participants (or mark left_at)
		if err := tx.Where("conversation_id = ?", id).Delete(&models.ConversationParticipant{}).Error; err != nil {
			return err
		}
		// Delete the conversation
		if err := tx.Delete(&models.Conversation{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}
