package cart

import (
	"project-setup/internal/models"

	"gorm.io/gorm"
)

type CartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{db: db}
}

// GetOrCreateActiveCart finds the user's active cart or creates one
func (r *CartRepository) GetOrCreateActiveCart(userID uint) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.Where("user_id = ? AND status = ?", userID, models.CartStatusActive).
		Preload("Items").
		First(&cart).Error

	if err == gorm.ErrRecordNotFound {
		cart = models.Cart{
			UserID: userID,
			Status: models.CartStatusActive,
		}
		if err := r.db.Create(&cart).Error; err != nil {
			return nil, err
		}
		return &cart, nil
	}

	return &cart, err
}

// GetCartWithItems fetches the user's active cart with all items
func (r *CartRepository) GetCartWithItems(userID uint) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.Where("user_id = ? AND status = ?", userID, models.CartStatusActive).
		Preload("Items").
		First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

// FindItemByDomain checks if a domain is already in the cart
func (r *CartRepository) FindItemByDomain(cartID uint, domainName string) (*models.CartItem, error) {
	var item models.CartItem
	err := r.db.Where("cart_id = ? AND domain_name = ?", cartID, domainName).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// FindItemByID finds a cart item by ID
func (r *CartRepository) FindItemByID(itemID uint) (*models.CartItem, error) {
	var item models.CartItem
	if err := r.db.First(&item, itemID).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// AddItem adds a new item to the cart
func (r *CartRepository) AddItem(item *models.CartItem) error {
	return r.db.Create(item).Error
}

// UpdateItem saves changes to a cart item
func (r *CartRepository) UpdateItem(item *models.CartItem) error {
	return r.db.Save(item).Error
}

// DeleteItem removes a cart item
func (r *CartRepository) DeleteItem(itemID uint, cartID uint) error {
	return r.db.Where("id = ? AND cart_id = ?", itemID, cartID).Delete(&models.CartItem{}).Error
}

// ClearCart removes all items from a cart
func (r *CartRepository) ClearCart(cartID uint) error {
	return r.db.Where("cart_id = ?", cartID).Delete(&models.CartItem{}).Error
}

// SetCoupon sets or removes the coupon code on a cart
func (r *CartRepository) SetCoupon(cartID uint, couponCode *string) error {
	return r.db.Model(&models.Cart{}).Where("id = ?", cartID).
		Update("coupon_code", couponCode).Error
}
