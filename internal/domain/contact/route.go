package contact

import (
	"project-setup/internal/middleware"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

func ContactRoutes(r *echo.Group, db *gorm.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	// Public contact submission endpoint (no auth needed)
	r.POST("", handler.SendMessage)

	// Admin inquiry management (requires JWT + role)
	r.GET("", handler.GetAll, middleware.JWTMiddleware, middleware.RequireRole("ADMIN", "SUPERADMIN"))
	r.PUT("/:id/status", handler.MarkStatus, middleware.JWTMiddleware, middleware.RequireRole("ADMIN", "SUPERADMIN"))
}
