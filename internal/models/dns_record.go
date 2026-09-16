package models

import "gorm.io/gorm"

// DNSRecordType represents the type of DNS record
type DNSRecordType string

const (
	DNSRecordTypeA     DNSRecordType = "A"
	DNSRecordTypeAAAA  DNSRecordType = "AAAA"
	DNSRecordTypeCNAME DNSRecordType = "CNAME"
	DNSRecordTypeMX    DNSRecordType = "MX"
	DNSRecordTypeTXT   DNSRecordType = "TXT"
	DNSRecordTypeNS    DNSRecordType = "NS"
	DNSRecordTypeSRV   DNSRecordType = "SRV"
)

// DNSSyncStatus represents the sync state with the external reseller API
type DNSSyncStatus string

const (
	DNSSyncStatusPending DNSSyncStatus = "PENDING"
	DNSSyncStatusSynced  DNSSyncStatus = "SYNCED"
	DNSSyncStatusFailed  DNSSyncStatus = "FAILED"
)

// DNSRecord represents a DNS host record stored locally and synced with the DNA reseller API
type DNSRecord struct {
	gorm.Model
	UserID     uint          `json:"user_id" gorm:"index;not null"`
	User       User          `json:"-" gorm:"foreignKey:UserID"`
	DomainID   uint          `json:"domain_id" gorm:"index;not null"`
	Domain     Domain        `json:"-" gorm:"foreignKey:DomainID"`
	DomainName string        `json:"domain_name" gorm:"type:varchar(255);not null;index"`
	HostName   string        `json:"host_name" gorm:"type:varchar(255);not null"`
	RecordType DNSRecordType `json:"record_type" gorm:"type:varchar(10);default:'A';not null"`
	IPAddress  string        `json:"ip_address" gorm:"type:varchar(255);not null"`
	IPVersion  string        `json:"ip_version" gorm:"type:varchar(10);default:'IPv4';not null"`
	TTL        int           `json:"ttl" gorm:"default:3600;not null"`
	Priority   int           `json:"priority" gorm:"default:0"`
	SyncStatus DNSSyncStatus `json:"sync_status" gorm:"type:varchar(20);default:'PENDING';not null;index"`
	SyncError  string        `json:"sync_error,omitempty" gorm:"type:text"`
}
