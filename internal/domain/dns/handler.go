package dns

import (
	"net/http"
	"strconv"

	"project-setup/internal/pkg/utils"

	"github.com/labstack/echo/v5"
)

// DNSHandler handles HTTP requests for DNS record operations
type DNSHandler struct {
	service *DNSService
}

// NewDNSHandler creates a new DNSHandler
func NewDNSHandler(service *DNSService) *DNSHandler {
	return &DNSHandler{service: service}
}

// ============================================================
// CREATE DNS RECORD
// ============================================================

// CreateRecord creates a new DNS host record and syncs to DNA API
// POST /api/v1/dns
func (h *DNSHandler) CreateRecord(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	var req CreateDNSRecordRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	records, err := h.service.CreateDNSRecord(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusCreated, "DNS record created and syncing to reseller API", records)
}

// ============================================================
// GET DNS RECORDS
// ============================================================

// GetRecordsByDomain returns all DNS records for a specific domain
// GET /api/v1/dns?domain=example.com
func (h *DNSHandler) GetRecordsByDomain(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	domainName := c.QueryParam("domain")
	if domainName == "" {
		// Return all DNS records for the user
		records, err := h.service.GetAllUserDNSRecords(userID)
		if err != nil {
			return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return utils.SuccessResponse(c, http.StatusOK, "DNS records retrieved successfully", records)
	}

	records, err := h.service.GetDNSRecordsByDomain(userID, domainName)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "DNS records retrieved successfully", records)
}

// GetRecordByID returns a single DNS record
// GET /api/v1/dns/:id
func (h *DNSHandler) GetRecordByID(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	recordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid record ID")
	}

	record, err := h.service.GetDNSRecordByID(userID, uint(recordID))
	if err != nil {
		return utils.ErrorResponse(c, http.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "DNS record retrieved successfully", record)
}

// ============================================================
// UPDATE DNS RECORD
// ============================================================

// UpdateRecord updates an existing DNS record and re-syncs
// PUT /api/v1/dns/:id
func (h *DNSHandler) UpdateRecord(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	recordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid record ID")
	}

	var req UpdateDNSRecordRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	record, err := h.service.UpdateDNSRecord(userID, uint(recordID), req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "DNS record updated and re-syncing", record)
}

// ============================================================
// DELETE DNS RECORD
// ============================================================

// DeleteRecord deletes a DNS record
// DELETE /api/v1/dns/:id
func (h *DNSHandler) DeleteRecord(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	recordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid record ID")
	}

	if err := h.service.DeleteDNSRecord(userID, uint(recordID)); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "DNS record deleted successfully", nil)
}

// DeleteByHost deletes DNS records by domain name + host name and syncs to DNA API
// DELETE /api/v1/dns/host?domainName=example.com&hostName=ns1
func (h *DNSHandler) DeleteByHost(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	domainName := c.QueryParam("domainName")
	hostName := c.QueryParam("hostName")

	if domainName == "" || hostName == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "domainName and hostName query parameters are required")
	}

	req := DeleteDNSByHostRequest{
		DomainName: domainName,
		HostName:   hostName,
	}

	if err := h.service.DeleteDNSByHost(userID, req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "DNS host records deleted and syncing to reseller API", nil)
}

// ============================================================
// RESYNC DNS RECORD
// ============================================================

// ResyncRecord re-attempts syncing a failed DNS record to the DNA API
// POST /api/v1/dns/:id/resync
func (h *DNSHandler) ResyncRecord(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	recordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid record ID")
	}

	record, err := h.service.ResyncDNSRecord(userID, uint(recordID))
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "DNS record re-syncing to reseller API", record)
}
