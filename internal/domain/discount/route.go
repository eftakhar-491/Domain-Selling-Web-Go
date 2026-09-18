package discount

import (
	"project-setup/internal/middleware"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

func DiscountRoutes(r *echo.Group, db *gorm.DB) {
	repo := NewDiscountRepository(db)
	service := NewDiscountService(repo)
	handler := NewDiscountHandler(service)

	// Public
	r.GET("/tlds", handler.GetActiveTLDs)
	r.GET("/active", handler.GetActive)

	// Admin only
	r.Use(middleware.JWTMiddleware)
	r.POST("", handler.Create, middleware.RequireRole("ADMIN", "SUPERADMIN"))
	r.GET("", handler.GetAll, middleware.RequireRole("ADMIN", "SUPERADMIN"))
	r.GET("/:id", handler.GetByID, middleware.RequireRole("ADMIN", "SUPERADMIN"))
	r.PUT("/:id", handler.Update, middleware.RequireRole("ADMIN", "SUPERADMIN"))
	r.DELETE("/:id", handler.Delete, middleware.RequireRole("ADMIN", "SUPERADMIN"))
}
