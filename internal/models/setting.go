package models

import (
	"gorm.io/gorm"
)

// SystemSetting stores configurable platform-wide settings
type SystemSetting struct {
	gorm.Model
	SiteName           string `gorm:"type:varchar(255);default:'Domain.BD - Digital Identity Registry'" json:"site_name"`
	SupportEmail       string `gorm:"type:varchar(255);default:'support@domain.bd'" json:"support_email"`
	Currency           string `gorm:"type:varchar(10);default:'BDT'" json:"currency"`
	RegistrarProvider  string `gorm:"type:varchar(50);default:'BTCL'" json:"registrar_provider"` // BTCL, ResellerClub, Namecheap, Demo
	RegistrarEndpoint  string `gorm:"type:varchar(500);default:'https://api.btcl.com.bd/v2/epp'" json:"registrar_endpoint"`
	EppClientID        string `gorm:"type:varchar(100);default:'DOMAINBD-REG-01'" json:"epp_client_id"`
	EppSecretKey       string `gorm:"type:varchar(255);default:''" json:"epp_secret_key"`
	DefaultNameservers string `gorm:"type:varchar(500);default:'ns1.domain.bd, ns2.domain.bd'" json:"default_nameservers"`
	DefaultTTL         int    `gorm:"default:3600" json:"default_ttl"`
	AutoRenewGraceDays int    `gorm:"default:30" json:"auto_renew_grace_days"`
	EmailAlertsEnabled bool   `gorm:"default:true" json:"email_alerts_enabled"`
	WebhookURL         string `gorm:"type:varchar(500);default:''" json:"webhook_url"`
}
