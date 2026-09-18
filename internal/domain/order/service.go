package order

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"project-setup/internal/domain/discount"
	"project-setup/internal/models"
	"project-setup/internal/pkg/mail"
	stripepkg "project-setup/internal/pkg/stripe"

	stripe "github.com/stripe/stripe-go/v78"
	"gorm.io/gorm"
)

// OrderService contains all business logic for creating and managing orders
type OrderService struct {
	repo            *OrderRepository
	cartDB          *gorm.DB // used for cart operations during checkout
	discountService *discount.DiscountService
	stripeService   *stripepkg.StripeService
	resellerService *ResellerService
}

// NewOrderService creates a new OrderService
func NewOrderService(
	repo *OrderRepository,
	db *gorm.DB,
	discountService *discount.DiscountService,
	stripeService *stripepkg.StripeService,
) *OrderService {
	return &OrderService{
		repo:            repo,
		cartDB:          db,
		discountService: discountService,
		stripeService:   stripeService,
		resellerService: NewResellerService(db),
	}
}

// ============================================================
// CHECKOUT FROM CART
// ============================================================

// CheckoutFromCart converts the user's active cart into a paid order via Stripe
func (s *OrderService) CheckoutFromCart(userID uint, req CreateOrderFromCartRequest) (*OrderResponse, error) {
	// 1. Fetch the user's active cart
	var cart models.Cart
	err := s.cartDB.Where("user_id = ? AND status = ?", userID, models.CartStatusActive).
		Preload("Items").
		First(&cart).Error
	if err != nil {
		return nil, errors.New("no active cart found")
	}

	if len(cart.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	// 2. Calculate pricing with discounts
	currency := "USD"
	if req.Currency != "" {
		currency = req.Currency
	}

	var subtotal float64
	var totalTLDDiscount float64
	var orderItems []models.OrderItem

	for _, item := range cart.Items {
		originalPrice := math.Round(item.UnitPrice*float64(item.Period)*100) / 100
		discountAmt, _ := s.discountService.CalculateItemDiscount(item.TLD, originalPrice)
		finalPrice := math.Max(0, originalPrice-discountAmt)
		finalPrice = math.Round(finalPrice*100) / 100

		subtotal += originalPrice
		totalTLDDiscount += discountAmt

		orderItems = append(orderItems, models.OrderItem{
			DomainName:     item.DomainName,
			TLD:            item.TLD,
			Period:         item.Period,
			UnitPrice:      item.UnitPrice,
			DiscountAmount: discountAmt,
			FinalPrice:     finalPrice,
			Currency:       currency,
			Status:         models.OrderItemStatusPending,
		})
	}

	subtotal = math.Round(subtotal*100) / 100
	totalTLDDiscount = math.Round(totalTLDDiscount*100) / 100

	// 3. Calculate coupon discount
	var couponDiscount float64
	cartBaseForCoupon := subtotal - totalTLDDiscount

	if cart.CouponCode != nil && *cart.CouponCode != "" {
		couponObj, err := s.discountService.ValidateCoupon(*cart.CouponCode, cartBaseForCoupon)
		if err == nil {
			couponDiscount = s.discountService.CalculateCouponDiscount(couponObj, cartBaseForCoupon)
		}
	}

	couponDiscount = math.Round(couponDiscount*100) / 100
	totalDiscount := math.Round((totalTLDDiscount+couponDiscount)*100) / 100
	totalAmount := math.Max(0, subtotal-totalDiscount)
	totalAmount = math.Round(totalAmount*100) / 100

	if totalAmount <= 0 {
		return nil, errors.New("order total must be greater than zero")
	}

	// 4. Generate order number
	orderNumber := generateOrderNumber()

	// 5. Payment processing
	paymentMethod := "CARD"
	if req.PaymentMethod != "" {
		paymentMethod = strings.ToUpper(strings.TrimSpace(req.PaymentMethod))
	}

	var paymentIntentID string
	var clientSecret string

	if s.stripeService.IsConfigured() && paymentMethod != "BKASH" {
		metadata := map[string]string{
			"order_number": orderNumber,
			"user_id":      fmt.Sprintf("%d", userID),
		}
		pi, err := s.stripeService.CreatePaymentIntent(totalAmount, strings.ToLower(currency), metadata)
		if err != nil {
			return nil, fmt.Errorf("payment processing failed: %w", err)
		}
		paymentIntentID = pi.ID
		clientSecret = pi.ClientSecret
	} else {
		paymentIntentID = fmt.Sprintf("pi_test_%s", orderNumber)
		clientSecret = fmt.Sprintf("pi_test_secret_%s", orderNumber)
	}

	// 6. Persist Order
	order := &models.Order{
		OrderNumber:           orderNumber,
		UserID:                userID,
		Subtotal:              subtotal,
		DiscountAmount:        totalTLDDiscount,
		CouponCode:            cart.CouponCode,
		CouponDiscount:        couponDiscount,
		TotalAmount:           totalAmount,
		Currency:              currency,
		Status:                models.OrderStatusPendingPayment,
		PaymentStatus:         models.PaymentStatusPending,
		PaymentMethod:         paymentMethod,
		StripePaymentIntentID: paymentIntentID,
		StripeClientSecret:    clientSecret,
		Items:                 orderItems,
	}

	if err := s.repo.CreateOrder(order); err != nil {
		return nil, errors.New("failed to create order")
	}

	// 7. Clear cart items and reset cart for future purchases
	s.cartDB.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{})
	s.cartDB.Model(&models.Cart{}).Where("id = ?", cart.ID).Updates(map[string]interface{}{
		"status":      models.CartStatusActive,
		"coupon_code": nil,
	})

	return s.buildOrderResponse(order, true), nil
}

// ============================================================
// DIRECT ORDER (without cart)
// ============================================================

// CreateDirectOrder creates an order directly from a list of domains (no cart needed)
func (s *OrderService) CreateDirectOrder(userID uint, req DirectOrderRequest) (*OrderResponse, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("at least one domain item is required")
	}

	currency := "USD"
	if req.Currency != "" {
		currency = req.Currency
	}

	var subtotal float64
	var totalDiscount float64
	var orderItems []models.OrderItem

	for _, item := range req.Items {
		domainName := strings.ToLower(strings.TrimSpace(item.DomainName))
		tld := extractTLD(domainName)
		if tld == "" {
			return nil, fmt.Errorf("invalid domain name: %s", item.DomainName)
		}

		originalPrice := math.Round(item.UnitPrice*float64(item.Period)*100) / 100
		discountAmt, _ := s.discountService.CalculateItemDiscount(tld, originalPrice)
		finalPrice := math.Max(0, originalPrice-discountAmt)
		finalPrice = math.Round(finalPrice*100) / 100

		subtotal += originalPrice
		totalDiscount += discountAmt

		orderItems = append(orderItems, models.OrderItem{
			DomainName:     domainName,
			TLD:            tld,
			Period:         item.Period,
			UnitPrice:      item.UnitPrice,
			DiscountAmount: discountAmt,
			FinalPrice:     finalPrice,
			Currency:       currency,
			Status:         models.OrderItemStatusPending,
		})
	}

	subtotal = math.Round(subtotal*100) / 100
	totalDiscount = math.Round(totalDiscount*100) / 100
	totalAmount := math.Max(0, subtotal-totalDiscount)
	totalAmount = math.Round(totalAmount*100) / 100

	if totalAmount <= 0 {
		return nil, errors.New("order total must be greater than zero")
	}

	orderNumber := generateOrderNumber()

	var paymentIntentID string
	var clientSecret string

	if s.stripeService.IsConfigured() {
		metadata := map[string]string{
			"order_number": orderNumber,
			"user_id":      fmt.Sprintf("%d", userID),
		}
		pi, err := s.stripeService.CreatePaymentIntent(totalAmount, strings.ToLower(currency), metadata)
		if err != nil {
			return nil, fmt.Errorf("payment processing failed: %w", err)
		}
		paymentIntentID = pi.ID
		clientSecret = pi.ClientSecret
	} else {
		paymentIntentID = fmt.Sprintf("pi_test_%s", orderNumber)
		clientSecret = fmt.Sprintf("pi_test_secret_%s", orderNumber)
	}

	order := &models.Order{
		OrderNumber:           orderNumber,
		UserID:                userID,
		Subtotal:              subtotal,
		DiscountAmount:        totalDiscount,
		TotalAmount:           totalAmount,
		Currency:              currency,
		Status:                models.OrderStatusPendingPayment,
		PaymentStatus:         models.PaymentStatusPending,
		PaymentMethod:         "CARD",
		StripePaymentIntentID: paymentIntentID,
		StripeClientSecret:    clientSecret,
		Items:                 orderItems,
	}

	if err := s.repo.CreateOrder(order); err != nil {
		return nil, errors.New("failed to create order")
	}

	return s.buildOrderResponse(order, true), nil
}

// ============================================================
// PAYMENT CONFIRMATION
// ============================================================

// ConfirmPayment verifies the Stripe payment from the client side and updates the order
func (s *OrderService) ConfirmPayment(userID uint, orderID uint, paymentIntentID string) (*OrderResponse, error) {
	order, err := s.repo.GetOrderByIDAndUser(orderID, userID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	if order.Status != models.OrderStatusPendingPayment {
		return nil, errors.New("order is not awaiting payment")
	}

	if order.StripePaymentIntentID != paymentIntentID {
		return nil, errors.New("payment intent does not match this order")
	}

	// Verify payment status with Stripe if configured and not test ID
	if s.stripeService.IsConfigured() && !strings.HasPrefix(paymentIntentID, "pi_test_") {
		pi, err := s.stripeService.GetPaymentIntent(paymentIntentID)
		if err != nil {
			return nil, errors.New("failed to verify payment with Stripe")
		}

		if pi.Status == stripe.PaymentIntentStatusSucceeded {
			order.Status = models.OrderStatusPaid
			order.PaymentStatus = models.PaymentStatusPaid
			s.repo.UpdateOrder(order)
			s.repo.UpdateOrderItemsStatus(order.ID, models.OrderItemStatusActive)

			// Reseller API integration commented out — uncomment later when ready
			// s.resellerService.RegisterOrderDomains(order)
			log.Printf("[Order] ConfirmPayment success for order #%s — saving domains to DB", order.OrderNumber)
			s.saveOrderDomainsToDB(order)
		} else if pi.Status == stripe.PaymentIntentStatusCanceled {
			order.Status = models.OrderStatusFailed
			order.PaymentStatus = models.PaymentStatusFailed
			s.repo.UpdateOrder(order)
			s.repo.UpdateOrderItemsStatus(order.ID, models.OrderItemStatusFailed)
		}
	} else {
		// Complete test/demo or local payment
		order.Status = models.OrderStatusPaid
		order.PaymentStatus = models.PaymentStatusPaid
		s.repo.UpdateOrder(order)
		s.repo.UpdateOrderItemsStatus(order.ID, models.OrderItemStatusActive)

		// Reseller API integration commented out — uncomment later when ready
		// s.resellerService.RegisterOrderDomains(order)
		log.Printf("[Order] ConfirmPayment (local/test) success for order #%s — saving domains to DB", order.OrderNumber)
		s.saveOrderDomainsToDB(order)
	}

	// Re-fetch for fresh state
	order, _ = s.repo.GetOrderByID(order.ID)
	return s.buildOrderResponse(order, false), nil
}

// ============================================================
// STRIPE WEBHOOK
// ============================================================

// HandleStripeWebhook processes incoming Stripe webhook events
func (s *OrderService) HandleStripeWebhook(payload []byte, sigHeader string) error {
	event, err := s.stripeService.ConstructWebhookEvent(payload, sigHeader)
	if err != nil {
		return fmt.Errorf("webhook signature verification failed: %w", err)
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return fmt.Errorf("failed to parse payment intent: %w", err)
		}
		return s.handlePaymentSuccess(pi.ID)

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return fmt.Errorf("failed to parse payment intent: %w", err)
		}
		return s.handlePaymentFailure(pi.ID)
	}

	return nil
}

func (s *OrderService) handlePaymentSuccess(paymentIntentID string) error {
	order, err := s.repo.GetOrderByPaymentIntentID(paymentIntentID)
	if err != nil {
		return fmt.Errorf("order not found for payment intent: %s", paymentIntentID)
	}

	if order.PaymentStatus == models.PaymentStatusPaid {
		return nil // Already processed (idempotent)
	}

	order.Status = models.OrderStatusPaid
	order.PaymentStatus = models.PaymentStatusPaid
	if err := s.repo.UpdateOrder(order); err != nil {
		return errors.New("failed to update order status")
	}

	s.repo.UpdateOrderItemsStatus(order.ID, models.OrderItemStatusActive)

	// Reseller API integration commented out — uncomment later when ready
	// s.resellerService.RegisterOrderDomains(order)
	log.Printf("[Order] Payment success for order #%s — saving domains to DB", order.OrderNumber)
	s.saveOrderDomainsToDB(order)

	return nil
}

func (s *OrderService) handlePaymentFailure(paymentIntentID string) error {
	order, err := s.repo.GetOrderByPaymentIntentID(paymentIntentID)
	if err != nil {
		return fmt.Errorf("order not found for payment intent: %s", paymentIntentID)
	}

	order.Status = models.OrderStatusFailed
	order.PaymentStatus = models.PaymentStatusFailed
	if err := s.repo.UpdateOrder(order); err != nil {
		return errors.New("failed to update order status")
	}

	s.repo.UpdateOrderItemsStatus(order.ID, models.OrderItemStatusFailed)
	return nil
}

// ============================================================
// QUERY METHODS
// ============================================================

// GetOrderByID fetches a single order — regular users can only see their own orders
func (s *OrderService) GetOrderByID(userID uint, orderID uint, role string) (*OrderResponse, error) {
	var order *models.Order
	var err error

	if role == string(models.RoleAdmin) || role == string(models.RoleSuperAdmin) {
		order, err = s.repo.GetOrderByID(orderID)
	} else {
		order, err = s.repo.GetOrderByIDAndUser(orderID, userID)
	}

	if err != nil {
		return nil, errors.New("order not found")
	}

	return s.buildOrderResponse(order, false), nil
}

// GetUserOrders returns a paginated list of orders for the authenticated user
func (s *OrderService) GetUserOrders(userID uint, page, limit int) (*OrderListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	orders, total, err := s.repo.GetUserOrders(userID, page, limit)
	if err != nil {
		return nil, errors.New("failed to fetch orders")
	}

	return s.buildOrderListResponse(orders, total, page, limit), nil
}

// AdminGetAllOrders returns a paginated list of all orders (admin/super-admin only)
func (s *OrderService) AdminGetAllOrders(page, limit int, status string, search string) (*OrderListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	orders, total, err := s.repo.GetAllOrders(page, limit, status, search)
	if err != nil {
		return nil, errors.New("failed to fetch orders")
	}

	return s.buildOrderListResponse(orders, total, page, limit), nil
}

// SendPaymentReminder sends an email reminder to customer for unpaid orders
func (s *OrderService) SendPaymentReminder(orderID uint) error {
	order, err := s.repo.GetOrderByID(orderID)
	if err != nil {
		return errors.New("order not found")
	}

	if order.PaymentStatus == models.PaymentStatusPaid {
		return errors.New("order is already paid")
	}

	if order.User.Email == "" {
		return errors.New("customer email not found for this order")
	}

	subject := fmt.Sprintf("Payment Reminder: Domain Order #%s", order.OrderNumber)
	body := fmt.Sprintf(`Hello %s,<br><br>
This is a reminder that payment for your domain order <strong>#%s</strong> (Total: %s %.2f) is still pending.<br><br>
Please complete your payment to secure your domain registration before it becomes available for others to register.<br><br>
Thank you for choosing domain.bd!`,
		order.User.Name,
		order.OrderNumber,
		order.Currency,
		order.TotalAmount,
	)

	return mail.SendNotificationEmail(order.User.Email, subject, body)
}

// AdminUpdateOrderStatus allows admin to update the order status and/or payment status
func (s *OrderService) AdminUpdateOrderStatus(orderID uint, req AdminUpdateOrderStatusRequest) (*OrderResponse, error) {
	order, err := s.repo.GetOrderByID(orderID)
	if err != nil {
		return nil, errors.New("order not found")
	}

	if req.Status != "" {
		order.Status = models.OrderStatus(req.Status)
	}
	if req.PaymentStatus != "" {
		order.PaymentStatus = models.PaymentStatus(req.PaymentStatus)
	}

	if err := s.repo.UpdateOrder(order); err != nil {
		return nil, errors.New("failed to update order: " + err.Error())
	}

	// If marked paid, update items to active and save domains to DB
	if order.PaymentStatus == models.PaymentStatusPaid {
		_ = s.repo.UpdateOrderItemsStatus(order.ID, models.OrderItemStatusActive)
		// Reseller API integration commented out — uncomment later when ready
		// go s.resellerService.RegisterOrderDomains(order)
		go s.saveOrderDomainsToDB(order)
	}

	fresh, _ := s.repo.GetOrderByID(order.ID)
	return s.buildOrderResponse(fresh, false), nil
}

// ============================================================
// RESPONSE BUILDERS
// ============================================================

func (s *OrderService) buildOrderResponse(order *models.Order, includeClientSecret bool) *OrderResponse {
	var items []OrderItemResponse
	for _, item := range order.Items {
		items = append(items, OrderItemResponse{
			ID:             item.ID,
			DomainName:     item.DomainName,
			TLD:            item.TLD,
			Period:         item.Period,
			UnitPrice:      item.UnitPrice,
			DiscountAmount: item.DiscountAmount,
			FinalPrice:     item.FinalPrice,
			Currency:       item.Currency,
			Status:         string(item.Status),
		})
	}

	resp := &OrderResponse{
		ID:                    order.ID,
		OrderNumber:           order.OrderNumber,
		Items:                 items,
		ItemsCount:            len(items),
		Subtotal:              order.Subtotal,
		DiscountAmount:        order.DiscountAmount,
		CouponCode:            order.CouponCode,
		CouponDiscount:        order.CouponDiscount,
		TotalAmount:           order.TotalAmount,
		Currency:              order.Currency,
		Status:                string(order.Status),
		PaymentStatus:         string(order.PaymentStatus),
		PaymentMethod:         order.PaymentMethod,
		StripePaymentIntentID: order.StripePaymentIntentID,
		CreatedAt:             order.CreatedAt.Format(time.RFC3339),
	}

	if order.User.ID > 0 {
		resp.User = &OrderUserResponse{
			ID:          order.User.ID,
			Name:        order.User.Name,
			Email:       order.User.Email,
			PhoneNumber: order.User.PhoneNumber,
		}
	}

	if includeClientSecret {
		resp.StripeClientSecret = order.StripeClientSecret
	}

	return resp
}

func (s *OrderService) buildOrderListResponse(orders []models.Order, total int64, page, limit int) *OrderListResponse {
	var orderResponses []OrderResponse
	for _, o := range orders {
		orderResponses = append(orderResponses, *s.buildOrderResponse(&o, false))
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &OrderListResponse{
		Orders:     orderResponses,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}

// ============================================================
// HELPERS
// ============================================================

// generateOrderNumber creates a unique order number: ORD-YYYYMMDD-XXXXXXXX
func generateOrderNumber() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("ORD-%s-%s",
		time.Now().Format("20060102"),
		strings.ToUpper(hex.EncodeToString(b)),
	)
}

// extractTLD extracts the TLD from a domain name (e.g. "example.com" -> "com")
func extractTLD(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[1:], ".")
}

// saveOrderDomainsToDB creates/updates Domain records directly in the local database
// upon successful order completion.
// NOTE: Reseller API calls are bypassed as requested and can be re-enabled later.
func (s *OrderService) saveOrderDomainsToDB(order *models.Order) {
	if order == nil {
		return
	}

	for _, item := range order.Items {
		domainName := strings.ToLower(strings.TrimSpace(item.DomainName))
		if domainName == "" {
			continue
		}

		period := item.Period
		if period < 1 {
			period = 1
		}

		now := time.Now()
		expiresAt := now.AddDate(period, 0, 0)

		tld := extractTLD(domainName)
		if tld == "" {
			tld = "com"
		}

		// Check if domain already exists in DB
		var existing models.Domain
		if s.cartDB.Where("domain_name = ?", domainName).First(&existing).Error == nil {
			// Domain already exists in DB — update active status, owner, and expiry
			s.cartDB.Model(&existing).Updates(map[string]interface{}{
				"user_id":         order.UserID,
				"status":          models.DomainStatusActive,
				"reseller_status": "LOCAL_ACTIVE",
				"expires_at":      expiresAt,
			})
			log.Printf("[Order] Updated existing domain record in DB: %s (User: %d)", domainName, order.UserID)
		} else {
			// Create new domain record directly in database
			newDomain := models.Domain{
				UserID:            order.UserID,
				DomainName:        domainName,
				TLD:               tld,
				Status:            models.DomainStatusActive,
				AutoRenew:         true,
				RegistrationDate:  now,
				ExpiresAt:         expiresAt,
				PrivacyProtection: true,
				Nameservers:       "ns1.domain.bd,ns2.domain.bd",
				ResellerDomainID:  "",
				ResellerStatus:    "LOCAL_ACTIVE",
			}

			if err := s.cartDB.Create(&newDomain).Error; err != nil {
				log.Printf("[Order] ❌ Failed to save domain %s to DB: %v", domainName, err)
			} else {
				log.Printf("[Order] 💾 Domain record saved to DB: %s (ID: %d, User: %d)", domainName, newDomain.ID, order.UserID)
			}
		}
	}
}
