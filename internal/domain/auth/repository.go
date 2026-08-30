package auth

import (
	"project-setup/internal/models"

	"gorm.io/gorm"
)

// AuthRepository handles database operations for authentication
type AuthRepository struct {
	db *gorm.DB
}

// NewAuthRepository creates a new AuthRepository instance
func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

// CreateUser inserts a new user into the database
func (r *AuthRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByEmail finds a user by their email address
func (r *AuthRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates an existing user's details (e.g. password)
func (r *AuthRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}
