package discount

import "time"

type CreateDiscountRequest struct {
	Code              *string    `json:"code,omitempty"`
	Name              string     `json:"name" validate:"required"`
	Description       string     `json:"description"`
	Type              string     `json:"type" validate:"required,oneof=PERCENTAGE FIXED_AMOUNT"`
	Scope             string     `json:"scope" validate:"required,oneof=TLD COUPON"`
	TargetTLD         *string    `json:"target_tld,omitempty"`
	Value             float64    `json:"value" validate:"required,gt=0"`
	MinSpend          float64    `json:"min_spend"`
	MaxDiscountAmount *float64   `json:"max_discount_amount,omitempty"`
	UsageLimit        int        `json:"usage_limit"`
	StartDate         *time.Time `json:"start_date,omitempty"`
	EndDate           *time.Time `json:"end_date,omitempty"`
}

type UpdateDiscountRequest struct {
	Name              *string    `json:"name,omitempty"`
	Description       *string    `json:"description,omitempty"`
	Value             *float64   `json:"value,omitempty"`
	MinSpend          *float64   `json:"min_spend,omitempty"`
	MaxDiscountAmount *float64   `json:"max_discount_amount,omitempty"`
	UsageLimit        *int       `json:"usage_limit,omitempty"`
	StartDate         *time.Time `json:"start_date,omitempty"`
	EndDate           *time.Time `json:"end_date,omitempty"`
	IsActive          *bool      `json:"is_active,omitempty"`
}
