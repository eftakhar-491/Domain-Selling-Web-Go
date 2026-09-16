package main

import (
	"fmt"
	"log"
	"os"

	"project-setup/internal/config"
	"project-setup/internal/models"
	"project-setup/internal/pkg/utils"
	"project-setup/internal/server"
)

func main() {
	// Initialize Database Connection (active)
	fmt.Println("Database Connecting...")
	db := config.ConnectDB()

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.User{},
		&models.Discount{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.Domain{},
		&models.DNSRecord{},
	); err != nil {
		log.Printf("⚠️ Auto-migrate warning: %v\n", err)
	}
	fmt.Println("Database Connected & Migrated")

	// Initialize Redis Connection
	fmt.Println("Redis Connecting...")
	redisClient := config.ConnectRedis()

	// Seed default SuperAdmin account
	seedSuperAdmin()

	// Seed starter discount coupons
	seedDefaultCoupons()

	// Start HTTP server
	fmt.Println("Server Starting...")
	server.StartServer(db, redisClient)
}

// seedSuperAdmin creates a default SUPERADMIN user if one doesn't already exist
func seedSuperAdmin() {
	db := config.DB

	email := os.Getenv("SUPERADMIN_EMAIL")
	password := os.Getenv("SUPERADMIN_PASSWORD")
	name := os.Getenv("SUPERADMIN_NAME")

	if email == "" || password == "" {
		log.Println("⚠️ SUPERADMIN_EMAIL or SUPERADMIN_PASSWORD not set. Skipping SuperAdmin seed.")
		return
	}

	if name == "" {
		name = "Super Admin"
	}

	// Check if SUPERADMIN already exists
	var existingUser models.User
	result := db.Where("email = ?", email).First(&existingUser)
	if result.Error == nil {
		log.Println("ℹ️ SuperAdmin account already exists:", email)
		return
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Fatal("Failed to hash SuperAdmin password:", err)
	}

	// Create SUPERADMIN user
	superAdmin := models.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Role:     models.RoleSuperAdmin,
		IsActive: true,
	}

	if err := db.Create(&superAdmin).Error; err != nil {
		log.Fatal("Failed to seed SuperAdmin:", err)
	}

	log.Println("🎉 SuperAdmin account created successfully:", email)
}

// seedDefaultCoupons inserts starter coupons if they do not exist
func seedDefaultCoupons() {
	db := config.DB
	if db == nil {
		return
	}

	coupons := []struct {
		code        string
		name        string
		description string
		typ         models.DiscountType
		value       float64
		minSpend    float64
	}{
		{
			code:        "WELCOME10",
			name:        "Welcome 10% Off",
			description: "Enjoy 10% discount on your domain order",
			typ:         models.DiscountTypePercentage,
			value:       10.00,
			minSpend:    0.00,
		},
		{
			code:        "SAVE5",
			name:        "$5 Off Domain Order",
			description: "$5 fixed discount on domain orders of $10 or more",
			typ:         models.DiscountTypeFixed,
			value:       5.00,
			minSpend:    10.00,
		},
	}

	for _, c := range coupons {
		var existing models.Discount
		codeStr := c.code
		if err := db.Where("code = ?", codeStr).First(&existing).Error; err != nil {
			discount := models.Discount{
				Code:        &codeStr,
				Name:        c.name,
				Description: c.description,
				Type:        c.typ,
				Scope:       models.DiscountScopeCoupon,
				Value:       c.value,
				MinSpend:    c.minSpend,
				UsageLimit:  5000,
				IsActive:    true,
			}
			if err := db.Create(&discount).Error; err == nil {
				log.Println("🏷️ Seeded default coupon code:", c.code)
			}
		}
	}
}
