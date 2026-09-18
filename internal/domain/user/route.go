package user

import (
	"project-setup/internal/middleware"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// UserRoutes registers all user routes (protected, JWT required)
func UserRoutes(r *echo.Group, db *gorm.DB) {
	// Dependency Injection
	userRepository := NewUserRepository(db)
	userService := NewUserService(userRepository)
	userHandler := NewUserHandler(userService)

	// Apply JWT middleware to ALL user routes
	r.Use(middleware.JWTMiddleware)

	// Routes accessible by any authenticated user
	r.GET("/profile", userHandler.GetProfile)
	r.PUT("/profile", userHandler.UpdateProfile)

	// Routes accessible by ADMIN and SUPERADMIN only
	r.GET("/admin/stats", userHandler.GetAdminStats, middleware.RequireRole("ADMIN", "SUPERADMIN"))
	r.GET("", userHandler.GetUsers, middleware.RequireRole("ADMIN", "SUPERADMIN"))
	r.PUT("/:id", userHandler.AdminUpdateUser, middleware.RequireRole("ADMIN", "SUPERADMIN"))
	r.DELETE("/:id", userHandler.DeleteUser, middleware.RequireRole("ADMIN", "SUPERADMIN"))

	// Routes accessible by SUPERADMIN only
	r.PUT("/:id/role", userHandler.UpdateRole, middleware.RequireRole("SUPERADMIN"))
}
