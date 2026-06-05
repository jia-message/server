package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"jia/server/internal/database"
	"jia/server/internal/models"
)

type MessageRepository struct{}

func NewMessageRepository() *MessageRepository {
	return &MessageRepository{}
}

func (r *MessageRepository) Create(msg *models.Message) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(msg).Error; err != nil {
			return err
		}
		// Update the updated_at timestamp of the parent conversation
		return tx.Model(&models.Conversation{}).
			Where("id = ?", msg.ConversationID).
			Update("updated_at", time.Now()).Error
	})
}

func (r *MessageRepository) GetByID(id uuid.UUID) (*models.Message, error) {
	var msg models.Message
	err := database.DB.Preload("Sender").Preload("Attachments").Preload("Reactions.User").Where("id = ?", id).First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *MessageRepository) GetHistory(convID uuid.UUID, before *time.Time, limit int) ([]models.Message, error) {
	var messages []models.Message
	query := database.DB.Preload("Sender").Preload("Attachments").Preload("Reactions.User").
		Where("conversation_id = ?", convID)

	if before != nil {
		query = query.Where("created_at < ?", before)
	}

	err := query.Order("created_at DESC").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// Reverse to chronological order (newest at bottom)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *MessageRepository) Update(msg *models.Message) error {
	return database.DB.Save(msg).Error
}

func (r *MessageRepository) Delete(id uuid.UUID) error {
	return database.DB.Delete(&models.Message{}, id).Error
}

func (r *MessageRepository) AddReaction(reaction *models.Reaction) error {
	return database.DB.Create(reaction).Error
}

func (r *MessageRepository) RemoveReaction(msgID uuid.UUID, userID uuid.UUID, emoji string) error {
	return database.DB.Where("message_id = ? AND user_id = ? AND emoji = ?", msgID, userID, emoji).
		Delete(&models.Reaction{}).Error
}

func (r *MessageRepository) SaveAttachment(att *models.Attachment) error {
	return database.DB.Create(att).Error
}

func (r *MessageRepository) GetAttachmentByID(id uuid.UUID) (*models.Attachment, error) {
	var att models.Attachment
	err := database.DB.Where("id = ?", id).First(&att).Error
	if err != nil {
		return nil, err
	}
	return &att, nil
}

func (r *MessageRepository) Count() (int64, error) {
	var count int64
	err := database.DB.Model(&models.Message{}).Count(&count).Error
	return count, err
}
