package repositories

import (
	"github.com/google/uuid"
	"jia/server/internal/database"
	"jia/server/internal/models"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(user *models.User) error {
	return database.DB.Create(user).Error
}

func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := database.DB.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := database.DB.Where("LOWER(username) = LOWER(?)", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := database.DB.Where("LOWER(email) = LOWER(?)", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	return database.DB.Save(user).Error
}

func (r *UserRepository) Search(query string) ([]models.User, error) {
	var users []models.User
	err := database.DB.Where("LOWER(username) LIKE LOWER(?) OR LOWER(display_name) LIKE LOWER(?)", "%"+query+"%", "%"+query+"%").
		Limit(20).Find(&users).Error
	return users, err
}

func (r *UserRepository) Count() (int64, error) {
	var count int64
	err := database.DB.Model(&models.User{}).Count(&count).Error
	return count, err
}

func (r *UserRepository) List(page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	err := database.DB.Model(&models.User{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = database.DB.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) Delete(id uuid.UUID) error {
	return database.DB.Delete(&models.User{}, id).Error
}
