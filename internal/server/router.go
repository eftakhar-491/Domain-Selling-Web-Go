package server

import (
	"net/http"

	"project-setup/internal/domain/auth"
	"project-setup/internal/domain/find_domain"
	"project-setup/internal/domain/user"

	"github.com/labstack/echo/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Routes registers all API route groups
func Routes(r *echo.Group, db *gorm.DB, redisClient *redis.Client) {
	// Health check
	r.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "API is running successfully!",
			"status":  "healthy",
		})
	})

	// Domain route groups
	auth.AuthRoutes(r.Group("/auth"), db, redisClient)
	user.UserRoutes(r.Group("/user"), db)
	find_domain.DomainRoutes(r.Group("/find-domain"), db)
}
