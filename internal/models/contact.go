package models

import (
	"gorm.io/gorm"
)

// ContactMessage stores inquiries submitted via the public contact form
type ContactMessage struct {
	gorm.Model
	Name    string `gorm:"type:varchar(255);not null" json:"name"`
	Email   string `gorm:"type:varchar(255);not null" json:"email"`
	Phone   string `gorm:"type:varchar(50)" json:"phone"`
	Subject string `gorm:"type:varchar(255);default:'General Inquiry'" json:"subject"`
	Message string `gorm:"type:text;not null" json:"message"`
	Status  string `gorm:"type:varchar(50);default:'UNREAD'" json:"status"` // UNREAD, READ, REPLIED, ARCHIVED
}
