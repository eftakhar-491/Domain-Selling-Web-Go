package dns

// ============================================================
// REQUEST DTOs
// ============================================================

// IPAddressInput represents a single IP address entry from the client
type IPAddressInput struct {
	IPAddress string `json:"ipAddress" validate:"required"`
	IPVersion string `json:"ipVersion" validate:"required,oneof=IPv4 IPv6"`
}

// CreateDNSRecordRequest is the payload for creating/adding a DNS host record
type CreateDNSRecordRequest struct {
	DomainName  string           `json:"domainName" validate:"required"`
	HostName    string           `json:"hostName" validate:"required"`
	RecordType  string           `json:"recordType"`
	TTL         int              `json:"ttl"`
	Priority    int              `json:"priority"`
	IPAddresses []IPAddressInput `json:"ipAddresses" validate:"required,min=1,dive"`
}

// UpdateDNSRecordRequest is the payload for updating an existing DNS record
type UpdateDNSRecordRequest struct {
	DomainName  string           `json:"domainName"`
	HostName    *string          `json:"hostName"`
	NewHostName *string          `json:"newHostName"`
	RecordType  *string          `json:"recordType"`
	TTL         *int             `json:"ttl"`
	Priority    *int             `json:"priority"`
	IPAddresses []IPAddressInput `json:"ipAddresses"`
}

// DeleteDNSByHostRequest is used for deleting DNS records by domain + host name
// (syncs the delete to the external DNA API)
type DeleteDNSByHostRequest struct {
	DomainName string `json:"domainName" validate:"required"`
	HostName   string `json:"hostName" validate:"required"`
}

// ============================================================
// RESPONSE DTOs
// ============================================================

// DNSRecordResponse represents a single DNS record returned to the client
type DNSRecordResponse struct {
	ID         uint   `json:"id"`
	DomainID   uint   `json:"domain_id"`
	DomainName string `json:"domain_name"`
	HostName   string `json:"host_name"`
	RecordType string `json:"record_type"`
	IPAddress  string `json:"ip_address"`
	IPVersion  string `json:"ip_version"`
	TTL        int    `json:"ttl"`
	Priority   int    `json:"priority"`
	SyncStatus string `json:"sync_status"`
	SyncError  string `json:"sync_error,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// DNSRecordListResponse is a paginated list of DNS records for a domain
type DNSRecordListResponse struct {
	Records    []DNSRecordResponse `json:"records"`
	DomainName string              `json:"domain_name"`
	TotalCount int64               `json:"total_count"`
}
