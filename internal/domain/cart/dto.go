package cart

type AddToCartRequest struct {
	DomainName string  `json:"domainName" validate:"required"`
	Period     int     `json:"period" validate:"required,min=1,max=10"`
	UnitPrice  float64 `json:"unitPrice" validate:"required,gt=0"`
	Currency   string  `json:"currency"`
}

type UpdateCartItemRequest struct {
	Period int `json:"period" validate:"required,min=1,max=10"`
}

type ApplyCouponRequest struct {
	Code string `json:"code" validate:"required"`
}

type CartItemResponse struct {
	ID             uint    `json:"id"`
	DomainName     string  `json:"domain_name"`
	TLD            string  `json:"tld"`
	Period         int     `json:"period"`
	UnitPrice      float64 `json:"unit_price"`
	OriginalPrice  float64 `json:"original_price"`
	DiscountAmount float64 `json:"discount_amount"`
	DiscountName   string  `json:"discount_name,omitempty"`
	FinalPrice     float64 `json:"final_price"`
	Currency       string  `json:"currency"`
}

type CartSummaryResponse struct {
	CartID            uint               `json:"cart_id"`
	Items             []CartItemResponse `json:"items"`
	ItemsCount        int                `json:"items_count"`
	Subtotal          float64            `json:"subtotal"`
	TLDDiscountsTotal float64            `json:"tld_discounts_total"`
	CouponCode        *string            `json:"coupon_code,omitempty"`
	CouponDiscount    float64            `json:"coupon_discount"`
	TotalDiscount     float64            `json:"total_discount"`
	TotalAmount       float64            `json:"total_amount"`
	Currency          string             `json:"currency"`
}
