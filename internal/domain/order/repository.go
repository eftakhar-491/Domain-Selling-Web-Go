package order

import (
	"strings"

	"project-setup/internal/models"

	"gorm.io/gorm"
)

// OrderRepository handles all database operations for orders
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new OrderRepository
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// CreateOrder persists a new order (with items) inside a transaction
func (r *OrderRepository) CreateOrder(order *models.Order) error {
	return r.db.Create(order).Error
}

// GetOrderByID fetches an order by primary key with items and user preloaded
func (r *OrderRepository) GetOrderByID(id uint) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Items").Preload("User").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrderByIDAndUser fetches an order only if it belongs to the given user
func (r *OrderRepository) GetOrderByIDAndUser(id uint, userID uint) (*models.Order, error) {
	var order models.Order
	err := r.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Items").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrderByOrderNumber looks up an order by its human-readable order number
func (r *OrderRepository) GetOrderByOrderNumber(orderNumber string) (*models.Order, error) {
	var order models.Order
	err := r.db.Where("order_number = ?", orderNumber).
		Preload("Items").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrderByPaymentIntentID finds the order associated with a Stripe PaymentIntent
func (r *OrderRepository) GetOrderByPaymentIntentID(intentID string) (*models.Order, error) {
	var order models.Order
	err := r.db.Where("stripe_payment_intent_id = ?", intentID).
		Preload("Items").
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetUserOrders returns a paginated list of orders for a specific user
func (r *OrderRepository) GetUserOrders(userID uint, page, limit int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	offset := (page - 1) * limit

	r.db.Model(&models.Order{}).Where("user_id = ?", userID).Count(&total)

	err := r.db.Where("user_id = ?", userID).
		Preload("Items").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders).Error

	return orders, total, err
}

// GetAllOrders returns a paginated list of all orders with User and Items preloaded, optionally filtered by status and search
func (r *OrderRepository) GetAllOrders(page, limit int, status string, search string) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{}).Joins("LEFT JOIN users ON users.id = orders.user_id")

	if strings.TrimSpace(status) != "" {
		st := strings.ToUpper(strings.TrimSpace(status))
		query = query.Where("orders.status = ? OR orders.payment_status = ?", st, st)
	}

	if strings.TrimSpace(search) != "" {
		s := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		query = query.Where("LOWER(orders.order_number) LIKE ? OR LOWER(users.name) LIKE ? OR LOWER(users.email) LIKE ?", s, s, s)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Preload("User").
		Preload("Items").
		Order("orders.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders).Error

	return orders, total, err
}


// UpdateOrder saves all changes to an existing order
func (r *OrderRepository) UpdateOrder(order *models.Order) error {
	return r.db.Save(order).Error
}

// UpdateOrderStatus atomically updates the order and payment status
func (r *OrderRepository) UpdateOrderStatus(orderID uint, status models.OrderStatus, paymentStatus models.PaymentStatus) error {
	return r.db.Model(&models.Order{}).Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"status":         status,
			"payment_status": paymentStatus,
		}).Error
}

// UpdateOrderItemsStatus updates all items in an order to the given status
func (r *OrderRepository) UpdateOrderItemsStatus(orderID uint, status models.OrderItemStatus) error {
	return r.db.Model(&models.OrderItem{}).Where("order_id = ?", orderID).
		Update("status", status).Error
}
