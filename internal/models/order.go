package models

import "gorm.io/gorm"

// OrderStatus represents the lifecycle status of an order
type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "PENDING_PAYMENT"
	OrderStatusPaid           OrderStatus = "PAID"
	OrderStatusProcessing     OrderStatus = "PROCESSING"
	OrderStatusCompleted      OrderStatus = "COMPLETED"
	OrderStatusCancelled      OrderStatus = "CANCELLED"
	OrderStatusFailed         OrderStatus = "FAILED"
)

// PaymentStatus represents the payment state
type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "PENDING"
	PaymentStatusPaid     PaymentStatus = "PAID"
	PaymentStatusFailed   PaymentStatus = "FAILED"
	PaymentStatusRefunded PaymentStatus = "REFUNDED"
)

// OrderItemStatus represents individual item status within an order
type OrderItemStatus string

const (
	OrderItemStatusPending OrderItemStatus = "PENDING"
	OrderItemStatusActive  OrderItemStatus = "ACTIVE"
	OrderItemStatusFailed  OrderItemStatus = "FAILED"
)

// Order represents a domain purchase order
type Order struct {
	gorm.Model
	OrderNumber           string        `json:"order_number" gorm:"type:varchar(30);uniqueIndex;not null"`
	UserID                uint          `json:"user_id" gorm:"index;not null"`
	User                  User          `json:"-" gorm:"foreignKey:UserID"`
	Subtotal              float64       `json:"subtotal" gorm:"type:decimal(10,2);not null"`
	DiscountAmount        float64       `json:"discount_amount" gorm:"type:decimal(10,2);default:0"`
	CouponCode            *string       `json:"coupon_code,omitempty" gorm:"type:varchar(50)"`
	CouponDiscount        float64       `json:"coupon_discount" gorm:"type:decimal(10,2);default:0"`
	TotalAmount           float64       `json:"total_amount" gorm:"type:decimal(10,2);not null"`
	Currency              string        `json:"currency" gorm:"type:varchar(10);default:'USD';not null"`
	Status                OrderStatus   `json:"status" gorm:"type:varchar(30);default:'PENDING_PAYMENT';not null;index"`
	PaymentStatus         PaymentStatus `json:"payment_status" gorm:"type:varchar(20);default:'PENDING';not null"`
	PaymentMethod         string        `json:"payment_method" gorm:"type:varchar(20);default:'STRIPE';not null"`
	StripePaymentIntentID string        `json:"stripe_payment_intent_id" gorm:"type:varchar(255);index"`
	StripeClientSecret    string        `json:"-" gorm:"type:varchar(255)"`
	Items                 []OrderItem   `json:"items" gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
}

// OrderItem represents a single domain within an order
type OrderItem struct {
	gorm.Model
	OrderID        uint            `json:"order_id" gorm:"not null;index"`
	DomainName     string          `json:"domain_name" gorm:"type:varchar(255);not null"`
	TLD            string          `json:"tld" gorm:"type:varchar(20);not null"`
	Period         int             `json:"period" gorm:"default:1;not null"`
	UnitPrice      float64         `json:"unit_price" gorm:"type:decimal(10,2);not null"`
	DiscountAmount float64         `json:"discount_amount" gorm:"type:decimal(10,2);default:0"`
	FinalPrice     float64         `json:"final_price" gorm:"type:decimal(10,2);not null"`
	Currency       string          `json:"currency" gorm:"type:varchar(10);default:'USD';not null"`
	Status         OrderItemStatus `json:"status" gorm:"type:varchar(20);default:'PENDING';not null"`
}
