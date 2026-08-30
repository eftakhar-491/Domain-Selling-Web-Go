package auth

import (
	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// AuthRoutes registers all authentication routes (public, no JWT required)
func AuthRoutes(r *echo.Group, db *gorm.DB, redisClient *redis.Client) {
	// Dependency Injection
	authRepository := NewAuthRepository(db)
	authService := NewAuthService(authRepository, redisClient)
	authHandler := NewAuthHandler(authService)

	// Public routes — no authentication required
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/forgot-password", authHandler.ForgotPassword)
	r.POST("/reset-password", authHandler.ResetPassword)
}
