package repositories

import (
	"github.com/google/uuid"
	"jia/server/internal/database"
	"jia/server/internal/models"
)

type ContactRepository struct{}

func NewContactRepository() *ContactRepository {
	return &ContactRepository{}
}

func (r *ContactRepository) List(userID uuid.UUID) ([]models.Contact, error) {
	var contacts []models.Contact
	err := database.DB.Preload("ContactUser").Where("user_id = ?", userID).Find(&contacts).Error
	return contacts, err
}

func (r *ContactRepository) Add(contact *models.Contact) error {
	var existing models.Contact
	err := database.DB.Where("user_id = ? AND contact_user_id = ?", contact.UserID, contact.ContactUserID).First(&existing).Error
	if err == nil {
		contact.ID = existing.ID
		return database.DB.Save(contact).Error
	}
	return database.DB.Create(contact).Error
}

func (r *ContactRepository) Remove(userID, contactUserID uuid.UUID) error {
	return database.DB.Where("user_id = ? AND contact_user_id = ?", userID, contactUserID).
		Delete(&models.Contact{}).Error
}

func (r *ContactRepository) Update(userID, contactUserID uuid.UUID, nickname *string, isBlocked bool) error {
	updates := map[string]interface{}{
		"is_blocked": isBlocked,
	}
	if nickname != nil {
		updates["nickname"] = *nickname
	} else {
		updates["nickname"] = nil
	}

	return database.DB.Model(&models.Contact{}).
		Where("user_id = ? AND contact_user_id = ?", userID, contactUserID).
		Updates(updates).Error
}

func (r *ContactRepository) Get(userID, contactUserID uuid.UUID) (*models.Contact, error) {
	var contact models.Contact
	err := database.DB.Where("user_id = ? AND contact_user_id = ?", userID, contactUserID).First(&contact).Error
	if err != nil {
		return nil, err
	}
	return &contact, nil
}
