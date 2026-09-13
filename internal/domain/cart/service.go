package cart

import (
	"errors"
	"math"
	"strings"

	"project-setup/internal/domain/discount"
	"project-setup/internal/models"
)

type CartService struct {
	repo            *CartRepository
	discountService *discount.DiscountService
}

func NewCartService(repo *CartRepository, discountService *discount.DiscountService) *CartService {
	return &CartService{
		repo:            repo,
		discountService: discountService,
	}
}

// GetCart returns the full cart summary with all discounts calculated
func (s *CartService) GetCart(userID uint) (*CartSummaryResponse, error) {
	cart, err := s.repo.GetOrCreateActiveCart(userID)
	if err != nil {
		return nil, errors.New("failed to load cart")
	}
	return s.buildCartSummary(cart), nil
}

// AddToCart adds a domain to the user's cart
func (s *CartService) AddToCart(userID uint, req AddToCartRequest) (*CartSummaryResponse, error) {
	domainName := strings.ToLower(strings.TrimSpace(req.DomainName))
	if domainName == "" {
		return nil, errors.New("domain name is required")
	}

	tld := extractTLD(domainName)
	if tld == "" {
		return nil, errors.New("invalid domain name")
	}

	if req.Currency == "" {
		req.Currency = "USD"
	}

	cart, err := s.repo.GetOrCreateActiveCart(userID)
	if err != nil {
		return nil, errors.New("failed to load cart")
	}

	// Check if domain already in cart
	existing, _ := s.repo.FindItemByDomain(cart.ID, domainName)
	if existing != nil {
		existing.Period = req.Period
		existing.UnitPrice = req.UnitPrice
		if err := s.repo.UpdateItem(existing); err != nil {
			return nil, errors.New("failed to update cart item")
		}
	} else {
		item := &models.CartItem{
			CartID:     cart.ID,
			DomainName: domainName,
			TLD:        tld,
			Period:     req.Period,
			UnitPrice:  req.UnitPrice,
			Currency:   req.Currency,
		}
		if err := s.repo.AddItem(item); err != nil {
			return nil, errors.New("failed to add item to cart")
		}
	}

	// Reload cart with updated items
	cart, _ = s.repo.GetCartWithItems(userID)
	return s.buildCartSummary(cart), nil
}

// UpdateItem updates the period of a cart item
func (s *CartService) UpdateItem(userID uint, itemID uint, req UpdateCartItemRequest) (*CartSummaryResponse, error) {
	cart, err := s.repo.GetCartWithItems(userID)
	if err != nil {
		return nil, errors.New("cart not found")
	}

	item, err := s.repo.FindItemByID(itemID)
	if err != nil {
		return nil, errors.New("item not found")
	}

	if item.CartID != cart.ID {
		return nil, errors.New("item does not belong to your cart")
	}

	item.Period = req.Period
	if err := s.repo.UpdateItem(item); err != nil {
		return nil, errors.New("failed to update item")
	}

	cart, _ = s.repo.GetCartWithItems(userID)
	return s.buildCartSummary(cart), nil
}

// RemoveItem removes an item from the cart
func (s *CartService) RemoveItem(userID uint, itemID uint) (*CartSummaryResponse, error) {
	cart, err := s.repo.GetCartWithItems(userID)
	if err != nil {
		return nil, errors.New("cart not found")
	}

	item, err := s.repo.FindItemByID(itemID)
	if err != nil {
		return nil, errors.New("item not found")
	}

	if item.CartID != cart.ID {
		return nil, errors.New("item does not belong to your cart")
	}

	if err := s.repo.DeleteItem(itemID, cart.ID); err != nil {
		return nil, errors.New("failed to remove item")
	}

	cart, _ = s.repo.GetCartWithItems(userID)
	return s.buildCartSummary(cart), nil
}

// ClearCart removes all items from the cart
func (s *CartService) ClearCart(userID uint) error {
	cart, err := s.repo.GetCartWithItems(userID)
	if err != nil {
		return errors.New("cart not found")
	}

	// Also remove any coupon
	s.repo.SetCoupon(cart.ID, nil)

	return s.repo.ClearCart(cart.ID)
}

// ApplyCoupon applies a coupon code to the cart
func (s *CartService) ApplyCoupon(userID uint, code string) (*CartSummaryResponse, error) {
	cart, err := s.repo.GetCartWithItems(userID)
	if err != nil {
		return nil, errors.New("cart not found")
	}

	if len(cart.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	// Calculate cart total before coupon for validation
	subtotalAfterTLD := s.calculateSubtotalAfterTLD(cart)

	_, err = s.discountService.ValidateCoupon(code, subtotalAfterTLD)
	if err != nil {
		return nil, err
	}

	upperCode := strings.ToUpper(strings.TrimSpace(code))
	if err := s.repo.SetCoupon(cart.ID, &upperCode); err != nil {
		return nil, errors.New("failed to apply coupon")
	}

	cart.CouponCode = &upperCode
	return s.buildCartSummary(cart), nil
}

// RemoveCoupon removes the coupon from the cart
func (s *CartService) RemoveCoupon(userID uint) (*CartSummaryResponse, error) {
	cart, err := s.repo.GetCartWithItems(userID)
	if err != nil {
		return nil, errors.New("cart not found")
	}

	if err := s.repo.SetCoupon(cart.ID, nil); err != nil {
		return nil, errors.New("failed to remove coupon")
	}

	cart.CouponCode = nil
	return s.buildCartSummary(cart), nil
}

// buildCartSummary calculates all prices and discounts for the cart
func (s *CartService) buildCartSummary(cart *models.Cart) *CartSummaryResponse {
	items := []CartItemResponse{}
	var subtotal float64
	var tldDiscountsTotal float64
	currency := "USD"

	for _, item := range cart.Items {
		originalPrice := item.UnitPrice * float64(item.Period)
		originalPrice = math.Round(originalPrice*100) / 100

		discountAmount, discountName := s.discountService.CalculateItemDiscount(item.TLD, originalPrice)
		finalPrice := math.Max(0, originalPrice-discountAmount)
		finalPrice = math.Round(finalPrice*100) / 100

		subtotal += originalPrice
		tldDiscountsTotal += discountAmount
		currency = item.Currency

		items = append(items, CartItemResponse{
			ID:             item.ID,
			DomainName:     item.DomainName,
			TLD:            item.TLD,
			Period:         item.Period,
			UnitPrice:      item.UnitPrice,
			OriginalPrice:  originalPrice,
			DiscountAmount: discountAmount,
			DiscountName:   discountName,
			FinalPrice:     finalPrice,
			Currency:       item.Currency,
		})
	}

	subtotal = math.Round(subtotal*100) / 100
	tldDiscountsTotal = math.Round(tldDiscountsTotal*100) / 100

	// Calculate coupon discount
	var couponDiscount float64
	cartBaseForCoupon := subtotal - tldDiscountsTotal

	if cart.CouponCode != nil && *cart.CouponCode != "" {
		couponObj, err := s.discountService.ValidateCoupon(*cart.CouponCode, cartBaseForCoupon)
		if err == nil {
			couponDiscount = s.discountService.CalculateCouponDiscount(couponObj, cartBaseForCoupon)
		}
	}

	couponDiscount = math.Round(couponDiscount*100) / 100
	totalDiscount := math.Round((tldDiscountsTotal+couponDiscount)*100) / 100
	totalAmount := math.Max(0, subtotal-totalDiscount)
	totalAmount = math.Round(totalAmount*100) / 100

	return &CartSummaryResponse{
		CartID:            cart.ID,
		Items:             items,
		ItemsCount:        len(items),
		Subtotal:          subtotal,
		TLDDiscountsTotal: tldDiscountsTotal,
		CouponCode:        cart.CouponCode,
		CouponDiscount:    couponDiscount,
		TotalDiscount:     totalDiscount,
		TotalAmount:       totalAmount,
		Currency:          currency,
	}
}

func (s *CartService) calculateSubtotalAfterTLD(cart *models.Cart) float64 {
	var total float64
	for _, item := range cart.Items {
		originalPrice := item.UnitPrice * float64(item.Period)
		discountAmount, _ := s.discountService.CalculateItemDiscount(item.TLD, originalPrice)
		total += originalPrice - discountAmount
	}
	return math.Round(total*100) / 100
}

// extractTLD extracts the TLD from a domain name
// eftakhar.com -> com
// eftakhar.co.uk -> co.uk
func extractTLD(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[1:], ".")
}
