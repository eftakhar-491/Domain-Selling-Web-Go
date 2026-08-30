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
	// Initialize Database Connection
	fmt.Println("Database Connecting...")
	db := config.ConnectDB()

	// Auto-migrate user models
	// if err := db.AutoMigrate(
	// 	&models.User{},
	// ); err != nil {
	// 	log.Printf("⚠️ Auto-migrate warning: %v\n", err)
	// }
	fmt.Println("Database Connected & Migrated")

	// Initialize Redis Connection
	fmt.Println("Redis Connecting...")
	redisClient := config.ConnectRedis()

	// Seed default SuperAdmin account
	seedSuperAdmin()

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
