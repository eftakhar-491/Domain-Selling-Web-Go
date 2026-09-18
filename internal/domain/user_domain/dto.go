package user_domain

// DomainResponse represents a user domain returned to client
type DomainResponse struct {
	ID                uint     `json:"id"`
	DomainName        string   `json:"domain_name"`
	TLD               string   `json:"tld"`
	Status            string   `json:"status"`
	StatusLabel       string   `json:"status_label"`
	Tone              string   `json:"tone"` // "green" | "amber" | "red"
	AutoRenew         bool     `json:"auto_renew"`
	AutoRenewLabel    string   `json:"auto_renew_label"` // "On" | "Off"
	RegistrationDate  string   `json:"registration_date"`
	ExpiresAt         string   `json:"expires_at"`
	FormattedExpiry   string   `json:"formatted_expiry"`
	PrivacyProtection bool     `json:"privacy_protection"`
	Nameservers       []string `json:"nameservers"`
	DaysUntilExpiry   int      `json:"days_until_expiry"`
}

// DomainMetrics represents summary counters
type DomainMetrics struct {
	TotalActive   int    `json:"total_active"`
	RenewingSoon  int    `json:"renewing_soon"`
	AutoRenew     string `json:"auto_renew"` // e.g. "04 / 05"
	TotalDomains  int    `json:"total_domains"`
	AutoRenewNote string `json:"auto_renew_note"`
}

// DomainsListResponse holds the domains array and calculated metrics
type DomainsListResponse struct {
	Domains []DomainResponse `json:"domains"`
	Metrics DomainMetrics    `json:"metrics"`
}

// DomainOwnerInfo represents user ownership in admin domain views
type DomainOwnerInfo struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	PhoneNumber *string `json:"phone_number,omitempty"`
}

// AdminDomainResponse represents a domain with owner details for admin console
type AdminDomainResponse struct {
	ID                uint            `json:"id"`
	DomainName        string          `json:"domain_name"`
	TLD               string          `json:"tld"`
	Status            string          `json:"status"`
	StatusLabel       string          `json:"status_label"`
	Tone              string          `json:"tone"`
	AutoRenew         bool            `json:"auto_renew"`
	AutoRenewLabel    string          `json:"auto_renew_label"`
	RegistrationDate  string          `json:"registration_date"`
	ExpiresAt         string          `json:"expires_at"`
	FormattedExpiry   string          `json:"formatted_expiry"`
	PrivacyProtection bool            `json:"privacy_protection"`
	Nameservers       []string        `json:"nameservers"`
	DaysUntilExpiry   int             `json:"days_until_expiry"`
	User              DomainOwnerInfo `json:"user"`
	CreatedAt         string          `json:"created_at"`
	UpdatedAt         string          `json:"updated_at"`
}

// AdminDomainListResponse represents paginated list of domains across the platform
type AdminDomainListResponse struct {
	Domains    []AdminDomainResponse `json:"domains"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	TotalPages int                   `json:"total_pages"`
	Metrics    DomainMetrics         `json:"metrics"`
}

// AdminUpdateDomainStatusRequest represents request to change a domain's status or settings
type AdminUpdateDomainStatusRequest struct {
	Status    string `json:"status" validate:"omitempty,oneof=ACTIVE EXPIRING_SOON EXPIRED PENDING SUSPENDED"`
	AutoRenew *bool  `json:"auto_renew,omitempty"`
}

