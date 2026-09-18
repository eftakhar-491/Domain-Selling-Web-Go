package order

import (
	"io"
	"net/http"
	"strconv"

	"project-setup/internal/pkg/utils"

	"github.com/labstack/echo/v5"
)

// OrderHandler handles HTTP requests for order operations
type OrderHandler struct {
	service *OrderService
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(service *OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// Checkout creates an order from the user's active cart
// POST /api/v1/orders/checkout
func (h *OrderHandler) Checkout(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	var req CreateOrderFromCartRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	order, err := h.service.CheckoutFromCart(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusCreated, "Order created successfully. Complete payment to confirm.", order)
}

// DirectOrder creates an order directly without using a cart
// POST /api/v1/orders/direct
func (h *OrderHandler) DirectOrder(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	var req DirectOrderRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	order, err := h.service.CreateDirectOrder(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusCreated, "Order created successfully. Complete payment to confirm.", order)
}

// ConfirmPayment verifies payment status from the client and updates the order
// POST /api/v1/orders/:id/confirm-payment
func (h *OrderHandler) ConfirmPayment(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID")
	}

	var req ConfirmPaymentRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	order, err := h.service.ConfirmPayment(userID, uint(orderID), req.PaymentIntentID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Payment confirmed", order)
}

// GetOrder returns a single order by ID
// GET /api/v1/orders/:id
func (h *OrderHandler) GetOrder(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID")
	}

	role, _ := c.Get("role").(string)

	order, err := h.service.GetOrderByID(userID, uint(orderID), role)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Order retrieved successfully", order)
}

// GetMyOrders returns the authenticated user's orders with pagination
// GET /api/v1/orders
func (h *OrderHandler) GetMyOrders(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	orders, err := h.service.GetUserOrders(userID, page, limit)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Orders retrieved successfully", orders)
}

// AdminGetAllOrders returns all orders (admin only) with pagination, search, and status filter
// GET /api/v1/orders/admin/all
func (h *OrderHandler) AdminGetAllOrders(c *echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	status := c.QueryParam("status")
	search := c.QueryParam("q")
	if search == "" {
		search = c.QueryParam("search")
	}

	orders, err := h.service.AdminGetAllOrders(page, limit, status, search)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Orders retrieved successfully", orders)
}

// SendPaymentReminder sends an email reminder to customer for an unpaid order
// POST /api/v1/orders/admin/:id/reminder
func (h *OrderHandler) SendPaymentReminder(c *echo.Context) error {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID")
	}

	if err := h.service.SendPaymentReminder(uint(orderID)); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Payment reminder sent successfully to customer", nil)
}

// AdminUpdateOrderStatus updates order status and/or payment status
// PUT /api/v1/orders/admin/:id/status
func (h *OrderHandler) AdminUpdateOrderStatus(c *echo.Context) error {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID")
	}

	var req AdminUpdateOrderStatusRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	order, err := h.service.AdminUpdateOrderStatus(uint(orderID), req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Order status updated successfully", order)
}


// HandleWebhook processes Stripe webhook events
// POST /api/v1/orders/webhook
func (h *OrderHandler) HandleWebhook(c *echo.Context) error {
	// Read the raw request body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Failed to read request body")
	}

	sigHeader := c.Request().Header.Get("Stripe-Signature")

	if err := h.service.HandleStripeWebhook(body, sigHeader); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Webhook processed", nil)
}
