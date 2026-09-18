package contact

import (
	"project-setup/internal/models"

	"gorm.io/gorm"
)

type Repository interface {
	Create(msg *models.ContactMessage) error
	FindAll(page, limit int) ([]models.ContactMessage, int64, error)
	FindByID(id uint) (*models.ContactMessage, error)
	UpdateStatus(id uint, status string) error
	GetSystemSetting() (*models.SystemSetting, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(msg *models.ContactMessage) error {
	return r.db.Create(msg).Error
}

func (r *repository) FindAll(page, limit int) ([]models.ContactMessage, int64, error) {
	var messages []models.ContactMessage
	var total int64

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	r.db.Model(&models.ContactMessage{}).Count(&total)

	err := r.db.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&messages).Error

	return messages, total, err
}

func (r *repository) FindByID(id uint) (*models.ContactMessage, error) {
	var msg models.ContactMessage
	err := r.db.First(&msg, id).Error
	return &msg, err
}

func (r *repository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.ContactMessage{}).Where("id = ?", id).Update("status", status).Error
}

func (r *repository) GetSystemSetting() (*models.SystemSetting, error) {
	var setting models.SystemSetting
	err := r.db.First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}
