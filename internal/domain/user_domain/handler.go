package user_domain

import (
	"log"
	"net/http"
	"strconv"

	"project-setup/internal/pkg/utils"

	"github.com/labstack/echo/v5"
)

// DomainHandler handles HTTP requests for user domain operations
type DomainHandler struct {
	service *DomainService
}

func NewDomainHandler(service *DomainService) *DomainHandler {
	return &DomainHandler{service: service}
}

// GetUserDomains returns all domains owned by the user
// GET /api/v1/user/domains or GET /api/v1/domains
func (h *DomainHandler) GetUserDomains(c *echo.Context) error {
	log.Printf("[UserDomain API] 📥 Received GET /domains request from IP: %s\n", c.RealIP())

	userID, ok := c.Get("user_id").(uint)
	if !ok {
		log.Printf("[UserDomain API] ❌ Unauthorized: user_id missing or invalid type in context\n")
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
	}

	search := c.QueryParam("q")
	if search == "" {
		search = c.QueryParam("search")
	}

	log.Printf("[UserDomain API] 🔍 Querying database for UserID: %d, Search: '%s'\n", userID, search)
	result, err := h.service.GetUserDomains(userID, search)
	if err != nil {
		log.Printf("[UserDomain API] ❌ Database error for UserID %d: %v\n", userID, err)
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to load domains: "+err.Error())
	}

	log.Printf("[UserDomain API] ✅ Successfully fetched %d domains for UserID: %d\n", len(result.Domains), userID)
	return utils.SuccessResponse(c, http.StatusOK, "Domains retrieved successfully", result)
}

// GetDomainByID returns a single domain by ID
// GET /api/v1/user/domains/:id
func (h *DomainHandler) GetDomainByID(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
	}

	idParam := c.Param("id")
	domainID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid domain ID")
	}

	domain, err := h.service.GetDomainByID(userID, uint(domainID))
	if err != nil {
		return utils.ErrorResponse(c, http.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Domain retrieved successfully", domain)
}

// AdminGetAllDomains returns all domains across the platform with owner details
// GET /api/v1/domains/admin/all
func (h *DomainHandler) AdminGetAllDomains(c *echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	search := c.QueryParam("q")
	if search == "" {
		search = c.QueryParam("search")
	}
	status := c.QueryParam("status")

	result, err := h.service.AdminGetAllDomains(page, limit, search, status)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Admin domains retrieved successfully", result)
}

// AdminUpdateDomain updates status and settings for a domain
// PUT /api/v1/domains/admin/:id/status
func (h *DomainHandler) AdminUpdateDomain(c *echo.Context) error {
	domainID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid domain ID")
	}

	var req AdminUpdateDomainStatusRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	domain, err := h.service.AdminUpdateDomain(uint(domainID), req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Domain updated successfully", domain)
}

