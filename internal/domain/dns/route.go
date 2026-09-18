package dns

import (
	"project-setup/internal/middleware"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// DNSRoutes registers all DNS management routes
func DNSRoutes(r *echo.Group, db *gorm.DB) {
	repo := NewDNSRepository(db)
	service := NewDNSService(repo)
	handler := NewDNSHandler(service)

	// All DNS routes require JWT authentication
	r.Use(middleware.JWTMiddleware)

	// Nameserver management (registered before /:id)
	r.PUT("/name-server", handler.UpdateNameServers) // Update nameservers + sync via PUT to DNA API
	r.GET("/name-server", handler.GetNameServers)     // Get nameservers for domain (?domain=...)

	// CRUD operations
	r.POST("", handler.CreateRecord)                // Create DNS record + sync to DNA API (POST)
	r.GET("", handler.GetRecordsByDomain)            // List DNS records (optionally filter by ?domain=)
	r.GET("/:id", handler.GetRecordByID)             // Get single DNS record
	r.PUT("/:id", handler.UpdateRecord)              // Update DNS record + sync via PUT to DNA API
	r.PATCH("/:id", handler.UpdateRecord)            // Update DNS record + sync via PUT (partial)
	r.DELETE("/host", handler.DeleteByHost)           // Delete by domain+host → syncs DELETE to DNA API
	r.DELETE("/:id", handler.DeleteRecord)           // Delete single record by ID → syncs DELETE to DNA API
	r.POST("/:id/resync", handler.ResyncRecord)      // Re-sync failed record to DNA API

	// Admin routes
	r.GET("/admin/zones", handler.AdminGetAllZones, middleware.RequireRole("ADMIN", "SUPERADMIN"))
}
