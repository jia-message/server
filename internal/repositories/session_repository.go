package repositories

import (
	"time"

	"github.com/google/uuid"
	"jia/server/internal/database"
	"jia/server/internal/models"
)

type SessionRepository struct{}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{}
}

func (r *SessionRepository) Create(session *models.Session) error {
	return database.DB.Create(session).Error
}

func (r *SessionRepository) GetByRefreshToken(tokenHash string) (*models.Session, error) {
	var session models.Session
	err := database.DB.Where("refresh_token_hash = ? AND expires_at > ?", tokenHash, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *SessionRepository) Delete(sessionID uuid.UUID) error {
	return database.DB.Delete(&models.Session{}, sessionID).Error
}

func (r *SessionRepository) DeleteAllByUserID(userID uuid.UUID) error {
	return database.DB.Where("user_id = ?", userID).Delete(&models.Session{}).Error
}

func (r *SessionRepository) CleanupExpired() error {
	return database.DB.Where("expires_at <= ?", time.Now()).Delete(&models.Session{}).Error
}
