package find_domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DomainService handles domain search business logic
type DomainService struct {
	client *http.Client
}

// NewDomainService creates a new DomainService
func NewDomainService() *DomainService {
	return &DomainService{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// ============================================================
// CONFIG
// ============================================================

const domainAPIBaseURL = "https://api.domainresellerapi.com/api/v1"

// IMPORTANT:
// Replace these with values from your ENV/config.
const resellerID = "YOUR_RESELLER_ID"
const apiKey = "YOUR_API_KEY"

// ============================================================
// SUPPORTED TLDs
// ============================================================

var supportedTLDs = []string{
	".com",
	".net",
	".info",
	".xyz",
	".dev",
}

// ============================================================
// DTO
// ============================================================

type DomainAPIRequest struct {
	DomainName string `json:"domainName"`
}

type BulkSearchRequest struct {
	DomainName string `json:"domainName"`
}

func (s *DomainService) Search(query string) (interface{}, error) {

	domain := cleanDomain(query)

	if domain == "" {
		return nil, errors.New("search query cannot be empty")
	}

	// If user doesn't provide TLD,
	// automatically use .com
	if !hasTLD(domain) {
		domain = domain + ".com"
	}

	payload := []DomainAPIRequest{
		{
			DomainName: domain,
		},
	}

	return s.callDomainAPI(
		"/domains/search",
		payload,
	)
}


func (s *DomainService) BulkSearch(query string) (interface{}, error) {

	domain := cleanDomain(query)

	if domain == "" {
		return nil, errors.New("domain name cannot be empty")
	}

	// ----------------------------------------------------------
	// User already provided TLD
	//
	// Example:
	// eftakhar.com
	// eftakhar.dev
	//
	// DON'T call external API
	// ----------------------------------------------------------

	if hasTLD(domain) {

		return map[string]interface{}{
			"domainName": domain,
			"searched":   false,
			"message":    "Top-level domain already provided",
			"results": []DomainAPIRequest{
				{
					DomainName: domain,
				},
			},
		}, nil
	}

	// ----------------------------------------------------------
	// Generate domains
	// ----------------------------------------------------------

	payload := make([]DomainAPIRequest, 0, len(supportedTLDs))

	for _, tld := range supportedTLDs {

		payload = append(
			payload,
			DomainAPIRequest{
				DomainName: domain + tld,
			},
		)
	}

	// ----------------------------------------------------------
	// Call external bulk API
	// ----------------------------------------------------------

	return s.callDomainAPI(
		"/domains/bulk-search",
		payload,
	)
}

// ============================================================
// EXTERNAL API REQUEST
// ============================================================

func (s *DomainService) callDomainAPI(
	endpoint string,
	payload interface{},
) (interface{}, error) {

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to marshal request: %w",
			err,
		)
	}

	// Full URL
	url := domainAPIBaseURL + endpoint

	// Create HTTP request
	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create request: %w",
			err,
		)
	}

	// ========================================================
	// HEADERS
	// ========================================================

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"Accept",
		"application/json",
	)

	req.Header.Set(
		"__reseller",
		resellerID,
	)

	req.Header.Set(
		"X-API-KEY",
		apiKey,
	)

	// ========================================================
	// SEND REQUEST
	// ========================================================

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to domain API: %w",
			err,
		)
	}

	defer resp.Body.Close()

	// ========================================================
	// READ RESPONSE
	// ========================================================

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read domain API response: %w",
			err,
		)
	}

	// ========================================================
	// CHECK STATUS
	// ========================================================

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {

		return nil, fmt.Errorf(
			"domain API returned status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	// ========================================================
	// PARSE JSON
	// ========================================================

	var result interface{}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf(
			"failed to parse domain API response: %w",
			err,
		)
	}

	return result, nil
}

// ============================================================
// HELPERS
// ============================================================
func cleanDomain(query string) string {

	q := strings.TrimSpace(
		strings.ToLower(query),
	)

	q = strings.TrimPrefix(q, "http://")
	q = strings.TrimPrefix(q, "https://")
	q = strings.TrimPrefix(q, "www.")

	q = strings.TrimSuffix(q, "/")

	return q
}
func hasTLD(domain string) bool {

	parts := strings.Split(domain, ".")

	if len(parts) < 2 {
		return false
	}

	// Invalid:
	// .com
	// eftakhar.
	if parts[0] == "" ||
		parts[len(parts)-1] == "" {
		return false
	}

	return true
}
