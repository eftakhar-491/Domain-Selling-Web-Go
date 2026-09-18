package user_domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"project-setup/internal/models"
)

type DomainService struct {
	repo *DomainRepository
}

func NewDomainService(repo *DomainRepository) *DomainService {
	return &DomainService{repo: repo}
}

// GetUserDomains retrieves and calculates stats for user's domains
func (s *DomainService) GetUserDomains(userID uint, search string) (*DomainsListResponse, error) {
	domains, err := s.repo.GetDomainsByUserID(userID, search)
	if err != nil {
		return nil, err
	}

	var items []DomainResponse
	activeCount := 0
	renewSoonCount := 0
	autoRenewCount := 0
	now := time.Now()

	for _, d := range domains {
		daysUntil := int(d.ExpiresAt.Sub(now).Hours() / 24)

		status := string(d.Status)
		tone := "green"
		statusLabel := "Active"

		if d.ExpiresAt.Before(now) {
			status = string(models.DomainStatusExpired)
			statusLabel = "Expired"
			tone = "red"
		} else if d.Status == models.DomainStatusPending {
			status = string(models.DomainStatusPending)
			statusLabel = "Pending"
			tone = "amber"
		} else if daysUntil <= 30 || d.Status == models.DomainStatusExpiringSoon {
			status = string(models.DomainStatusExpiringSoon)
			statusLabel = "Renew soon"
			tone = "amber"
			renewSoonCount++
		} else {
			activeCount++
		}

		if d.AutoRenew {
			autoRenewCount++
		}

		autoRenewLabel := "Off"
		if d.AutoRenew {
			autoRenewLabel = "On"
		}

		nameserverList := []string{}
		if d.Nameservers != "" {
			for _, ns := range strings.Split(d.Nameservers, ",") {
				if trimmed := strings.TrimSpace(ns); trimmed != "" {
					nameserverList = append(nameserverList, trimmed)
				}
			}
		}

		items = append(items, DomainResponse{
			ID:                d.ID,
			DomainName:        d.DomainName,
			TLD:               d.TLD,
			Status:            status,
			StatusLabel:       statusLabel,
			Tone:              tone,
			AutoRenew:         d.AutoRenew,
			AutoRenewLabel:    autoRenewLabel,
			RegistrationDate:  d.RegistrationDate.Format("Jan 02, 2006"),
			ExpiresAt:         d.ExpiresAt.Format(time.RFC3339),
			FormattedExpiry:   d.ExpiresAt.Format("Jan 02, 2006"),
			PrivacyProtection: d.PrivacyProtection,
			Nameservers:       nameserverList,
			DaysUntilExpiry:   daysUntil,
		})
	}

	totalDomains := len(items)
	autoRenewRatio := fmt.Sprintf("%02d / %02d", autoRenewCount, totalDomains)
	autoRenewNote := "All protected"
	if totalDomains == 0 {
		autoRenewNote = "No domains"
	} else if autoRenewCount < totalDomains {
		autoRenewNote = fmt.Sprintf("%d needs attention", totalDomains-autoRenewCount)
	}

	return &DomainsListResponse{
		Domains: items,
		Metrics: DomainMetrics{
			TotalActive:   activeCount,
			RenewingSoon:  renewSoonCount,
			AutoRenew:     autoRenewRatio,
			TotalDomains:  totalDomains,
			AutoRenewNote: autoRenewNote,
		},
	}, nil
}

// GetDomainByID gets single domain for user
func (s *DomainService) GetDomainByID(userID uint, domainID uint) (*DomainResponse, error) {
	domain, err := s.repo.GetDomainByIDAndUser(domainID, userID)
	if err != nil {
		return nil, errors.New("domain not found")
	}

	return s.toDomainResponse(domain), nil
}

func (s *DomainService) toDomainResponse(d *models.Domain) *DomainResponse {
	now := time.Now()
	daysUntil := int(d.ExpiresAt.Sub(now).Hours() / 24)

	status := string(d.Status)
	tone := "green"
	statusLabel := "Active"

	if d.ExpiresAt.Before(now) {
		status = string(models.DomainStatusExpired)
		statusLabel = "Expired"
		tone = "red"
	} else if d.Status == models.DomainStatusPending {
		status = string(models.DomainStatusPending)
		statusLabel = "Pending"
		tone = "amber"
	} else if daysUntil <= 30 || d.Status == models.DomainStatusExpiringSoon {
		status = string(models.DomainStatusExpiringSoon)
		statusLabel = "Renew soon"
		tone = "amber"
	}

	autoRenewLabel := "Off"
	if d.AutoRenew {
		autoRenewLabel = "On"
	}

	nameserverList := []string{}
	if d.Nameservers != "" {
		for _, ns := range strings.Split(d.Nameservers, ",") {
			if trimmed := strings.TrimSpace(ns); trimmed != "" {
				nameserverList = append(nameserverList, trimmed)
			}
		}
	}

	return &DomainResponse{
		ID:                d.ID,
		DomainName:        d.DomainName,
		TLD:               d.TLD,
		Status:            status,
		StatusLabel:       statusLabel,
		Tone:              tone,
		AutoRenew:         d.AutoRenew,
		AutoRenewLabel:    autoRenewLabel,
		RegistrationDate:  d.RegistrationDate.Format("Jan 02, 2006"),
		ExpiresAt:         d.ExpiresAt.Format(time.RFC3339),
		FormattedExpiry:   d.ExpiresAt.Format("Jan 02, 2006"),
		PrivacyProtection: d.PrivacyProtection,
		Nameservers:       nameserverList,
		DaysUntilExpiry:   daysUntil,
	}
}

// AdminGetAllDomains returns a paginated list of all domains across the platform for admins
func (s *DomainService) AdminGetAllDomains(page, limit int, search string, status string) (*AdminDomainListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	domains, total, err := s.repo.GetAllDomainsAdmin(page, limit, search, status)
	if err != nil {
		return nil, errors.New("failed to fetch domains: " + err.Error())
	}

	var items []AdminDomainResponse
	activeCount := 0
	renewSoonCount := 0
	autoRenewCount := 0

	for _, d := range domains {
		resp := s.toAdminDomainResponse(&d)
		if resp.Status == string(models.DomainStatusActive) {
			activeCount++
		} else if resp.Status == string(models.DomainStatusExpiringSoon) {
			renewSoonCount++
		}
		if resp.AutoRenew {
			autoRenewCount++
		}
		items = append(items, resp)
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &AdminDomainListResponse{
		Domains:    items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		Metrics: DomainMetrics{
			TotalActive:   activeCount,
			RenewingSoon:  renewSoonCount,
			AutoRenew:     fmt.Sprintf("%02d / %02d", autoRenewCount, len(items)),
			TotalDomains:  int(total),
			AutoRenewNote: fmt.Sprintf("%d total registered", total),
		},
	}, nil
}

// AdminUpdateDomain updates status and settings for a domain (admin only)
func (s *DomainService) AdminUpdateDomain(domainID uint, req AdminUpdateDomainStatusRequest) (*AdminDomainResponse, error) {
	domain, err := s.repo.UpdateDomainStatus(domainID, req.Status, req.AutoRenew)
	if err != nil {
		return nil, errors.New("failed to update domain: " + err.Error())
	}

	resp := s.toAdminDomainResponse(domain)
	return &resp, nil
}

func (s *DomainService) toAdminDomainResponse(d *models.Domain) AdminDomainResponse {
	now := time.Now()
	daysUntil := int(d.ExpiresAt.Sub(now).Hours() / 24)

	status := string(d.Status)
	tone := "green"
	statusLabel := "Active"

	if d.ExpiresAt.Before(now) {
		status = string(models.DomainStatusExpired)
		statusLabel = "Expired"
		tone = "red"
	} else if d.Status == models.DomainStatusPending {
		status = string(models.DomainStatusPending)
		statusLabel = "Pending"
		tone = "amber"
	} else if d.Status == models.DomainStatusSuspended {
		status = string(models.DomainStatusSuspended)
		statusLabel = "Suspended"
		tone = "red"
	} else if daysUntil <= 30 || d.Status == models.DomainStatusExpiringSoon {
		status = string(models.DomainStatusExpiringSoon)
		statusLabel = "Renew soon"
		tone = "amber"
	}

	autoRenewLabel := "Off"
	if d.AutoRenew {
		autoRenewLabel = "On"
	}

	nameserverList := []string{}
	if d.Nameservers != "" {
		for _, ns := range strings.Split(d.Nameservers, ",") {
			if trimmed := strings.TrimSpace(ns); trimmed != "" {
				nameserverList = append(nameserverList, trimmed)
			}
		}
	}

	return AdminDomainResponse{
		ID:                d.ID,
		DomainName:        d.DomainName,
		TLD:               d.TLD,
		Status:            status,
		StatusLabel:       statusLabel,
		Tone:              tone,
		AutoRenew:         d.AutoRenew,
		AutoRenewLabel:    autoRenewLabel,
		RegistrationDate:  d.RegistrationDate.Format("Jan 02, 2006"),
		ExpiresAt:         d.ExpiresAt.Format(time.RFC3339),
		FormattedExpiry:   d.ExpiresAt.Format("Jan 02, 2006"),
		PrivacyProtection: d.PrivacyProtection,
		Nameservers:       nameserverList,
		DaysUntilExpiry:   daysUntil,
		User: DomainOwnerInfo{
			ID:          d.User.ID,
			Name:        d.User.Name,
			Email:       d.User.Email,
			PhoneNumber: d.User.PhoneNumber,
		},
		CreatedAt: d.CreatedAt.Format(time.RFC3339),
		UpdatedAt: d.UpdatedAt.Format(time.RFC3339),
	}
}

