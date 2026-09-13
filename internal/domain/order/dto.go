package order

// ---- Request DTOs ----

// CreateOrderFromCartRequest is the payload for checking out from the user's active cart
type CreateOrderFromCartRequest struct {
	// Currency override (optional, defaults to cart items' currency)
	Currency      string `json:"currency"`
	PaymentMethod string `json:"payment_method"`
}

// DirectOrderRequest allows placing an order for domains without using a cart
type DirectOrderRequest struct {
	Items    []DirectOrderItemRequest `json:"items" validate:"required,min=1,dive"`
	Currency string                   `json:"currency"`
}

// DirectOrderItemRequest represents a single domain in a direct order
type DirectOrderItemRequest struct {
	DomainName string  `json:"domainName" validate:"required"`
	Period     int     `json:"period" validate:"required,min=1,max=10"`
	UnitPrice  float64 `json:"unitPrice" validate:"required,gt=0"`
}

// ConfirmPaymentRequest is the payload to verify payment from the client side
type ConfirmPaymentRequest struct {
	PaymentIntentID string `json:"payment_intent_id" validate:"required"`
}

// ---- Response DTOs ----

// OrderItemResponse represents a single domain in an order response
type OrderItemResponse struct {
	ID             uint    `json:"id"`
	DomainName     string  `json:"domain_name"`
	TLD            string  `json:"tld"`
	Period         int     `json:"period"`
	UnitPrice      float64 `json:"unit_price"`
	DiscountAmount float64 `json:"discount_amount"`
	FinalPrice     float64 `json:"final_price"`
	Currency       string  `json:"currency"`
	Status         string  `json:"status"`
}

// OrderResponse is the full order detail returned to the client
type OrderResponse struct {
	ID                    uint                `json:"id"`
	OrderNumber           string              `json:"order_number"`
	Items                 []OrderItemResponse `json:"items"`
	ItemsCount            int                 `json:"items_count"`
	Subtotal              float64             `json:"subtotal"`
	DiscountAmount        float64             `json:"discount_amount"`
	CouponCode            *string             `json:"coupon_code,omitempty"`
	CouponDiscount        float64             `json:"coupon_discount"`
	TotalAmount           float64             `json:"total_amount"`
	Currency              string              `json:"currency"`
	Status                string              `json:"status"`
	PaymentStatus         string              `json:"payment_status"`
	PaymentMethod         string              `json:"payment_method"`
	StripeClientSecret    string              `json:"stripe_client_secret,omitempty"`
	StripePaymentIntentID string              `json:"stripe_payment_intent_id,omitempty"`
	CreatedAt             string              `json:"created_at"`
}

// OrderListResponse is a paginated list of orders
type OrderListResponse struct {
	Orders     []OrderResponse `json:"orders"`
	TotalCount int64           `json:"total_count"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}
