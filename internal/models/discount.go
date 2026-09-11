package models

import (
	"time"

	"gorm.io/gorm"
)

type DiscountType string

const (
	DiscountTypePercentage DiscountType = "PERCENTAGE"
	DiscountTypeFixed      DiscountType = "FIXED_AMOUNT"
)

type DiscountScope string

const (
	DiscountScopeTLD    DiscountScope = "TLD"
	DiscountScopeCoupon DiscountScope = "COUPON"
)

type Discount struct {
	gorm.Model
	Code              *string       `json:"code,omitempty" gorm:"type:varchar(50);uniqueIndex"`
	Name              string        `json:"name" gorm:"type:varchar(100);not null"`
	Description       string        `json:"description" gorm:"type:text"`
	Type              DiscountType  `json:"type" gorm:"type:varchar(20);not null"`
	Scope             DiscountScope `json:"scope" gorm:"type:varchar(20);not null"`
	TargetTLD         *string       `json:"target_tld,omitempty" gorm:"type:varchar(20);index"`
	Value             float64       `json:"value" gorm:"type:decimal(10,2);not null"`
	MinSpend          float64       `json:"min_spend" gorm:"type:decimal(10,2);default:0"`
	MaxDiscountAmount *float64      `json:"max_discount_amount,omitempty" gorm:"type:decimal(10,2)"`
	UsageLimit        int           `json:"usage_limit" gorm:"default:0"`
	UsedCount         int           `json:"used_count" gorm:"default:0"`
	StartDate         *time.Time    `json:"start_date,omitempty"`
	EndDate           *time.Time    `json:"end_date,omitempty"`
	IsActive          bool          `json:"is_active" gorm:"default:true;index"`
}
