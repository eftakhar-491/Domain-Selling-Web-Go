package discount

import (
	"project-setup/internal/models"

	"gorm.io/gorm"
)

type DiscountRepository struct {
	db *gorm.DB
}

func NewDiscountRepository(db *gorm.DB) *DiscountRepository {
	return &DiscountRepository{db: db}
}

func (r *DiscountRepository) Create(discount *models.Discount) error {
	return r.db.Create(discount).Error
}

func (r *DiscountRepository) FindByID(id uint) (*models.Discount, error) {
	var discount models.Discount
	if err := r.db.First(&discount, id).Error; err != nil {
		return nil, err
	}
	return &discount, nil
}

func (r *DiscountRepository) FindByCode(code string) (*models.Discount, error) {
	var discount models.Discount
	if err := r.db.Where("code = ?", code).First(&discount).Error; err != nil {
		return nil, err
	}
	return &discount, nil
}

func (r *DiscountRepository) FindActiveByTLD(tld string) (*models.Discount, error) {
	var discount models.Discount
	err := r.db.Where(
		"scope = ? AND target_tld = ? AND is_active = ? AND (start_date IS NULL OR start_date <= NOW()) AND (end_date IS NULL OR end_date >= NOW())",
		models.DiscountScopeTLD, tld, true,
	).First(&discount).Error
	if err != nil {
		return nil, err
	}
	return &discount, nil
}

func (r *DiscountRepository) FindAllActiveTLD() ([]models.Discount, error) {
	var discounts []models.Discount
	err := r.db.Where(
		"scope = ? AND is_active = ? AND (start_date IS NULL OR start_date <= NOW()) AND (end_date IS NULL OR end_date >= NOW())",
		models.DiscountScopeTLD, true,
	).Find(&discounts).Error
	return discounts, err
}

func (r *DiscountRepository) FindAll(page, limit int) ([]models.Discount, int64, error) {
	var discounts []models.Discount
	var total int64

	if err := r.db.Model(&models.Discount{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&discounts).Error; err != nil {
		return nil, 0, err
	}

	return discounts, total, nil
}

func (r *DiscountRepository) Update(discount *models.Discount) error {
	return r.db.Save(discount).Error
}

func (r *DiscountRepository) Delete(id uint) error {
	return r.db.Delete(&models.Discount{}, id).Error
}

func (r *DiscountRepository) IncrementUsage(id uint) error {
	return r.db.Model(&models.Discount{}).Where("id = ?", id).
		UpdateColumn("used_count", gorm.Expr("used_count + 1")).Error
}
