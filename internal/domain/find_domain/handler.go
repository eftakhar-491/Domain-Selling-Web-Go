package find_domain

import (
	"net/http"

	"project-setup/internal/pkg/utils"

	"github.com/labstack/echo/v5"
)

// DomainHandler handles HTTP requests for domain search
type DomainHandler struct {
	service *DomainService
}

// NewDomainHandler creates a new DomainHandler instance
func NewDomainHandler(service *DomainService) *DomainHandler {
	return &DomainHandler{
		service: service,
	}
}

// Search
//
// GET /api/v1/find-domain/search?query=eftakhar
// GET /api/v1/find-domain/search?domain=eftakhar.com
func (h *DomainHandler) Search(c *echo.Context) error {

	query := c.QueryParam("query")

	if query == "" {
		query = c.QueryParam("domain")
	}

	if query == "" {
		return utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"query or domain parameter is required",
		)
	}

	result, err := h.service.Search(query)
	if err != nil {
		return utils.ErrorResponse(
			c,
			http.StatusBadGateway,
			err.Error(),
		)
	}

	return utils.SuccessResponse(
		c,
		http.StatusOK,
		"Domain search completed successfully",
		result,
	)
}

// BulkSearch
//
// POST /api/v1/find-domain/bulk-search
//
// Body:
//
//	{
//	  "domainName": "eftakhar"
//	}
func (h *DomainHandler) BulkSearch(c *echo.Context) error {

	var req BulkSearchRequest

	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid request body",
		)
	}

	if req.DomainName == "" {
		return utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"domainName is required",
		)
	}

	result, err := h.service.BulkSearch(req.DomainName)
	if err != nil {
		return utils.ErrorResponse(
			c,
			http.StatusBadGateway,
			err.Error(),
		)
	}

	return utils.SuccessResponse(
		c,
		http.StatusOK,
		"Bulk domain search completed successfully",
		result,
	)
}
