package find_domain

import (
	"fmt"
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
// GET /api/v1/find-domain/search?query=eftakhar
// GET /api/v1/find-domain/search?domain=eftakhar.com
func (h *DomainHandler) Search(c *echo.Context) error {

	query := c.QueryParam("query")

	if query == "" {
		query = c.QueryParam("domain")
	}
	fmt.Println("Search query: ", query)
	if query == "" {
		return utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"query=example or domain=example.com parameter is required",
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
// GET /api/v1/find-domain/bulk-search?domain=eftakhar
// GET /api/v1/find-domain/bulk-search?query=eftakhar
func (h *DomainHandler) BulkSearch(c *echo.Context) error {

	domain := c.QueryParam("domain")
	if domain == "" {
		domain = c.QueryParam("query")
	}

	if domain == "" {
		return utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"domain=example or query=example parameter is required",
		)
	}

	result, err := h.service.BulkSearch(domain)
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
