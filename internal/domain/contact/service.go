package contact

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"project-setup/internal/models"
	"project-setup/internal/pkg/mail"
)

type Service interface {
	SendMessage(req *CreateContactRequest) (*models.ContactMessage, error)
	GetAllMessages(page, limit int) ([]models.ContactMessage, int64, error)
	MarkStatus(id uint, status string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) SendMessage(req *CreateContactRequest) (*models.ContactMessage, error) {
	name := strings.TrimSpace(req.Name)
	email := strings.TrimSpace(req.Email)
	subject := strings.TrimSpace(req.Subject)
	phone := strings.TrimSpace(req.Phone)
	messageText := strings.TrimSpace(req.Message)

	if name == "" {
		return nil, errors.New("name is required")
	}
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.New("valid email address is required")
	}
	if messageText == "" {
		return nil, errors.New("message content is required")
	}
	if subject == "" {
		subject = "General Inquiry from domain.bd"
	}

	msg := &models.ContactMessage{
		Name:    name,
		Email:   email,
		Phone:   phone,
		Subject: subject,
		Message: messageText,
		Status:  "UNREAD",
	}

	if err := s.repo.Create(msg); err != nil {
		return nil, fmt.Errorf("failed to save contact message: %w", err)
	}

	// Determine recipient support email
	supportEmail := "support@domain.bd"
	setting, err := s.repo.GetSystemSetting()
	if err == nil && setting != nil && strings.TrimSpace(setting.SupportEmail) != "" {
		supportEmail = strings.TrimSpace(setting.SupportEmail)
	}

	// Async email dispatch so user doesn't wait for SMTP latency
	go func(toAdmin string, userEmail string, userName string, userPhone string, userSub string, userMsg string) {
		// 1. Notify Admin Support
		adminBody := fmt.Sprintf(
			"You have received a new inquiry via domain.bd contact form:<br><br>"+
				"<strong>From:</strong> %s (%s)<br>"+
				"<strong>Phone:</strong> %s<br>"+
				"<strong>Subject:</strong> %s<br>"+
				"<strong>Message:</strong><br><blockquote style=\"background:#f5f5f7;padding:12px;border-left:4px solid #f5a623;\">%s</blockquote>",
			userName, userEmail, userPhone, userSub, userMsg,
		)
		if err := mail.SendNotificationEmail(toAdmin, fmt.Sprintf("📬 New Contact Inquiry: %s", userSub), adminBody); err != nil {
			log.Printf("⚠️ Failed to send contact email to admin: %v\n", err)
		}

		// 2. Send Acknowledgment Receipt to User
		userAckBody := fmt.Sprintf(
			"Hello %s,<br><br>"+
				"Thank you for contacting <strong>domain.bd</strong>. We have safely received your inquiry regarding <em>\"%s\"</em>.<br><br>"+
				"Our technical support engineers in Dhaka are reviewing your request and will respond within 15–60 minutes during support hours.<br><br>"+
				"Your submitted message:<br>"+
				"<blockquote style=\"background:#f5f5f7;padding:12px;border-left:4px solid #3b82f6;\">%s</blockquote><br>"+
				"Warm regards,<br>"+
				"<strong>The domain.bd Support Team</strong><br>"+
				"<small>Dhaka, Bangladesh</small>",
			userName, userSub, userMsg,
		)
		if err := mail.SendNotificationEmail(userEmail, "We received your message — domain.bd Support", userAckBody); err != nil {
			log.Printf("⚠️ Failed to send contact acknowledgment email to user: %v\n", err)
		}
	}(supportEmail, email, name, phone, subject, messageText)

	return msg, nil
}

func (s *service) GetAllMessages(page, limit int) ([]models.ContactMessage, int64, error) {
	return s.repo.FindAll(page, limit)
}

func (s *service) MarkStatus(id uint, status string) error {
	return s.repo.UpdateStatus(id, status)
}
