package find_domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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
const resellerID = "aa8ba854-e7c9-ae05-3ac3-3a2283c93d2e"
const apiKey = "rz7UYB8Uo02uloexdaRY9kligZwMeRtJyqWI3LXM6E"

// IMPORTANT:
// Replace these with values from your ENV/config.

// ============================================================
// SUPPORTED TLDs
// ============================================================

var supportedTLDs = []string{
	".com",
	".net",
	".info",
	".xyz",
	".dev",
	".ai",
	".io",
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

func hasTLD(domain string) bool {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}
	if parts[0] == "" ||
		parts[len(parts)-1] == "" {
		return false
	}
	return true
}

func (s *DomainService) Search(query string) (interface{}, error) {

	domain := query

	if domain == "" {
		return nil, errors.New("search query or domain cannot be empty")
	}

	// If user doesn't provide TLD,
	// automatically use .com
	if !hasTLD(domain) {
		domain = domain + ".com"
	}

	payload := DomainAPIRequest{
		DomainName: domain,
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to marshal request: %w",
			err,
		)
	}

	baseURL := os.Getenv("DNA_BASE_URL")
	if baseURL == "" {
		baseURL = domainAPIBaseURL
	}
	rID := os.Getenv("DNA_RESELLER_ID")
	if rID == "" {
		rID = resellerID
	}
	key := os.Getenv("DNA_API_KEY")
	if key == "" {
		key = apiKey
	}
	// Full URL
	url := baseURL + "/domains/search"
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
	// HEADERS
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
		rID,
	)

	req.Header.Set(
		"X-API-KEY",
		key,
	)

	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
	)
	// SEND REQUEST
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to domain API: %w",
			err,
		)
	}
	defer resp.Body.Close()
	// READ RESPONSE
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read domain API response: %w",
			err,
		)
	}
	// CHECK STATUS
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"domain API returned status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}
	// PARSE JSON
	var result interface{}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf(
			"failed to parse domain API response: %w",
			err,
		)
	}

	return result, nil
}

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

func (s *DomainService) BulkSearch(query string) (interface{}, error) {
	domain := query
	if domain == "" {
		return nil, errors.New("domain name cannot be empty")
	}
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
	// Generate domains
	payload := make([]DomainAPIRequest, 0, len(supportedTLDs))

	for _, tld := range supportedTLDs {
		payload = append(
			payload,
			DomainAPIRequest{
				DomainName: domain + tld,
			},
		)
	}
	// Call external bulk API

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to marshal request: %w",
			err,
		)
	}

	baseURL := os.Getenv("DNA_BASE_URL")
	if baseURL == "" {
		baseURL = domainAPIBaseURL
	}
	rID := os.Getenv("DNA_RESELLER_ID")
	if rID == "" {
		rID = resellerID
	}
	key := os.Getenv("DNA_API_KEY")
	if key == "" {
		key = apiKey
	}

	// Full URL
	url := baseURL + "/domains/bulk-search"

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
		rID,
	)

	req.Header.Set(
		"X-API-KEY",
		key,
	)

	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
	)
	// SEND REQUEST
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to domain API: %w",
			err,
		)
	}

	defer resp.Body.Close()

	// READ RESPONSE

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read domain API response: %w",
			err,
		)
	}

	// CHECK STATUS

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {

		return nil, fmt.Errorf(
			"domain API returned status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	// PARSE JSON

	var result interface{}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf(
			"failed to parse domain API response: %w",
			err,
		)
	}

	return result, nil
}
