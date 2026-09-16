package user_domain

import (
	"strings"
	"time"

	"project-setup/internal/models"

	"gorm.io/gorm"
)

type DomainRepository struct {
	db *gorm.DB
}

func NewDomainRepository(db *gorm.DB) *DomainRepository {
	return &DomainRepository{db: db}
}

// GetDomainsByUserID retrieves all domains owned by the user, with optional search query
func (r *DomainRepository) GetDomainsByUserID(userID uint, search string) ([]models.Domain, error) {
	// If the user has no domains, seed starter domains automatically
	var count int64
	r.db.Model(&models.Domain{}).Where("user_id = ?", userID).Count(&count)
	if count == 0 {
		_ = r.SeedStarterDomains(userID)
	}

	query := r.db.Where("user_id = ?", userID)
	if strings.TrimSpace(search) != "" {
		searchTerm := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		query = query.Where("LOWER(domain_name) LIKE ?", searchTerm)
	}

	var domains []models.Domain
	err := query.Order("created_at ASC").Find(&domains).Error
	return domains, err
}

// GetDomainByIDAndUser fetches a single domain owned by a specific user
func (r *DomainRepository) GetDomainByIDAndUser(domainID uint, userID uint) (*models.Domain, error) {
	var domain models.Domain
	err := r.db.Where("id = ? AND user_id = ?", domainID, userID).First(&domain).Error
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

// CreateDomain creates a new domain entry in the database
func (r *DomainRepository) CreateDomain(domain *models.Domain) error {
	return r.db.Create(domain).Error
}

// UpdateDomain updates an existing domain record
func (r *DomainRepository) UpdateDomain(domain *models.Domain) error {
	return r.db.Save(domain).Error
}

// DeleteDomain deletes a domain record
func (r *DomainRepository) DeleteDomain(domainID uint, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", domainID, userID).Delete(&models.Domain{}).Error
}

// SeedStarterDomains populates realistic starter domains for a user if they have none
func (r *DomainRepository) SeedStarterDomains(userID uint) error {
	now := time.Now()
	starterDomains := []models.Domain{
		{
			UserID:            userID,
			DomainName:        "orbitstudio.com",
			TLD:               "com",
			Status:            models.DomainStatusActive,
			AutoRenew:         true,
			RegistrationDate:  now.AddDate(-1, 2, 0),
			ExpiresAt:         now.AddDate(1, 1, 9),
			PrivacyProtection: true,
			Nameservers:       "ns1.domain.bd,ns2.domain.bd",
		},
		{
			UserID:            userID,
			DomainName:        "nexora.io",
			TLD:               "io",
			Status:            models.DomainStatusActive,
			AutoRenew:         true,
			RegistrationDate:  now.AddDate(-1, 0, 0),
			ExpiresAt:         now.AddDate(0, 8, 24),
			PrivacyProtection: true,
			Nameservers:       "ns1.domain.bd,ns2.domain.bd",
		},
		{
			UserID:            userID,
			DomainName:        "wearebengal.bd",
			TLD:               "bd",
			Status:            models.DomainStatusExpiringSoon,
			AutoRenew:         false,
			RegistrationDate:  now.AddDate(-1, -1, 0),
			ExpiresAt:         now.AddDate(0, 0, 18),
			PrivacyProtection: true,
			Nameservers:       "ns1.btcl.gov.bd,ns2.btcl.gov.bd",
		},
		{
			UserID:            userID,
			DomainName:        "kinetic.xyz",
			TLD:               "xyz",
			Status:            models.DomainStatusActive,
			AutoRenew:         true,
			RegistrationDate:  now.AddDate(0, -2, 0),
			ExpiresAt:         now.AddDate(0, 10, 2),
			PrivacyProtection: true,
			Nameservers:       "ns1.domain.bd,ns2.domain.bd",
		},
		{
			UserID:            userID,
			DomainName:        "arifspace.dev",
			TLD:               "dev",
			Status:            models.DomainStatusActive,
			AutoRenew:         true,
			RegistrationDate:  now.AddDate(0, -6, 0),
			ExpiresAt:         now.AddDate(1, 6, 12),
			PrivacyProtection: true,
			Nameservers:       "ns1.domain.bd,ns2.domain.bd",
		},
	}

	for _, d := range starterDomains {
		// Verify not already exists to prevent duplicate error
		var existing models.Domain
		if err := r.db.Where("domain_name = ?", d.DomainName).First(&existing).Error; err != nil {
			_ = r.db.Create(&d).Error
		}
	}

	return nil
}
