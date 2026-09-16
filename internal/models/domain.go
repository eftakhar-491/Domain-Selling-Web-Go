package models

import (
	"time"

	"gorm.io/gorm"
)

type DomainStatus string

const (
	DomainStatusActive       DomainStatus = "ACTIVE"
	DomainStatusExpiringSoon DomainStatus = "EXPIRING_SOON"
	DomainStatusExpired      DomainStatus = "EXPIRED"
	DomainStatusPending      DomainStatus = "PENDING"
	DomainStatusSuspended    DomainStatus = "SUSPENDED"
)

// Domain represents a registered domain owned by a user
type Domain struct {
	gorm.Model
	UserID            uint         `json:"user_id" gorm:"index;not null"`
	User              User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	DomainName        string       `json:"domain_name" gorm:"type:varchar(255);uniqueIndex;not null"`
	TLD               string       `json:"tld" gorm:"type:varchar(20);not null"`
	Status            DomainStatus `json:"status" gorm:"type:varchar(30);default:'ACTIVE';not null;index"`
	AutoRenew         bool         `json:"auto_renew" gorm:"default:true;not null"`
	RegistrationDate  time.Time    `json:"registration_date"`
	ExpiresAt         time.Time    `json:"expires_at" gorm:"index;not null"`
	PrivacyProtection bool         `json:"privacy_protection" gorm:"default:true;not null"`
	Nameservers       string       `json:"nameservers" gorm:"type:text;default:'ns1.domain.bd,ns2.domain.bd'"`
	ResellerDomainID  string       `json:"reseller_domain_id,omitempty" gorm:"type:varchar(255);index"`
	ResellerStatus    string       `json:"reseller_status,omitempty" gorm:"type:varchar(50);default:'PENDING'"`
}
