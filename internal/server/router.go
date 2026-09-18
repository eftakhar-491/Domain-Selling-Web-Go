package server

import (
	"net/http"

	"project-setup/internal/domain/auth"
	"project-setup/internal/domain/cart"
	"project-setup/internal/domain/contact"
	"project-setup/internal/domain/discount"
	"project-setup/internal/domain/dns"
	"project-setup/internal/domain/find_domain"
	"project-setup/internal/domain/order"
	"project-setup/internal/domain/user"
	"project-setup/internal/domain/user_domain"

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
	user_domain.DomainRoutes(r.Group("/domains"), db)
	find_domain.DomainRoutes(r.Group("/find-domain"))
	discount.DiscountRoutes(r.Group("/discounts"), db)
	cart.CartRoutes(r.Group("/cart"), db)
	order.OrderRoutes(r.Group("/orders"), db)
	dns.DNSRoutes(r.Group("/dns"), db)
	contact.ContactRoutes(r.Group("/contact"), db)
}
