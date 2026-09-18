package user

import (
	"project-setup/internal/models"

	"gorm.io/gorm"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID finds a user by their ID
func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindAll returns a paginated list of users with optional search
func (r *UserRepository) FindAll(page, limit int, search string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	if search != "" {
		s := "%" + search + "%"
		query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?)", s, s)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Fetch paginated results
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// GetAdminStats computes aggregated platform metrics
func (r *UserRepository) GetAdminStats() (*AdminStatsResponse, error) {
	var stats AdminStatsResponse

	// User metrics
	r.db.Model(&models.User{}).Count(&stats.TotalUsers)
	r.db.Model(&models.User{}).Where("is_active = ?", true).Count(&stats.ActiveUsers)

	// Domain metrics
	r.db.Model(&models.Domain{}).Count(&stats.TotalDomains)
	r.db.Model(&models.Domain{}).Where("status = ?", models.DomainStatusActive).Count(&stats.ActiveDomains)

	// Order metrics
	r.db.Model(&models.Order{}).Count(&stats.TotalOrders)
	r.db.Model(&models.Order{}).Where("payment_status = ?", models.PaymentStatusPaid).Count(&stats.PaidOrders)
	r.db.Model(&models.Order{}).Where("payment_status = ?", models.PaymentStatusPending).Count(&stats.PendingOrders)

	// Revenue: sum of total_amount where payment_status is PAID
	var rev struct {
		Total float64
	}
	r.db.Model(&models.Order{}).
		Select("COALESCE(SUM(total_amount), 0) as total").
		Where("payment_status = ?", models.PaymentStatusPaid).
		Scan(&rev)
	stats.TotalRevenue = rev.Total

	// Discount metrics
	r.db.Model(&models.Discount{}).Count(&stats.TotalDiscounts)

	return &stats, nil
}

// Update saves changes to an existing user
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete removes a user by ID (hard delete)
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// GetSystemSettings retrieves the active system settings or creates defaults
func (r *UserRepository) GetSystemSettings() (*models.SystemSetting, error) {
	var settings models.SystemSetting
	err := r.db.First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			settings = models.SystemSetting{
				SiteName:           "Domain.BD - Digital Identity Registry",
				SupportEmail:       "support@domain.bd",
				Currency:           "BDT",
				RegistrarProvider:  "BTCL",
				RegistrarEndpoint:  "https://api.btcl.com.bd/v2/epp",
				EppClientID:        "DOMAINBD-REG-01",
				DefaultNameservers: "ns1.domain.bd, ns2.domain.bd",
				DefaultTTL:         3600,
				AutoRenewGraceDays: 30,
				EmailAlertsEnabled: true,
			}
			if createErr := r.db.Create(&settings).Error; createErr != nil {
				return nil, createErr
			}
			return &settings, nil
		}
		return nil, err
	}
	return &settings, nil
}

// UpdateSystemSettings updates the platform settings in database
func (r *UserRepository) UpdateSystemSettings(req SystemSettingRequest) (*models.SystemSetting, error) {
	settings, err := r.GetSystemSettings()
	if err != nil {
		return nil, err
	}

	if req.SiteName != "" {
		settings.SiteName = req.SiteName
	}
	if req.SupportEmail != "" {
		settings.SupportEmail = req.SupportEmail
	}
	if req.Currency != "" {
		settings.Currency = req.Currency
	}
	if req.RegistrarProvider != "" {
		settings.RegistrarProvider = req.RegistrarProvider
	}
	if req.RegistrarEndpoint != "" {
		settings.RegistrarEndpoint = req.RegistrarEndpoint
	}
	if req.EppClientID != "" {
		settings.EppClientID = req.EppClientID
	}
	if req.EppSecretKey != "" {
		settings.EppSecretKey = req.EppSecretKey
	}
	if req.DefaultNameservers != "" {
		settings.DefaultNameservers = req.DefaultNameservers
	}
	if req.DefaultTTL > 0 {
		settings.DefaultTTL = req.DefaultTTL
	}
	if req.AutoRenewGraceDays > 0 {
		settings.AutoRenewGraceDays = req.AutoRenewGraceDays
	}
	if req.EmailAlertsEnabled != nil {
		settings.EmailAlertsEnabled = *req.EmailAlertsEnabled
	}
	if req.WebhookURL != "" {
		settings.WebhookURL = req.WebhookURL
	}

	if err := r.db.Save(settings).Error; err != nil {
		return nil, err
	}

	return settings, nil
}
