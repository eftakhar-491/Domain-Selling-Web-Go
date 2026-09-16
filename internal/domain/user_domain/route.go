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
	r.GET("/:id", handler.GetDomainByID)
	r.PUT("/:id", handler.UpdateDomain)
	r.PATCH("/:id", handler.UpdateDomain)
	r.POST("", handler.CreateDomain)
	r.DELETE("/:id", handler.DeleteDomain)
}
