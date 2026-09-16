package dns

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"project-setup/internal/models"
)

// DNSService contains all business logic for DNS record management
type DNSService struct {
	repo     *DNSRepository
	reseller *DNAResellerClient
}

// NewDNSService creates a new DNSService
func NewDNSService(repo *DNSRepository) *DNSService {
	return &DNSService{
		repo:     repo,
		reseller: NewDNAResellerClient(),
	}
}

// ============================================================
// CREATE DNS RECORD
// ============================================================

// CreateDNSRecord saves the DNS record to DB and syncs it to the DNA reseller API
func (s *DNSService) CreateDNSRecord(userID uint, req CreateDNSRecordRequest) ([]DNSRecordResponse, error) {
	domainName := strings.ToLower(strings.TrimSpace(req.DomainName))
	hostName := strings.TrimSpace(req.HostName)

	if domainName == "" {
		return nil, errors.New("domain name is required")
	}
	if hostName == "" {
		return nil, errors.New("host name is required")
	}
	if len(req.IPAddresses) == 0 {
		return nil, errors.New("at least one IP address is required")
	}

	// Validate that the user owns this domain
	domain, err := s.repo.GetDomainByNameAndUser(domainName, userID)
	if err != nil {
		return nil, fmt.Errorf("domain '%s' not found or you don't have access", domainName)
	}

	// Set defaults
	recordType := models.DNSRecordTypeA
	if req.RecordType != "" {
		recordType = models.DNSRecordType(strings.ToUpper(req.RecordType))
	}
	ttl := 3600
	if req.TTL > 0 {
		ttl = req.TTL
	}

	// Save each IP address as a separate DNS record in DB
	var createdRecords []models.DNSRecord
	for _, ip := range req.IPAddresses {
		ipVersion := strings.TrimSpace(ip.IPVersion)
		if ipVersion == "" {
			ipVersion = "IPv4"
		}

		record := models.DNSRecord{
			UserID:     userID,
			DomainID:   domain.ID,
			DomainName: domainName,
			HostName:   hostName,
			RecordType: recordType,
			IPAddress:  strings.TrimSpace(ip.IPAddress),
			IPVersion:  ipVersion,
			TTL:        ttl,
			Priority:   req.Priority,
			SyncStatus: models.DNSSyncStatusPending,
		}

		if err := s.repo.CreateRecord(&record); err != nil {
			log.Printf("[DNS] ❌ Failed to save DNS record to DB: %v", err)
			return nil, fmt.Errorf("failed to save DNS record: %w", err)
		}

		createdRecords = append(createdRecords, record)
	}

	// Build DNA API IP list and sync to external API (async)
	go s.syncToDNA(domainName, hostName, req.IPAddresses, createdRecords)

	// Build response
	var responses []DNSRecordResponse
	for _, rec := range createdRecords {
		responses = append(responses, s.toRecordResponse(&rec))
	}

	return responses, nil
}

// ============================================================
// GET DNS RECORDS
// ============================================================

// GetDNSRecordsByDomain returns all DNS records for a domain name owned by the user
func (s *DNSService) GetDNSRecordsByDomain(userID uint, domainName string) (*DNSRecordListResponse, error) {
	domainName = strings.ToLower(strings.TrimSpace(domainName))
	if domainName == "" {
		return nil, errors.New("domain name is required")
	}

	// Verify domain ownership
	_, err := s.repo.GetDomainByNameAndUser(domainName, userID)
	if err != nil {
		return nil, fmt.Errorf("domain '%s' not found or you don't have access", domainName)
	}

	records, err := s.repo.GetRecordsByDomainName(userID, domainName)
	if err != nil {
		return nil, errors.New("failed to fetch DNS records")
	}

	var items []DNSRecordResponse
	for _, r := range records {
		items = append(items, s.toRecordResponse(&r))
	}

	return &DNSRecordListResponse{
		Records:    items,
		DomainName: domainName,
		TotalCount: int64(len(items)),
	}, nil
}

// GetDNSRecordByID returns a single DNS record by ID
func (s *DNSService) GetDNSRecordByID(userID uint, recordID uint) (*DNSRecordResponse, error) {
	record, err := s.repo.GetRecordByIDAndUser(recordID, userID)
	if err != nil {
		return nil, errors.New("DNS record not found")
	}

	resp := s.toRecordResponse(record)
	return &resp, nil
}

// GetAllUserDNSRecords returns all DNS records for the authenticated user
func (s *DNSService) GetAllUserDNSRecords(userID uint) (*DNSRecordListResponse, error) {
	records, err := s.repo.GetAllUserRecords(userID)
	if err != nil {
		return nil, errors.New("failed to fetch DNS records")
	}

	var items []DNSRecordResponse
	for _, r := range records {
		items = append(items, s.toRecordResponse(&r))
	}

	return &DNSRecordListResponse{
		Records:    items,
		TotalCount: int64(len(items)),
	}, nil
}

// ============================================================
// UPDATE DNS RECORD
// ============================================================

// UpdateDNSRecord updates an existing DNS record and syncs via PUT to DNA API
func (s *DNSService) UpdateDNSRecord(userID uint, recordID uint, req UpdateDNSRecordRequest) (*DNSRecordResponse, error) {
	record, err := s.repo.GetRecordByIDAndUser(recordID, userID)
	if err != nil {
		return nil, errors.New("DNS record not found")
	}

	// Track the old hostname for the DNA API update call
	oldHostName := record.HostName
	newHostName := oldHostName

	// Apply updates
	if req.NewHostName != nil && *req.NewHostName != "" {
		newHostName = strings.TrimSpace(*req.NewHostName)
		record.HostName = newHostName
	} else if req.HostName != nil && *req.HostName != "" {
		newHostName = strings.TrimSpace(*req.HostName)
		record.HostName = newHostName
	}

	if req.RecordType != nil && *req.RecordType != "" {
		record.RecordType = models.DNSRecordType(strings.ToUpper(*req.RecordType))
	}
	if req.TTL != nil && *req.TTL > 0 {
		record.TTL = *req.TTL
	}
	if req.Priority != nil {
		record.Priority = *req.Priority
	}

	// If IP addresses provided, update the first one (since each record stores one IP)
	if len(req.IPAddresses) > 0 {
		record.IPAddress = strings.TrimSpace(req.IPAddresses[0].IPAddress)
		ipVersion := strings.TrimSpace(req.IPAddresses[0].IPVersion)
		if ipVersion != "" {
			record.IPVersion = ipVersion
		}
	}

	// Resolve the domainName — use from request if provided, otherwise from record
	domainName := record.DomainName
	if req.DomainName != "" {
		domainName = strings.ToLower(strings.TrimSpace(req.DomainName))
	}

	// Mark as pending sync
	record.SyncStatus = models.DNSSyncStatusPending
	record.SyncError = ""

	if err := s.repo.UpdateRecord(record); err != nil {
		return nil, errors.New("failed to update DNS record")
	}

	// Sync update to DNA API via PUT (async)
	go s.syncUpdateToDNA(record, domainName, oldHostName, newHostName)

	resp := s.toRecordResponse(record)
	return &resp, nil
}

// ============================================================
// DELETE DNS RECORD
// ============================================================

// DeleteDNSRecord deletes a DNS record from the database and syncs delete to DNA API
func (s *DNSService) DeleteDNSRecord(userID uint, recordID uint) error {
	record, err := s.repo.GetRecordByIDAndUser(recordID, userID)
	if err != nil {
		return errors.New("DNS record not found")
	}

	// Delete from local DB
	if err := s.repo.DeleteRecord(recordID, userID); err != nil {
		return errors.New("failed to delete DNS record")
	}

	// Sync delete to DNA API (async)
	go s.syncDeleteToDNA(record.DomainName, record.HostName)

	return nil
}

// DeleteDNSByHost deletes DNS records by domain + host name and syncs to DNA API
func (s *DNSService) DeleteDNSByHost(userID uint, req DeleteDNSByHostRequest) error {
	domainName := strings.ToLower(strings.TrimSpace(req.DomainName))
	hostName := strings.TrimSpace(req.HostName)

	if domainName == "" || hostName == "" {
		return errors.New("domainName and hostName are required")
	}

	// Verify domain ownership
	_, err := s.repo.GetDomainByNameAndUser(domainName, userID)
	if err != nil {
		return fmt.Errorf("domain '%s' not found or you don't have access", domainName)
	}

	// Delete matching records from local DB
	records, _ := s.repo.GetRecordsByDomainName(userID, domainName)
	for _, rec := range records {
		if strings.EqualFold(rec.HostName, hostName) {
			_ = s.repo.DeleteRecord(rec.ID, userID)
		}
	}

	// Sync delete to DNA API (async)
	go s.syncDeleteToDNA(domainName, hostName)

	return nil
}

// ============================================================
// RESYNC DNS RECORD
// ============================================================

// ResyncDNSRecord re-syncs a failed DNS record to the DNA API
func (s *DNSService) ResyncDNSRecord(userID uint, recordID uint) (*DNSRecordResponse, error) {
	record, err := s.repo.GetRecordByIDAndUser(recordID, userID)
	if err != nil {
		return nil, errors.New("DNS record not found")
	}

	// Reset sync status
	record.SyncStatus = models.DNSSyncStatusPending
	record.SyncError = ""
	_ = s.repo.UpdateRecord(record)

	// Sync in background
	go s.syncSingleRecordToDNA(record)

	resp := s.toRecordResponse(record)
	return &resp, nil
}

// ============================================================
// DNA API SYNC (INTERNAL)
// ============================================================

// syncToDNA syncs DNS host records to the DNA reseller API
func (s *DNSService) syncToDNA(domainName string, hostName string, ipInputs []IPAddressInput, records []models.DNSRecord) {
	// Build DNA API IP list
	var dnaIPs []DNAIPAddressItem
	for _, ip := range ipInputs {
		ipVersion := strings.TrimSpace(ip.IPVersion)
		if ipVersion == "" {
			ipVersion = "IPv4"
		}
		dnaIPs = append(dnaIPs, DNAIPAddressItem{
			IPAddress: strings.TrimSpace(ip.IPAddress),
			IPVersion: ipVersion,
		})
	}

	// Call DNA API
	_, err := s.reseller.SyncDNSHost(domainName, hostName, dnaIPs)

	if err != nil {
		log.Printf("[DNS] ⚠️ DNA sync failed for %s/%s: %v", domainName, hostName, err)
		// Mark all records as failed
		for _, rec := range records {
			_ = s.repo.UpdateSyncStatus(rec.ID, models.DNSSyncStatusFailed, err.Error())
		}
		return
	}

	// Mark all records as synced
	for _, rec := range records {
		_ = s.repo.UpdateSyncStatus(rec.ID, models.DNSSyncStatusSynced, "")
	}
	log.Printf("[DNS] ✅ All %d DNS records synced for %s/%s", len(records), domainName, hostName)
}

// syncSingleRecordToDNA syncs a single DNS record to the DNA API via POST (create)
func (s *DNSService) syncSingleRecordToDNA(record *models.DNSRecord) {
	dnaIPs := []DNAIPAddressItem{
		{
			IPAddress: record.IPAddress,
			IPVersion: record.IPVersion,
		},
	}

	_, err := s.reseller.SyncDNSHost(record.DomainName, record.HostName, dnaIPs)

	if err != nil {
		log.Printf("[DNS] ⚠️ DNA sync failed for record #%d: %v", record.ID, err)
		_ = s.repo.UpdateSyncStatus(record.ID, models.DNSSyncStatusFailed, err.Error())
		return
	}

	_ = s.repo.UpdateSyncStatus(record.ID, models.DNSSyncStatusSynced, "")
	log.Printf("[DNS] ✅ DNS record #%d synced successfully", record.ID)
}

// syncUpdateToDNA syncs a DNS record update to the DNA API via PUT
func (s *DNSService) syncUpdateToDNA(record *models.DNSRecord, domainName string, oldHostName string, newHostName string) {
	dnaIPs := []DNAIPAddressItem{
		{
			IPAddress: record.IPAddress,
			IPVersion: record.IPVersion,
		},
	}

	_, err := s.reseller.UpdateDNSHost(domainName, oldHostName, newHostName, dnaIPs)

	if err != nil {
		log.Printf("[DNS] ⚠️ DNA PUT sync failed for record #%d: %v", record.ID, err)
		_ = s.repo.UpdateSyncStatus(record.ID, models.DNSSyncStatusFailed, err.Error())
		return
	}

	_ = s.repo.UpdateSyncStatus(record.ID, models.DNSSyncStatusSynced, "")
	log.Printf("[DNS] ✅ DNS record #%d updated via PUT successfully", record.ID)
}

// syncDeleteToDNA syncs a DNS record deletion to the DNA API via DELETE
func (s *DNSService) syncDeleteToDNA(domainName string, hostName string) {
	_, err := s.reseller.DeleteDNSHost(domainName, hostName)
	if err != nil {
		log.Printf("[DNS] ⚠️ DNA DELETE sync failed for %s/%s: %v", domainName, hostName, err)
		return
	}
	log.Printf("[DNS] ✅ DNS host %s/%s deleted via DELETE successfully", domainName, hostName)
}

// ============================================================
// RESPONSE BUILDER
// ============================================================

func (s *DNSService) toRecordResponse(r *models.DNSRecord) DNSRecordResponse {
	return DNSRecordResponse{
		ID:         r.ID,
		DomainID:   r.DomainID,
		DomainName: r.DomainName,
		HostName:   r.HostName,
		RecordType: string(r.RecordType),
		IPAddress:  r.IPAddress,
		IPVersion:  r.IPVersion,
		TTL:        r.TTL,
		Priority:   r.Priority,
		SyncStatus: string(r.SyncStatus),
		SyncError:  r.SyncError,
		CreatedAt:  r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  r.UpdatedAt.Format(time.RFC3339),
	}
}
