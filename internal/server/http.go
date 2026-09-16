package server

import (
	"net/http"
	"os"

	customValidator "project-setup/internal/validator"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// StartServer initializes and starts the Echo HTTP server
func StartServer(db *gorm.DB, redisClient *redis.Client) {
	e := echo.New()
	e.Validator = customValidator.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:5173",
	}
	if fe := os.Getenv("FRONTEND_URL"); fe != "" {
		allowedOrigins = append(allowedOrigins, fe)
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
	}))

	// Root Health Check Handlers (used by Render & uptime monitoring)
	e.GET("/", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "healthy",
			"message": "Domain Selling API is running successfully!",
		})
	})
	e.HEAD("/", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	// API Version 1 Group
	r := e.Group("/api/v1")
	Routes(r, db, redisClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	if err := e.Start(":" + port); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
