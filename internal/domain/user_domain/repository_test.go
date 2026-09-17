package user_domain

import (
	"testing"

	"project-setup/internal/config"

	"github.com/joho/godotenv"
)

func TestGetDomainsByUserID(t *testing.T) {
	_ = godotenv.Load("../../../.env")
	db := config.ConnectDB()
	repo := NewDomainRepository(db)

	domains18, err := repo.GetDomainsByUserID(18, "")
	if err != nil {
		t.Fatalf("Failed to get domains for user 18: %v", err)
	}

	t.Logf("User 18 domains count: %d", len(domains18))
	for _, d := range domains18 {
		t.Logf("  User 18 domain: %s (Status: %s)", d.DomainName, d.Status)
		if d.DomainName == "orbitstudio.com" || d.DomainName == "nexora.io" {
			t.Errorf("Forbidden fake domain found: %s", d.DomainName)
		}
	}

	domains20, err := repo.GetDomainsByUserID(20, "")
	if err != nil {
		t.Fatalf("Failed to get domains for user 20: %v", err)
	}
	t.Logf("User 20 domains count: %d", len(domains20))
	for _, d := range domains20 {
		t.Logf("  User 20 domain: %s (Status: %s)", d.DomainName, d.Status)
		if d.DomainName == "orbitstudio.com" || d.DomainName == "nexora.io" {
			t.Errorf("Forbidden fake domain found: %s", d.DomainName)
		}
	}
}
