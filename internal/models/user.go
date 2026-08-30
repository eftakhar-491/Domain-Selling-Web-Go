package models

import "gorm.io/gorm"

// Role type for user roles
type Role string

const (
	RoleUser       Role = "USER"
	RoleAdmin      Role = "ADMIN"
	RoleSuperAdmin Role = "SUPERADMIN"
)

// User represents the user entity in the database
type User struct {
	gorm.Model
	Name        string  `json:"name" gorm:"type:varchar(255);not null"`
	Email       string  `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Password    string  `json:"-" gorm:"type:varchar(255);not null"`
	PhoneNumber *string `json:"phone_number,omitempty" gorm:"type:varchar(20)"`
	Role        Role    `json:"role" gorm:"type:varchar(20);default:'USER';not null"`
	IsActive    bool    `json:"is_active" gorm:"default:true"`
}
