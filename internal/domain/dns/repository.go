package dns

import (
	"strings"

	"project-setup/internal/models"

	"gorm.io/gorm"
)

// DNSRepository handles all database operations for DNS records
type DNSRepository struct {
	db *gorm.DB
}

// NewDNSRepository creates a new DNSRepository
func NewDNSRepository(db *gorm.DB) *DNSRepository {
	return &DNSRepository{db: db}
}

// CreateRecord persists a new DNS record
func (r *DNSRepository) CreateRecord(record *models.DNSRecord) error {
	return r.db.Create(record).Error
}

// GetRecordByID fetches a DNS record by its primary key
func (r *DNSRepository) GetRecordByID(id uint) (*models.DNSRecord, error) {
	var record models.DNSRecord
	err := r.db.First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetRecordByIDAndUser fetches a DNS record only if it belongs to the given user
func (r *DNSRepository) GetRecordByIDAndUser(id uint, userID uint) (*models.DNSRecord, error) {
	var record models.DNSRecord
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetRecordsByDomainName retrieves all DNS records for a specific domain name, scoped to user
func (r *DNSRepository) GetRecordsByDomainName(userID uint, domainName string) ([]models.DNSRecord, error) {
	var records []models.DNSRecord
	err := r.db.Where("user_id = ? AND domain_name = ?", userID, domainName).
		Order("created_at ASC").
		Find(&records).Error
	return records, err
}

// GetRecordsByDomainID retrieves all DNS records for a specific domain ID, scoped to user
func (r *DNSRepository) GetRecordsByDomainID(userID uint, domainID uint) ([]models.DNSRecord, error) {
	var records []models.DNSRecord
	err := r.db.Where("user_id = ? AND domain_id = ?", userID, domainID).
		Order("created_at ASC").
		Find(&records).Error
	return records, err
}

// GetAllUserRecords retrieves all DNS records owned by a user
func (r *DNSRepository) GetAllUserRecords(userID uint) ([]models.DNSRecord, error) {
	var records []models.DNSRecord
	err := r.db.Where("user_id = ?", userID).
		Order("domain_name ASC, created_at ASC").
		Find(&records).Error
	return records, err
}

// UpdateRecord saves all changes to an existing DNS record
func (r *DNSRepository) UpdateRecord(record *models.DNSRecord) error {
	return r.db.Save(record).Error
}

// UpdateSyncStatus atomically updates the sync status and error message
func (r *DNSRepository) UpdateSyncStatus(id uint, status models.DNSSyncStatus, syncError string) error {
	return r.db.Model(&models.DNSRecord{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"sync_status": status,
			"sync_error":  syncError,
		}).Error
}

// DeleteRecord deletes a DNS record scoped to the user
func (r *DNSRepository) DeleteRecord(id uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.DNSRecord{}).Error
}

// DeleteRecordsByDomainName deletes all DNS records for a domain, scoped to user
func (r *DNSRepository) DeleteRecordsByDomainName(userID uint, domainName string) error {
	return r.db.Where("user_id = ? AND domain_name = ?", userID, domainName).
		Delete(&models.DNSRecord{}).Error
}

// GetDomainByNameAndUser finds a domain record by name and user (for validation)
func (r *DNSRepository) GetDomainByNameAndUser(domainName string, userID uint) (*models.Domain, error) {
	var domain models.Domain
	err := r.db.Where("LOWER(domain_name) = LOWER(?) AND user_id = ?", strings.ToLower(strings.TrimSpace(domainName)), userID).First(&domain).Error
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

// GetDomainByName finds a domain record by name across all users (for admin operations)
func (r *DNSRepository) GetDomainByName(domainName string) (*models.Domain, error) {
	var domain models.Domain
	err := r.db.Where("LOWER(domain_name) = LOWER(?)", strings.ToLower(strings.TrimSpace(domainName))).First(&domain).Error
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

// GetRecordsByDomainNameAdmin gets all DNS records for a domain across all users
func (r *DNSRepository) GetRecordsByDomainNameAdmin(domainName string) ([]models.DNSRecord, error) {
	var records []models.DNSRecord
	err := r.db.Where("LOWER(domain_name) = LOWER(?)", strings.ToLower(strings.TrimSpace(domainName))).
		Order("created_at ASC").
		Find(&records).Error
	return records, err
}

// DeleteRecordsByDomainNameAdmin deletes all DNS records for a domain without user scope
func (r *DNSRepository) DeleteRecordsByDomainNameAdmin(domainName string) error {
	return r.db.Where("LOWER(domain_name) = LOWER(?)", strings.ToLower(strings.TrimSpace(domainName))).
		Delete(&models.DNSRecord{}).Error
}

// DeleteRecordAdmin deletes a DNS record by ID without user scope
func (r *DNSRepository) DeleteRecordAdmin(id uint) error {
	return r.db.Delete(&models.DNSRecord{}, id).Error
}

// GetAllDNSZones returns all domains in the system with their DNS record counts and sync status
func (r *DNSRepository) GetAllDNSZones() ([]DNSZoneResponse, error) {
	var domains []models.Domain
	if err := r.db.Order("domain_name ASC").Find(&domains).Error; err != nil {
		return nil, err
	}

	var zones []DNSZoneResponse
	for _, d := range domains {
		var count int64
		r.db.Model(&models.DNSRecord{}).Where("domain_id = ?", d.ID).Count(&count)

		var failedCount int64
		r.db.Model(&models.DNSRecord{}).Where("domain_id = ? AND sync_status = ?", d.ID, models.DNSSyncStatusFailed).Count(&failedCount)

		status := "Synced"
		if failedCount > 0 {
			status = "Failed"
		} else if count == 0 {
			status = "No records"
		}

		zones = append(zones, DNSZoneResponse{
			DomainName:  d.DomainName,
			UserID:      d.UserID,
			RecordCount: int(count),
			SyncStatus:  status,
			Nameservers: d.Nameservers,
			UpdatedAt:   d.UpdatedAt.Format("Jan 02, 2006"),
		})
	}
	return zones, nil
}


// UpdateDomainNameservers updates the nameservers string on the Domain model
func (r *DNSRepository) UpdateDomainNameservers(domainID uint, nameservers string) error {
	return r.db.Model(&models.Domain{}).Where("id = ?", domainID).Update("nameservers", nameservers).Error
}

// DeleteNSRecordsByDomain deletes all NS records for a domain
func (r *DNSRepository) DeleteNSRecordsByDomain(domainID uint) error {
	return r.db.Where("domain_id = ? AND record_type = ?", domainID, models.DNSRecordTypeNS).Delete(&models.DNSRecord{}).Error
}

// UpdateNSRecordsSyncStatus updates sync status for all NS records of a domain
func (r *DNSRepository) UpdateNSRecordsSyncStatus(domainID uint, status models.DNSSyncStatus, syncError string) error {
	return r.db.Model(&models.DNSRecord{}).
		Where("domain_id = ? AND record_type = ?", domainID, models.DNSRecordTypeNS).
		Updates(map[string]interface{}{
			"sync_status": status,
			"sync_error":  syncError,
		}).Error
}
