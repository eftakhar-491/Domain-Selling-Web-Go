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

// UpdateDomainRequest for editing domain settings
type UpdateDomainRequest struct {
	AutoRenew         *bool   `json:"auto_renew"`
	Status            *string `json:"status"`
	PrivacyProtection *bool   `json:"privacy_protection"`
	Nameservers       *string `json:"nameservers"`
}

// CreateDomainRequest for manually adding / registering a domain
type CreateDomainRequest struct {
	DomainName        string `json:"domain_name" validate:"required"`
	Period            int    `json:"period"`
	AutoRenew         bool   `json:"auto_renew"`
	PrivacyProtection bool   `json:"privacy_protection"`
	Nameservers       string `json:"nameservers"`
}
