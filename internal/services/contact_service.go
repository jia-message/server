package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"jia/server/internal/models"
	"jia/server/internal/repositories"
)

type ContactService struct {
	contactRepo *repositories.ContactRepository
	userRepo    *repositories.UserRepository
}

func NewContactService(
	contactRepo *repositories.ContactRepository,
	userRepo *repositories.UserRepository,
) *ContactService {
	return &ContactService{
		contactRepo: contactRepo,
		userRepo:    userRepo,
	}
}

func (s *ContactService) ListContacts(userID uuid.UUID) ([]models.Contact, error) {
	return s.contactRepo.List(userID)
}

func (s *ContactService) AddContact(userID uuid.UUID, contactUsername string, nickname *string) (*models.Contact, error) {
	contactUser, err := s.userRepo.GetByUsername(contactUsername)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if contactUser.ID == userID {
		return nil, errors.New("you cannot add yourself as a contact")
	}

	contact := &models.Contact{
		UserID:        userID,
		ContactUserID: contactUser.ID,
		Nickname:      nickname,
		IsBlocked:     false,
		CreatedAt:     time.Now(),
	}

	if err := s.contactRepo.Add(contact); err != nil {
		return nil, err
	}

	// Fetch contact with user details preloaded
	return s.contactRepo.Get(userID, contactUser.ID)
}

func (s *ContactService) RemoveContact(userID, contactUserID uuid.UUID) error {
	return s.contactRepo.Remove(userID, contactUserID)
}

func (s *ContactService) UpdateContact(userID, contactUserID uuid.UUID, nickname *string, isBlocked bool) error {
	return s.contactRepo.Update(userID, contactUserID, nickname, isBlocked)
}
