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
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Please log in to continue")
	}

	var req CreateDNSRecordRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "We couldn't process your request. Please check the form and try again")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Please fill in all required fields correctly")
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
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Please log in to continue")
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
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Please log in to continue")
	}

	recordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "The requested DNS record could not be found")
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
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Please log in to continue")
	}

	recordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "The requested DNS record could not be found")
	}

	var req UpdateDNSRecordRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "We couldn't process your request. Please check the form and try again")
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
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Please log in to continue")
	}

	recordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "The requested DNS record could not be found")
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
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Please log in to continue")
	}

	domainName := c.QueryParam("domainName")
	hostName := c.QueryParam("hostName")

	if domainName == "" || hostName == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Please specify both domain and host name")
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
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Please log in to continue")
	}

	recordID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "The requested DNS record could not be found")
	}

	record, err := h.service.ResyncDNSRecord(userID, uint(recordID))
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "DNS record re-syncing to reseller API", record)
}

// ============================================================
// NAMESERVER MANAGEMENT
// ============================================================

// UpdateNameServers updates nameservers for a domain and syncs to DNA API via PUT
// PUT /api/v1/dns/name-server
func (h *DNSHandler) UpdateNameServers(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Please log in to continue")
	}

	var req UpdateNameServerRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "We couldn't process your request. Please check the form and try again")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Please fill in all required fields correctly")
	}

	result, err := h.service.UpdateNameServers(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Nameservers updated and syncing to reseller API", result)
}

// GetNameServers retrieves nameservers for a domain
// GET /api/v1/dns/name-server?domain=example.com
func (h *DNSHandler) GetNameServers(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Please log in to continue")
	}

	domainName := c.QueryParam("domain")
	if domainName == "" {
		domainName = c.QueryParam("domainName")
	}
	if domainName == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Please select a domain to view nameservers")
	}

	result, err := h.service.GetNameServers(userID, domainName)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Nameservers retrieved successfully", result)
}

