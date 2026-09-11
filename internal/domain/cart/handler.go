package cart

import (
	"net/http"
	"strconv"

	"project-setup/internal/pkg/utils"

	"github.com/labstack/echo/v5"
)

type CartHandler struct {
	service *CartService
}

func NewCartHandler(service *CartService) *CartHandler {
	return &CartHandler{service: service}
}

// GetCart returns the user's cart with all discounts calculated
// GET /api/v1/cart
func (h *CartHandler) GetCart(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	cart, err := h.service.GetCart(userID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Cart retrieved successfully", cart)
}

// AddToCart adds a domain to the cart
// POST /api/v1/cart/items
func (h *CartHandler) AddToCart(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	var req AddToCartRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	cart, err := h.service.AddToCart(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Item added to cart", cart)
}

// UpdateCartItem updates a cart item's period
// PUT /api/v1/cart/items/:id
func (h *CartHandler) UpdateCartItem(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID")
	}

	var req UpdateCartItemRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	cart, err := h.service.UpdateItem(userID, uint(itemID), req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Cart item updated", cart)
}

// RemoveCartItem removes an item from the cart
// DELETE /api/v1/cart/items/:id
func (h *CartHandler) RemoveCartItem(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	itemID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID")
	}

	cart, err := h.service.RemoveItem(userID, uint(itemID))
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Item removed from cart", cart)
}

// ClearCart removes all items from the cart
// DELETE /api/v1/cart
func (h *CartHandler) ClearCart(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	if err := h.service.ClearCart(userID); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Cart cleared successfully", nil)
}

// ApplyCoupon applies a coupon code to the cart
// POST /api/v1/cart/coupon
func (h *CartHandler) ApplyCoupon(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	var req ApplyCouponRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	cart, err := h.service.ApplyCoupon(userID, req.Code)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Coupon applied successfully", cart)
}

// RemoveCoupon removes the coupon from the cart
// DELETE /api/v1/cart/coupon
func (h *CartHandler) RemoveCoupon(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	cart, err := h.service.RemoveCoupon(userID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Coupon removed successfully", cart)
}
