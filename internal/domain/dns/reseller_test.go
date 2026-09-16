package dns

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestUpdateNameServer_Success(t *testing.T) {
	// Set dummy env variables
	os.Setenv("DNA_RESELLER_ID", "test-reseller-id")
	os.Setenv("DNA_API_KEY", "test-api-key")

	var capturedReq struct {
		DomainName  string   `json:"domainName"`
		NameServers []string `json:"nameServers"`
	}
	var capturedHeaders http.Header
	var capturedMethod string
	var capturedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		capturedHeaders = r.Header.Clone()

		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedReq)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"message":"Nameservers updated successfully"}`))
	}))
	defer server.Close()

	os.Setenv("DNA_BASE_URL", server.URL)

	client := NewDNAResellerClient()
	resp, err := client.UpdateNameServer("example.com", []string{"ns1.example.com", "ns2.example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}

	if capturedMethod != http.MethodPut {
		t.Errorf("expected PUT method, got %s", capturedMethod)
	}

	if capturedPath != "/domains/dns/name-server" {
		t.Errorf("expected path /domains/dns/name-server, got %s", capturedPath)
	}

	// Verify headers
	if capturedHeaders.Get("DNA_RESELLER_ID") != "test-reseller-id" {
		t.Errorf("expected DNA_RESELLER_ID header, got %s", capturedHeaders.Get("DNA_RESELLER_ID"))
	}
	if capturedHeaders.Get("DNA_API_KEY") != "test-api-key" {
		t.Errorf("expected DNA_API_KEY header, got %s", capturedHeaders.Get("DNA_API_KEY"))
	}
	if capturedHeaders.Get("__reseller") != "test-reseller-id" {
		t.Errorf("expected __reseller header, got %s", capturedHeaders.Get("__reseller"))
	}
	if capturedHeaders.Get("X-API-KEY") != "test-api-key" {
		t.Errorf("expected X-API-KEY header, got %s", capturedHeaders.Get("X-API-KEY"))
	}

	// Verify body
	if capturedReq.DomainName != "example.com" {
		t.Errorf("expected domainName example.com, got %s", capturedReq.DomainName)
	}
	if len(capturedReq.NameServers) != 2 || capturedReq.NameServers[0] != "ns1.example.com" {
		t.Errorf("unexpected nameServers payload: %v", capturedReq.NameServers)
	}
}
