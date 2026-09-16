package dns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"project-setup/internal/config"
)

// ============================================================
// DNA RESELLER API — DNS HOST MANAGEMENT
// ============================================================

// DNAResellerClient handles all outbound calls to the DNA Reseller API for DNS operations
type DNAResellerClient struct {
	client *http.Client
}

// NewDNAResellerClient creates a new DNAResellerClient
func NewDNAResellerClient() *DNAResellerClient {
	return &DNAResellerClient{
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

// ============================================================
// DNA API DTOs
// ============================================================

// DNADNSHostRequest is the body sent to POST /domains/dns/host
type DNADNSHostRequest struct {
	DomainName  string             `json:"domainName"`
	HostName    string             `json:"hostName"`
	IPAddresses []DNAIPAddressItem `json:"ipAddresses"`
}

// DNADNSHostUpdateRequest is the body sent to PUT /domains/dns/host
type DNADNSHostUpdateRequest struct {
	DomainName  string             `json:"domainName"`
	HostName    string             `json:"hostName"`
	NewHostName string             `json:"newHostName,omitempty"`
	IPAddresses []DNAIPAddressItem `json:"ipAddresses"`
}

// DNAIPAddressItem represents a single IP in the DNA API format
type DNAIPAddressItem struct {
	IPAddress string `json:"ipAddress"`
	IPVersion string `json:"ipVersion"`
}

// DNANameServerRequest is the body sent to PUT /domains/dns/name-server
type DNANameServerRequest struct {
	DomainName  string   `json:"domainName"`
	NameServers []string `json:"nameServers"`
}

// DNADNSHostResponse is the response from the DNA API
type DNADNSHostResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ============================================================
// SHARED HELPERS
// ============================================================

// dnaConfig holds the resolved DNA API configuration
type dnaConfig struct {
	BaseURL    string
	ResellerID string
	APIKey     string
}

// getDNAConfig loads and validates DNA API credentials from env
func getDNAConfig() (*dnaConfig, error) {
	env := config.GetEnv()

	baseURL := env.DNABaseURL
	if baseURL == "" {
		baseURL = "https://api.domainresellerapi.com/api/v1"
	}

	if env.DNAResellerID == "" || env.DNAAPIKey == "" {
		return nil, fmt.Errorf("DNA reseller credentials not configured (DNA_RESELLER_ID / DNA_API_KEY)")
	}

	return &dnaConfig{
		BaseURL:    baseURL,
		ResellerID: env.DNAResellerID,
		APIKey:     env.DNAAPIKey,
	}, nil
}

// setDNAHeaders sets the standard DNA API auth & content headers
func setDNAHeaders(req *http.Request, cfg *dnaConfig) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("__reseller", cfg.ResellerID)
	req.Header.Set("X-API-KEY", cfg.APIKey)
	req.Header.Set("DNA_RESELLER_ID", cfg.ResellerID)
	req.Header.Set("DNA_API_KEY", cfg.APIKey)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
}

// executeDNARequest sends a request and parses the DNA API response
func (rc *DNAResellerClient) executeDNARequest(req *http.Request) (*DNADNSHostResponse, error) {
	resp, err := rc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DNA API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read DNA API response: %w", err)
	}

	log.Printf("[DNS-Reseller] DNA API response (status %d): %s", resp.StatusCode, string(body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp DNADNSHostResponse
		_ = json.Unmarshal(body, &errResp)
		errMsg := errResp.Message
		if errMsg == "" {
			errMsg = string(body)
		}
		return nil, fmt.Errorf("DNA API error (HTTP %d): %s", resp.StatusCode, errMsg)
	}

	var result DNADNSHostResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse DNA API response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("DNA API operation failed: %s", result.Message)
	}

	return &result, nil
}

// ============================================================
// CREATE — POST /domains/dns/host
// ============================================================

// SyncDNSHost sends a DNS host record to the DNA reseller API (create)
func (rc *DNAResellerClient) SyncDNSHost(domainName string, hostName string, ipAddresses []DNAIPAddressItem) (*DNADNSHostResponse, error) {
	cfg, err := getDNAConfig()
	if err != nil {
		return nil, err
	}

	reqBody := DNADNSHostRequest{
		DomainName:  domainName,
		HostName:    hostName,
		IPAddresses: ipAddresses,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DNS host request: %w", err)
	}

	apiURL := cfg.BaseURL + "/domains/dns/host"
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create DNS host request: %w", err)
	}

	setDNAHeaders(req, cfg)

	log.Printf("[DNS-Reseller] POST DNS host: %s → %s (IPs: %d) → %s",
		domainName, hostName, len(ipAddresses), apiURL)
	log.Printf("[DNS-Reseller] Request body: %s", string(jsonData))

	result, err := rc.executeDNARequest(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DNS-Reseller] ✅ DNS host created successfully: %s → %s", domainName, hostName)
	return result, nil
}

// ============================================================
// UPDATE — PUT /domains/dns/host
// ============================================================

// UpdateDNSHost updates a DNS host record via the DNA reseller API
func (rc *DNAResellerClient) UpdateDNSHost(domainName string, hostName string, newHostName string, ipAddresses []DNAIPAddressItem) (*DNADNSHostResponse, error) {
	cfg, err := getDNAConfig()
	if err != nil {
		return nil, err
	}

	reqBody := DNADNSHostUpdateRequest{
		DomainName:  domainName,
		HostName:    hostName,
		NewHostName: newHostName,
		IPAddresses: ipAddresses,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DNS host update request: %w", err)
	}

	apiURL := cfg.BaseURL + "/domains/dns/host"
	req, err := http.NewRequest(http.MethodPut, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create DNS host update request: %w", err)
	}

	setDNAHeaders(req, cfg)

	log.Printf("[DNS-Reseller] PUT DNS host: %s → %s (newHost: %s, IPs: %d) → %s",
		domainName, hostName, newHostName, len(ipAddresses), apiURL)
	log.Printf("[DNS-Reseller] Request body: %s", string(jsonData))

	result, err := rc.executeDNARequest(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DNS-Reseller] ✅ DNS host updated successfully: %s → %s", domainName, hostName)
	return result, nil
}

// ============================================================
// DELETE — DELETE /domains/dns/host?DomainName=&HostName=
// ============================================================

// DeleteDNSHost deletes a DNS host record via the DNA reseller API
func (rc *DNAResellerClient) DeleteDNSHost(domainName string, hostName string) (*DNADNSHostResponse, error) {
	cfg, err := getDNAConfig()
	if err != nil {
		return nil, err
	}

	// Build URL with query parameters
	apiURL := cfg.BaseURL + "/domains/dns/host"
	params := url.Values{}
	params.Set("DomainName", domainName)
	params.Set("HostName", hostName)
	fullURL := apiURL + "?" + params.Encode()

	req, err := http.NewRequest(http.MethodDelete, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create DNS host delete request: %w", err)
	}

	setDNAHeaders(req, cfg)

	log.Printf("[DNS-Reseller] DELETE DNS host: %s → %s → %s", domainName, hostName, fullURL)

	result, err := rc.executeDNARequest(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DNS-Reseller] ✅ DNS host deleted successfully: %s → %s", domainName, hostName)
	return result, nil
}

// ============================================================
// UPDATE NAMESERVERS — PUT /domains/dns/name-server
// ============================================================

// UpdateNameServer updates nameservers for a domain via the DNA reseller API (PUT)
func (rc *DNAResellerClient) UpdateNameServer(domainName string, nameServers []string) (*DNADNSHostResponse, error) {
	cfg, err := getDNAConfig()
	if err != nil {
		return nil, err
	}

	reqBody := DNANameServerRequest{
		DomainName:  domainName,
		NameServers: nameServers,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal nameserver request: %w", err)
	}

	baseURL := cfg.BaseURL
	var apiURL string
	if strings.HasSuffix(baseURL, "/domains/dns/name-server") {
		apiURL = baseURL
	} else {
		apiURL = strings.TrimRight(baseURL, "/") + "/domains/dns/name-server"
	}

	req, err := http.NewRequest(http.MethodPut, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create nameserver update request: %w", err)
	}

	setDNAHeaders(req, cfg)

	log.Printf("[DNS-Reseller] PUT nameserver: %s → %v → %s", domainName, nameServers, apiURL)
	log.Printf("[DNS-Reseller] Request body: %s", string(jsonData))

	result, err := rc.executeDNARequest(req)
	if err != nil {
		return nil, err
	}

	log.Printf("[DNS-Reseller] ✅ Nameservers updated successfully for %s: %v", domainName, nameServers)
	return result, nil
}

