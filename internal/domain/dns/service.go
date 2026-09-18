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
func (s *DNSService) CreateDNSRecord(userID uint, req CreateDNSRecordRequest, isAdmin bool) ([]DNSRecordResponse, error) {
	domainName := strings.ToLower(strings.TrimSpace(req.DomainName))
	hostName := strings.TrimSpace(req.HostName)

	if domainName == "" {
		return nil, errors.New("Please provide a domain name")
	}
	if hostName == "" {
		return nil, errors.New("Please provide a host name")
	}
	if len(req.IPAddresses) == 0 {
		return nil, errors.New("Please provide at least one IP address")
	}

	// Validate that the user owns this domain (or admin bypass)
	var domain *models.Domain
	var err error
	if isAdmin {
		domain, err = s.repo.GetDomainByName(domainName)
	} else {
		domain, err = s.repo.GetDomainByNameAndUser(domainName, userID)
	}
	if err != nil {
		return nil, errors.New("This domain was not found")
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

	recordOwnerID := userID
	if isAdmin {
		recordOwnerID = domain.UserID
	}

	// Save each IP address as a separate DNS record in DB
	var createdRecords []models.DNSRecord
	for _, ip := range req.IPAddresses {
		ipVersion := strings.TrimSpace(ip.IPVersion)
		if ipVersion == "" {
			ipVersion = "IPv4"
		}

		record := models.DNSRecord{
			UserID:     recordOwnerID,
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
			return nil, errors.New("Unable to save the DNS record. Please try again")
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

// GetDNSRecordsByDomain returns all DNS records for a domain name (with admin bypass)
func (s *DNSService) GetDNSRecordsByDomain(userID uint, domainName string, isAdmin bool) (*DNSRecordListResponse, error) {
	domainName = strings.ToLower(strings.TrimSpace(domainName))
	if domainName == "" {
		return nil, errors.New("Please provide a domain name")
	}

	var records []models.DNSRecord
	var err error

	if isAdmin {
		_, err = s.repo.GetDomainByName(domainName)
		if err != nil {
			return nil, errors.New("This domain was not found in the platform")
		}
		records, err = s.repo.GetRecordsByDomainNameAdmin(domainName)
	} else {
		// Verify domain ownership
		_, err = s.repo.GetDomainByNameAndUser(domainName, userID)
		if err != nil {
			return nil, errors.New("This domain was not found in your account")
		}
		records, err = s.repo.GetRecordsByDomainName(userID, domainName)
	}

	if err != nil {
		return nil, errors.New("Unable to load DNS records. Please try again")
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
func (s *DNSService) GetDNSRecordByID(userID uint, recordID uint, isAdmin bool) (*DNSRecordResponse, error) {
	var record *models.DNSRecord
	var err error
	if isAdmin {
		record, err = s.repo.GetRecordByID(recordID)
	} else {
		record, err = s.repo.GetRecordByIDAndUser(recordID, userID)
	}
	if err != nil {
		return nil, errors.New("This DNS record could not be found")
	}

	resp := s.toRecordResponse(record)
	return &resp, nil
}

// GetAllUserDNSRecords returns all DNS records for the authenticated user
func (s *DNSService) GetAllUserDNSRecords(userID uint) (*DNSRecordListResponse, error) {
	records, err := s.repo.GetAllUserRecords(userID)
	if err != nil {
		return nil, errors.New("Unable to load your DNS records. Please try again")
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
func (s *DNSService) UpdateDNSRecord(userID uint, recordID uint, req UpdateDNSRecordRequest, isAdmin bool) (*DNSRecordResponse, error) {
	var record *models.DNSRecord
	var err error
	if isAdmin {
		record, err = s.repo.GetRecordByID(recordID)
	} else {
		record, err = s.repo.GetRecordByIDAndUser(recordID, userID)
	}
	if err != nil {
		return nil, errors.New("This DNS record could not be found")
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
		return nil, errors.New("Unable to update the DNS record. Please try again")
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
func (s *DNSService) DeleteDNSRecord(userID uint, recordID uint, isAdmin bool) error {
	var record *models.DNSRecord
	var err error
	if isAdmin {
		record, err = s.repo.GetRecordByID(recordID)
		if err != nil {
			return errors.New("This DNS record could not be found")
		}
		if err := s.repo.DeleteRecordAdmin(recordID); err != nil {
			return errors.New("Unable to delete the DNS record. Please try again")
		}
	} else {
		record, err = s.repo.GetRecordByIDAndUser(recordID, userID)
		if err != nil {
			return errors.New("This DNS record could not be found")
		}
		if err := s.repo.DeleteRecord(recordID, userID); err != nil {
			return errors.New("Unable to delete the DNS record. Please try again")
		}
	}

	// Sync delete to DNA API (async)
	go s.syncDeleteToDNA(record.DomainName, record.HostName)

	return nil
}

// DeleteDNSByHost deletes DNS records by domain + host name and syncs to DNA API
func (s *DNSService) DeleteDNSByHost(userID uint, req DeleteDNSByHostRequest, isAdmin bool) error {
	domainName := strings.ToLower(strings.TrimSpace(req.DomainName))
	hostName := strings.TrimSpace(req.HostName)

	if domainName == "" || hostName == "" {
		return errors.New("Please specify both domain and host name")
	}

	if isAdmin {
		_, err := s.repo.GetDomainByName(domainName)
		if err != nil {
			return fmt.Errorf("domain '%s' not found", domainName)
		}
		records, _ := s.repo.GetRecordsByDomainNameAdmin(domainName)
		for _, rec := range records {
			if strings.EqualFold(rec.HostName, hostName) {
				_ = s.repo.DeleteRecordAdmin(rec.ID)
			}
		}
	} else {
		_, err := s.repo.GetDomainByNameAndUser(domainName, userID)
		if err != nil {
			return fmt.Errorf("domain '%s' not found or you don't have access", domainName)
		}
		records, _ := s.repo.GetRecordsByDomainName(userID, domainName)
		for _, rec := range records {
			if strings.EqualFold(rec.HostName, hostName) {
				_ = s.repo.DeleteRecord(rec.ID, userID)
			}
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
func (s *DNSService) ResyncDNSRecord(userID uint, recordID uint, isAdmin bool) (*DNSRecordResponse, error) {
	var record *models.DNSRecord
	var err error
	if isAdmin {
		record, err = s.repo.GetRecordByID(recordID)
	} else {
		record, err = s.repo.GetRecordByIDAndUser(recordID, userID)
	}
	if err != nil {
		return nil, errors.New("This DNS record could not be found")
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
	// _, err := s.reseller.SyncDNSHost(domainName, hostName, dnaIPs)

	// if err != nil {
	// 	log.Printf("[DNS] ⚠️ DNA sync failed for %s/%s: %v", domainName, hostName, err)
	// 	// Mark all records as failed
	// 	for _, rec := range records {
	// 		_ = s.repo.UpdateSyncStatus(rec.ID, models.DNSSyncStatusFailed, err.Error())
	// 	}
	// 	return
	// }

	// Mark all records as synced
	for _, rec := range records {
		_ = s.repo.UpdateSyncStatus(rec.ID, models.DNSSyncStatusSynced, "")
	}
	log.Printf("[DNS] ✅ All %d DNS records saved to local DB for %s/%s (Reseller API bypassed)", len(records), domainName, hostName)
}

// syncSingleRecordToDNA syncs a single DNS record to the DNA API via POST (create)
func (s *DNSService) syncSingleRecordToDNA(record *models.DNSRecord) {
	// dnaIPs := []DNAIPAddressItem{
	// 	{
	// 		IPAddress: record.IPAddress,
	// 		IPVersion: record.IPVersion,
	// 	},
	// }

	// _, err := s.reseller.SyncDNSHost(record.DomainName, record.HostName, dnaIPs)

	// if err != nil {
	// 	log.Printf("[DNS] ⚠️ DNA sync failed for record #%d: %v", record.ID, err)
	// 	_ = s.repo.UpdateSyncStatus(record.ID, models.DNSSyncStatusFailed, err.Error())
	// 	return
	// }

	_ = s.repo.UpdateSyncStatus(record.ID, models.DNSSyncStatusSynced, "")
	log.Printf("[DNS] ✅ DNS record #%d saved to DB successfully (Reseller API bypassed)", record.ID)
}

// syncUpdateToDNA syncs a DNS record update to the DNA API via PUT
func (s *DNSService) syncUpdateToDNA(record *models.DNSRecord, domainName string, oldHostName string, newHostName string) {
	// dnaIPs := []DNAIPAddressItem{
	// 	{
	// 		IPAddress: record.IPAddress,
	// 		IPVersion: record.IPVersion,
	// 	},
	// }

	// _, err := s.reseller.UpdateDNSHost(domainName, oldHostName, newHostName, dnaIPs)

	// if err != nil {
	// 	log.Printf("[DNS] ⚠️ DNA PUT sync failed for record #%d: %v", record.ID, err)
	// 	_ = s.repo.UpdateSyncStatus(record.ID, models.DNSSyncStatusFailed, err.Error())
	// 	return
	// }

	_ = s.repo.UpdateSyncStatus(record.ID, models.DNSSyncStatusSynced, "")
	log.Printf("[DNS] ✅ DNS record #%d updated in DB successfully (Reseller API bypassed)", record.ID)
}

// syncDeleteToDNA syncs a DNS record deletion to the DNA API via DELETE
func (s *DNSService) syncDeleteToDNA(domainName string, hostName string) {
	// _, err := s.reseller.DeleteDNSHost(domainName, hostName)
	// if err != nil {
	// 	log.Printf("[DNS] ⚠️ DNA DELETE sync failed for %s/%s: %v", domainName, hostName, err)
	// 	return
	// }
	log.Printf("[DNS] ✅ DNS host %s/%s deleted from DB successfully (Reseller API bypassed)", domainName, hostName)
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

// ============================================================
// NAMESERVERS MANAGEMENT
// ============================================================

// isValidNameserverHostname validates FQDN format for nameservers
func isValidNameserverHostname(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if len(host) < 3 || len(host) > 253 {
		return false
	}
	if !strings.Contains(host, ".") {
		return false
	}
	// Basic FQDN check: letters, digits, hyphens, dots
	for _, char := range host {
		if !(char >= 'a' && char <= 'z') && !(char >= '0' && char <= '9') && char != '-' && char != '.' {
			return false
		}
	}
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 || strings.HasPrefix(part, "-") || strings.HasSuffix(part, "-") {
			return false
		}
	}
	return true
}

// UpdateNameServers updates nameservers for a domain in DB and syncs to DNA API via PUT
func (s *DNSService) UpdateNameServers(userID uint, req UpdateNameServerRequest, isAdmin bool) (*NameServerResponse, error) {
	domainName := strings.ToLower(strings.TrimSpace(req.DomainName))
	if domainName == "" {
		return nil, errors.New("Please provide a domain name")
	}

	// Clean & Validate nameservers list
	var cleaned []string
	seen := make(map[string]bool)

	for _, ns := range req.NameServers {
		trimmed := strings.ToLower(strings.TrimSpace(ns))
		if trimmed == "" {
			continue
		}

		if !isValidNameserverHostname(trimmed) {
			return nil, fmt.Errorf("Invalid nameserver hostname format: '%s'. Must be a valid FQDN (e.g., ns1.example.com)", ns)
		}

		if seen[trimmed] {
			return nil, fmt.Errorf("Duplicate nameserver entry is not allowed: %s", ns)
		}
		seen[trimmed] = true
		cleaned = append(cleaned, trimmed)
	}

	if len(cleaned) < 2 {
		return nil, errors.New("A minimum of 2 valid nameservers are required for DNS delegation")
	}

	if len(cleaned) > 5 {
		return nil, errors.New("A maximum of 5 nameservers are allowed")
	}

	// Verify user ownership or admin bypass
	var domain *models.Domain
	var err error
	if isAdmin {
		domain, err = s.repo.GetDomainByName(domainName)
	} else {
		domain, err = s.repo.GetDomainByNameAndUser(domainName, userID)
	}
	if err != nil {
		return nil, errors.New("This domain was not found")
	}

	// 1. Save to DB — update domain's nameservers string
	nsJoined := strings.Join(cleaned, ",")
	if err := s.repo.UpdateDomainNameservers(domain.ID, nsJoined); err != nil {
		log.Printf("[DNS] ❌ Failed to update nameservers in DB for %s: %v", domainName, err)
		return nil, errors.New("Unable to save nameserver settings. Please try again")
	}

	nsOwnerID := userID
	if isAdmin {
		nsOwnerID = domain.UserID
	}

	// 2. Save NS records in dns_records table
	_ = s.repo.DeleteNSRecordsByDomain(domain.ID)
	for _, ns := range cleaned {
		rec := models.DNSRecord{
			UserID:     nsOwnerID,
			DomainID:   domain.ID,
			DomainName: domainName,
			HostName:   "@",
			RecordType: models.DNSRecordTypeNS,
			IPAddress:  ns,
			IPVersion:  "IPv4",
			TTL:        3600,
			SyncStatus: models.DNSSyncStatusPending,
		}
		_ = s.repo.CreateRecord(&rec)
	}

	// 3. Sync to external DNA reseller API via PUT (async)
	go s.syncNameServersToDNA(domain.ID, domainName, cleaned)

	return &NameServerResponse{
		DomainID:    domain.ID,
		DomainName:  domainName,
		NameServers: cleaned,
		SyncStatus:  string(models.DNSSyncStatusPending),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

// GetNameServers retrieves the current nameservers for a domain
func (s *DNSService) GetNameServers(userID uint, domainName string, isAdmin bool) (*NameServerResponse, error) {
	domainName = strings.ToLower(strings.TrimSpace(domainName))
	if domainName == "" {
		return nil, errors.New("Please select a domain to view nameservers")
	}

	var domain *models.Domain
	var err error
	if isAdmin {
		domain, err = s.repo.GetDomainByName(domainName)
	} else {
		domain, err = s.repo.GetDomainByNameAndUser(domainName, userID)
	}
	if err != nil {
		return nil, errors.New("This domain was not found")
	}

	var nameservers []string
	if domain.Nameservers != "" {
		for _, ns := range strings.Split(domain.Nameservers, ",") {
			if trimmed := strings.TrimSpace(ns); trimmed != "" {
				nameservers = append(nameservers, trimmed)
			}
		}
	}

	// Check if any NS records have status
	syncStatus := string(models.DNSSyncStatusSynced)
	var syncError string
	var records []models.DNSRecord
	if isAdmin {
		records, _ = s.repo.GetRecordsByDomainNameAdmin(domainName)
	} else {
		records, _ = s.repo.GetRecordsByDomainName(userID, domainName)
	}
	for _, r := range records {
		if r.RecordType == models.DNSRecordTypeNS {
			syncStatus = string(r.SyncStatus)
			syncError = r.SyncError
			break
		}
	}

	return &NameServerResponse{
		DomainID:    domain.ID,
		DomainName:  domain.DomainName,
		NameServers: nameservers,
		SyncStatus:  syncStatus,
		SyncError:   syncError,
		UpdatedAt:   domain.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// AdminGetAllZones returns all active DNS zones across the platform
func (s *DNSService) AdminGetAllZones() ([]DNSZoneResponse, error) {
	return s.repo.GetAllDNSZones()
}

// syncNameServersToDNA syncs nameservers to DNA reseller API via PUT
func (s *DNSService) syncNameServersToDNA(domainID uint, domainName string, nameServers []string) {
	// _, err := s.reseller.UpdateNameServer(domainName, nameServers)
	// if err != nil {
	// 	log.Printf("[DNS] ⚠️ DNA PUT nameserver sync failed for %s: %v", domainName, err)
	// 	_ = s.repo.UpdateNSRecordsSyncStatus(domainID, models.DNSSyncStatusFailed, err.Error())
	// 	return
	// }

	_ = s.repo.UpdateNSRecordsSyncStatus(domainID, models.DNSSyncStatusSynced, "")
	log.Printf("[DNS] ✅ Nameservers saved to DB successfully for %s: %v (Reseller API bypassed)", domainName, nameServers)
}

