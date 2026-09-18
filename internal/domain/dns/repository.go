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
