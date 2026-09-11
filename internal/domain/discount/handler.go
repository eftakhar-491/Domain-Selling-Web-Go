package discount

import (
	"net/http"
	"strconv"

	"project-setup/internal/pkg/utils"

	"github.com/labstack/echo/v5"
)

type DiscountHandler struct {
	service *DiscountService
}

func NewDiscountHandler(service *DiscountService) *DiscountHandler {
	return &DiscountHandler{service: service}
}

// Create creates a new discount rule
// POST /api/v1/discounts
func (h *DiscountHandler) Create(c *echo.Context) error {
	var req CreateDiscountRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	discount, err := h.service.Create(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusCreated, "Discount created successfully", discount)
}

// GetAll returns a paginated list of all discounts
// GET /api/v1/discounts?page=1&limit=10
func (h *DiscountHandler) GetAll(c *echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	discounts, total, err := h.service.GetAll(page, limit)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch discounts")
	}

	return utils.SuccessResponse(c, http.StatusOK, "Discounts retrieved successfully", map[string]interface{}{
		"discounts": discounts,
		"total":     total,
		"page":      page,
		"limit":     limit,
	})
}

// GetByID returns a single discount by ID
// GET /api/v1/discounts/:id
func (h *DiscountHandler) GetByID(c *echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid discount ID")
	}

	discount, err := h.service.GetByID(uint(id))
	if err != nil {
		return utils.ErrorResponse(c, http.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Discount retrieved successfully", discount)
}

// Update updates an existing discount
// PUT /api/v1/discounts/:id
func (h *DiscountHandler) Update(c *echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid discount ID")
	}

	var req UpdateDiscountRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	discount, err := h.service.Update(uint(id), req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Discount updated successfully", discount)
}

// Delete removes a discount
// DELETE /api/v1/discounts/:id
func (h *DiscountHandler) Delete(c *echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid discount ID")
	}

	if err := h.service.Delete(uint(id)); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Discount deleted successfully", nil)
}

// GetActiveTLDs returns all active TLD-based discounts
// GET /api/v1/discounts/tlds
func (h *DiscountHandler) GetActiveTLDs(c *echo.Context) error {
	discounts, err := h.service.GetActiveTLDDiscounts()
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch TLD discounts")
	}

	return utils.SuccessResponse(c, http.StatusOK, "Active TLD discounts retrieved", discounts)
}
