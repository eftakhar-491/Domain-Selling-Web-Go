package user_domain

import (
	"errors"
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

// GetDomainsByUserID retrieves domains owned by the user, dynamically synced from orders
func (r *DomainRepository) GetDomainsByUserID(userID uint, search string) ([]models.Domain, error) {
	// 1. Clean up legacy seeded starter domains that were never ordered by this user
	legacyStarters := []string{"orbitstudio.com", "nexora.io", "wearebengal.bd", "kinetic.xyz", "arifspace.dev"}
	for _, starterName := range legacyStarters {
		var orderCount int64
		r.db.Table("order_items").
			Joins("JOIN orders ON orders.id = order_items.order_id").
			Where("orders.user_id = ? AND LOWER(order_items.domain_name) = ?", userID, strings.ToLower(starterName)).
			Count(&orderCount)
		if orderCount == 0 {
			r.db.Where("user_id = ? AND LOWER(domain_name) = ?", userID, strings.ToLower(starterName)).
				Delete(&models.Domain{})
		}
	}

	// 2. Sync ordered domains from orders and order_items into models.Domain
	var orderItems []struct {
		DomainName    string
		TLD           string
		Period        int
		OrderStatus   models.OrderStatus
		PaymentStatus models.PaymentStatus
		CreatedAt     time.Time
	}

	r.db.Table("order_items").
		Select("order_items.domain_name, order_items.tld, order_items.period, orders.status as order_status, orders.payment_status, orders.created_at").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.user_id = ? AND orders.status NOT IN (?)", userID, []models.OrderStatus{models.OrderStatusCancelled, models.OrderStatusFailed}).
		Scan(&orderItems)

	for _, oi := range orderItems {
		dName := strings.ToLower(strings.TrimSpace(oi.DomainName))
		if dName == "" {
			continue
		}

		var existing models.Domain
		err := r.db.Where("user_id = ? AND LOWER(domain_name) = ?", userID, dName).First(&existing).Error
		if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
			status := models.DomainStatusPending
			if oi.OrderStatus == models.OrderStatusPaid || oi.OrderStatus == models.OrderStatusCompleted || oi.PaymentStatus == models.PaymentStatusPaid {
				status = models.DomainStatusActive
			}

			period := oi.Period
			if period < 1 {
				period = 1
			}
			regDate := oi.CreatedAt
			if regDate.IsZero() {
				regDate = time.Now()
			}
			expiresAt := regDate.AddDate(period, 0, 0)

			tld := oi.TLD
			if tld == "" {
				parts := strings.Split(dName, ".")
				if len(parts) > 1 {
					tld = strings.Join(parts[1:], ".")
				} else {
					tld = "com"
				}
			}

			newDomain := models.Domain{
				UserID:            userID,
				DomainName:        dName,
				TLD:               tld,
				Status:            status,
				AutoRenew:         true,
				RegistrationDate:  regDate,
				ExpiresAt:         expiresAt,
				PrivacyProtection: true,
				Nameservers:       "ns1.domain.bd,ns2.domain.bd",
			}
			_ = r.db.Create(&newDomain).Error
		} else if err == nil {
			if (oi.OrderStatus == models.OrderStatusPaid || oi.OrderStatus == models.OrderStatusCompleted || oi.PaymentStatus == models.PaymentStatusPaid) &&
				existing.Status == models.DomainStatusPending {
				existing.Status = models.DomainStatusActive
				_ = r.db.Save(&existing).Error
			}
		}
	}

	// 3. Query all domains owned by the user
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

// GetDomainByID fetches a domain by ID with owner user preloaded
func (r *DomainRepository) GetDomainByID(domainID uint) (*models.Domain, error) {
	var domain models.Domain
	err := r.db.Preload("User").Where("id = ?", domainID).First(&domain).Error
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

// GetAllDomainsAdmin fetches all domains across the platform with owner details
func (r *DomainRepository) GetAllDomainsAdmin(page, limit int, search string, status string) ([]models.Domain, int64, error) {
	var domains []models.Domain
	var total int64

	query := r.db.Model(&models.Domain{}).Joins("LEFT JOIN users ON users.id = domains.user_id")

	if strings.TrimSpace(search) != "" {
		s := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		query = query.Where("LOWER(domains.domain_name) LIKE ? OR LOWER(users.name) LIKE ? OR LOWER(users.email) LIKE ?", s, s, s)
	}

	if strings.TrimSpace(status) != "" {
		query = query.Where("domains.status = ?", strings.ToUpper(strings.TrimSpace(status)))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Preload("User").
		Order("domains.created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&domains).Error

	return domains, total, err
}

// UpdateDomainStatus updates the status and auto_renew of a domain
func (r *DomainRepository) UpdateDomainStatus(domainID uint, status string, autoRenew *bool) (*models.Domain, error) {
	var domain models.Domain
	if err := r.db.Preload("User").Where("id = ?", domainID).First(&domain).Error; err != nil {
		return nil, err
	}

	if status != "" {
		domain.Status = models.DomainStatus(strings.ToUpper(strings.TrimSpace(status)))
	}
	if autoRenew != nil {
		domain.AutoRenew = *autoRenew
	}

	if err := r.db.Save(&domain).Error; err != nil {
		return nil, err
	}

	return &domain, nil
}

