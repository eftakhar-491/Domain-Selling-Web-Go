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

// UpdateDomain modifies settings for a specific domain
// PUT /api/v1/user/domains/:id or PATCH /api/v1/user/domains/:id
func (h *DomainHandler) UpdateDomain(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		log.Printf("[UserDomain API] ❌ Unauthorized UpdateDomain: user_id missing\n")
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
	}

	idParam := c.Param("id")
	domainID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid domain ID")
	}

	var req UpdateDomainRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	log.Printf("[UserDomain API] ✏️ Updating domain %d for UserID: %d\n", domainID, userID)
	domain, err := h.service.UpdateDomain(userID, uint(domainID), req)
	if err != nil {
		log.Printf("[UserDomain API] ❌ Update failed for domain %d: %v\n", domainID, err)
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	log.Printf("[UserDomain API] ✅ Successfully updated domain %d (%s)\n", domainID, domain.DomainName)
	return utils.SuccessResponse(c, http.StatusOK, "Domain updated successfully", domain)
}

// CreateDomain registers / creates a new domain entry
// POST /api/v1/user/domains
func (h *DomainHandler) CreateDomain(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		log.Printf("[UserDomain API] ❌ Unauthorized CreateDomain: user_id missing\n")
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
	}

	var req CreateDomainRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	log.Printf("[UserDomain API] ➕ Creating domain '%s' for UserID: %d\n", req.DomainName, userID)
	domain, err := h.service.CreateDomain(userID, req)
	if err != nil {
		log.Printf("[UserDomain API] ❌ Create domain failed for '%s': %v\n", req.DomainName, err)
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	log.Printf("[UserDomain API] ✅ Created domain '%s' (ID: %d) successfully\n", domain.DomainName, domain.ID)
	return utils.SuccessResponse(c, http.StatusCreated, "Domain registered successfully", domain)
}

// DeleteDomain removes a domain
// DELETE /api/v1/user/domains/:id
func (h *DomainHandler) DeleteDomain(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		log.Printf("[UserDomain API] ❌ Unauthorized DeleteDomain: user_id missing\n")
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
	}

	idParam := c.Param("id")
	domainID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid domain ID")
	}

	log.Printf("[UserDomain API] 🗑️ Deleting domain %d for UserID: %d\n", domainID, userID)
	if err := h.service.DeleteDomain(userID, uint(domainID)); err != nil {
		log.Printf("[UserDomain API] ❌ Delete domain %d failed: %v\n", domainID, err)
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete domain")
	}

	log.Printf("[UserDomain API] ✅ Deleted domain %d successfully\n", domainID)
	return utils.SuccessResponse(c, http.StatusOK, "Domain deleted successfully", nil)
}
