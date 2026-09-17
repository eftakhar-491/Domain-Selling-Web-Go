package user_domain

import (
	"testing"
	"time"

	"project-setup/internal/models"
)

func TestToDomainResponse(t *testing.T) {
	service := &DomainService{}
	now := time.Now()

	tests := []struct {
		name         string
		domain       models.Domain
		expectedTone string
		expectedLbl  string
	}{
		{
			name: "Active domain",
			domain: models.Domain{
				DomainName: "test-active.com",
				TLD:        "com",
				Status:     models.DomainStatusActive,
				ExpiresAt:  now.AddDate(1, 0, 0),
			},
			expectedTone: "green",
			expectedLbl:  "Active",
		},
		{
			name: "Pending domain from recent order",
			domain: models.Domain{
				DomainName: "test-pending.com",
				TLD:        "com",
				Status:     models.DomainStatusPending,
				ExpiresAt:  now.AddDate(1, 0, 0),
			},
			expectedTone: "amber",
			expectedLbl:  "Pending",
		},
		{
			name: "Expiring soon domain",
			domain: models.Domain{
				DomainName: "test-expiring.com",
				TLD:        "com",
				Status:     models.DomainStatusExpiringSoon,
				ExpiresAt:  now.AddDate(0, 0, 10),
			},
			expectedTone: "amber",
			expectedLbl:  "Renew soon",
		},
		{
			name: "Expired domain",
			domain: models.Domain{
				DomainName: "test-expired.com",
				TLD:        "com",
				Status:     models.DomainStatusActive,
				ExpiresAt:  now.AddDate(0, 0, -5),
			},
			expectedTone: "red",
			expectedLbl:  "Expired",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := service.toDomainResponse(&tc.domain)
			if resp.Tone != tc.expectedTone {
				t.Errorf("expected tone %q, got %q", tc.expectedTone, resp.Tone)
			}
			if resp.StatusLabel != tc.expectedLbl {
				t.Errorf("expected label %q, got %q", tc.expectedLbl, resp.StatusLabel)
			}
		})
	}
}
