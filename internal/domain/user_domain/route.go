package user_domain

import (
	"project-setup/internal/middleware"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// DomainRoutes registers all user domain management routes
func DomainRoutes(r *echo.Group, db *gorm.DB) {
	repo := NewDomainRepository(db)
	service := NewDomainService(repo)
	handler := NewDomainHandler(service)

	// Require JWT authentication for all user domain operations
	r.Use(middleware.JWTMiddleware)

	r.GET("", handler.GetUserDomains)
	r.GET("/admin/all", handler.AdminGetAllDomains, middleware.RequireRole("ADMIN", "SUPERADMIN"))
	r.PUT("/admin/:id/status", handler.AdminUpdateDomain, middleware.RequireRole("ADMIN", "SUPERADMIN"))
	r.GET("/:id", handler.GetDomainByID)
}
