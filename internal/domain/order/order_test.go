package order

import (
	"testing"
)

func TestExtractTLD(t *testing.T) {
	tests := []struct {
		domain   string
		expected string
	}{
		{"example.com", "com"},
		{"mycompany.org", "org"},
		{"techblog.co.uk", "co.uk"},
		{"sub.domain.xyz", "domain.xyz"},
		{"singleword", ""},
	}

	for _, tt := range tests {
		result := extractTLD(tt.domain)
		if result != tt.expected {
			t.Errorf("extractTLD(%q) = %q; expected %q", tt.domain, result, tt.expected)
		}
	}
}

func TestGenerateOrderNumber(t *testing.T) {
	num1 := generateOrderNumber()
	num2 := generateOrderNumber()

	if len(num1) == 0 {
		t.Error("expected non-empty order number")
	}

	if num1[:4] != "ORD-" {
		t.Errorf("expected order number to start with 'ORD-', got %q", num1)
	}

	if num1 == num2 {
		t.Errorf("expected generated order numbers to be unique, got both %q", num1)
	}
}
