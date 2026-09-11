package models

import "gorm.io/gorm"

type CartStatus string

const (
	CartStatusActive    CartStatus = "ACTIVE"
	CartStatusConverted CartStatus = "CONVERTED"
)

type Cart struct {
	gorm.Model
	UserID     uint       `json:"user_id" gorm:"uniqueIndex;not null"`
	User       User       `json:"-" gorm:"foreignKey:UserID"`
	CouponCode *string    `json:"coupon_code,omitempty" gorm:"type:varchar(50)"`
	Status     CartStatus `json:"status" gorm:"type:varchar(20);default:'ACTIVE';not null"`
	Items      []CartItem `json:"items" gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE"`
}

type CartItem struct {
	gorm.Model
	CartID     uint    `json:"cart_id" gorm:"not null;index"`
	DomainName string  `json:"domain_name" gorm:"type:varchar(255);not null"`
	TLD        string  `json:"tld" gorm:"type:varchar(20);not null"`
	Period     int     `json:"period" gorm:"default:1;not null"`
	UnitPrice  float64 `json:"unit_price" gorm:"type:decimal(10,2);not null"`
	Currency   string  `json:"currency" gorm:"type:varchar(10);default:'USD';not null"`
}
