package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"project-setup/internal/config"
	"project-setup/internal/models"

	"gorm.io/gorm"
)

// ============================================================
// DNA RESELLER API — DOMAIN REGISTRATION
// ============================================================

// ResellerService handles domain registration via DomainNameAPI (DNA)
type ResellerService struct {
	client *http.Client
	db     *gorm.DB
}

// NewResellerService creates a new ResellerService
func NewResellerService(db *gorm.DB) *ResellerService {
	return &ResellerService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		db: db,
	}
}

// ============================================================
// REQUEST / RESPONSE DTOs
// ============================================================

// DNARegisterRequest is the payload sent to POST /domains/register
type DNARegisterRequest struct {
	DomainName    string             `json:"domainName"`
	Period        int                `json:"period"`
	Nameservers   []string           `json:"nameservers"`
	Handles       []DNAContactHandle `json:"handles"`
	TLDAttributes map[string]string  `json:"tldAttributes,omitempty"`
}

// DNAContactHandle represents a contact handle for domain registration
type DNAContactHandle struct {
	ContactType string `json:"contactType"`
	Handle      string `json:"handle"`
}

// DNARegisterResponse is the success response from the DNA API
type DNARegisterResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	DomainID       string `json:"domainId"`
	Status         string `json:"status"`
	ExpirationDate string `json:"expirationDate"`
}

// DNAErrorResponse is the error response from the DNA API
type DNAErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ============================================================
// REGISTER DOMAIN VIA DNA API
// ============================================================

// RegisterDomain calls the DNA /domains/register endpoint for a single domain
func (rs *ResellerService) RegisterDomain(domainName string, period int) (*DNARegisterResponse, error) {
	env := config.GetEnv()

	baseURL := env.DNABaseURL
	if baseURL == "" {
		baseURL = "https://api.domainresellerapi.com/api/v1"
	}
	resellerID := env.DNAResellerID
	apiKey := env.DNAAPIKey

	if resellerID == "" || apiKey == "" {
		return nil, fmt.Errorf("DNA reseller credentials not configured (DNA_RESELLER_ID / DNA_API_KEY)")
	}

	// Build request body
	reqBody := DNARegisterRequest{
		DomainName: domainName,
		Period:     period,
		Nameservers: []string{
			"ns1.domain.bd",
			"ns2.domain.bd",
		},
		Handles: []DNAContactHandle{
			{ContactType: "registrant", Handle: "default"},
			{ContactType: "admin", Handle: "default"},
			{ContactType: "tech", Handle: "default"},
			{ContactType: "billing", Handle: "default"},
		},
		TLDAttributes: map[string]string{},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal register request: %w", err)
	}

	// Build HTTP request
	url := baseURL + "/domains/register"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create register request: %w", err)
	}

	// Set headers (same pattern as find_domain service)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("__reseller", resellerID)
	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	log.Printf("[Reseller] Registering domain: %s (period: %d years) → %s", domainName, period, url)

	// Execute request
	resp, err := rs.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DNA API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read DNA API response: %w", err)
	}

	log.Printf("[Reseller] DNA API response (status %d): %s", resp.StatusCode, string(body))

	// Parse response
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp DNAErrorResponse
		_ = json.Unmarshal(body, &errResp)
		errMsg := errResp.Message
		if errMsg == "" {
			errMsg = string(body)
		}
		return nil, fmt.Errorf("DNA API error (HTTP %d): %s", resp.StatusCode, errMsg)
	}

	var result DNARegisterResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse DNA API response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("DNA registration failed: %s", result.Message)
	}

	return &result, nil
}

// ============================================================
// REGISTER ALL DOMAINS IN AN ORDER + SAVE TO DB
// ============================================================

// RegisterOrderDomains registers all domain items in an order via the DNA API,
// then creates corresponding Domain records in the database.
// This is called after payment is confirmed.
func (rs *ResellerService) RegisterOrderDomains(order *models.Order) {
	for _, item := range order.Items {
		go rs.registerAndSaveDomain(order.UserID, item)
	}
}

// registerAndSaveDomain registers a single domain via DNA API and saves the result to DB
func (rs *ResellerService) registerAndSaveDomain(userID uint, item models.OrderItem) {
	domainName := strings.ToLower(strings.TrimSpace(item.DomainName))
	period := item.Period
	if period < 1 {
		period = 1
	}

	log.Printf("[Reseller] Starting local domain provisioning for: %s (user: %d)", domainName, userID)

	// =========================================================================
	// NOTE: External DNA API registration call is commented out as requested.
	// Uncomment the block below when you are ready to activate external reseller API:
	// =========================================================================
	// result, err := rs.RegisterDomain(domainName, period)
	// =========================================================================

	now := time.Now()
	resellerStatus := "LOCAL_ACTIVE"
	resellerDomainID := ""
	expiresAt := now.AddDate(period, 0, 0)
	domainStatus := models.DomainStatusActive

	log.Printf("[Reseller] 💾 Saving domain to database: %s (Status: %s, Expires: %s)",
		domainName, domainStatus, expiresAt.Format("2006-01-02"))

	// Extract TLD
	parts := strings.Split(domainName, ".")
	tld := "com"
	if len(parts) > 1 {
		tld = strings.Join(parts[1:], ".")
	}

	// Check if domain already exists in DB
	var existing models.Domain
	if rs.db.Where("domain_name = ?", domainName).First(&existing).Error == nil {
		// Domain already exists — update with reseller info
		rs.db.Model(&existing).Updates(map[string]interface{}{
			"reseller_domain_id": resellerDomainID,
			"reseller_status":    resellerStatus,
			"expires_at":         expiresAt,
			"status":             domainStatus,
		})
		log.Printf("[Reseller] Updated existing domain record: %s", domainName)
		return
	}

	// Create new domain record
	domain := models.Domain{
		UserID:            userID,
		DomainName:        domainName,
		TLD:               tld,
		Status:            domainStatus,
		AutoRenew:         true,
		RegistrationDate:  now,
		ExpiresAt:         expiresAt,
		PrivacyProtection: true,
		Nameservers:       "ns1.domain.bd,ns2.domain.bd",
		ResellerDomainID:  resellerDomainID,
		ResellerStatus:    resellerStatus,
	}

	if err := rs.db.Create(&domain).Error; err != nil {
		log.Printf("[Reseller] ❌ Failed to save domain %s to DB: %v", domainName, err)
		return
	}

	log.Printf("[Reseller] 💾 Domain saved to DB: %s (ID: %d, ResellerID: %s)",
		domainName, domain.ID, resellerDomainID)
}
