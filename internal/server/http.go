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
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
	}))

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
