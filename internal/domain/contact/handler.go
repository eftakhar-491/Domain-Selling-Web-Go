package contact

import (
	"net/http"
	"strconv"

	"project-setup/internal/pkg/utils"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// SendMessage handles public contact submissions
// POST /api/v1/contact
func (h *Handler) SendMessage(c *echo.Context) error {
	var req CreateContactRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	msg, err := h.service.SendMessage(&req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Thank you! Your message has been sent successfully. Our team will get back to you shortly.", msg)
}

// GetAll handles admin retrieval of contact messages
// GET /api/v1/contact
func (h *Handler) GetAll(c *echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	messages, total, err := h.service.GetAllMessages(page, limit)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch contact messages: "+err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Contact messages fetched successfully", map[string]interface{}{
		"messages": messages,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// MarkStatus handles marking message status (READ/ARCHIVED)
// PUT /api/v1/contact/:id/status
func (h *Handler) MarkStatus(c *echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid message ID")
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&payload); err != nil || payload.Status == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Valid status is required")
	}

	if err := h.service.MarkStatus(uint(id), payload.Status); err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update status: "+err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Message status updated successfully", nil)
}
