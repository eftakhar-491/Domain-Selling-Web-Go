package order

import (
	"project-setup/internal/domain/discount"
	"project-setup/internal/middleware"
	stripepkg "project-setup/internal/pkg/stripe"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// OrderRoutes registers all order-related API routes
func OrderRoutes(r *echo.Group, db *gorm.DB) {
	// Dependencies
	discountRepo := discount.NewDiscountRepository(db)
	discountService := discount.NewDiscountService(discountRepo)
	stripeService := stripepkg.NewStripeService()

	repo := NewOrderRepository(db)
	service := NewOrderService(repo, db, discountService, stripeService)
	handler := NewOrderHandler(service)

	// Public — Stripe webhook (no JWT, Stripe validates via signature)
	r.POST("/webhook", handler.HandleWebhook)

	// Protected routes — require authentication
	protected := r.Group("")
	protected.Use(middleware.JWTMiddleware)

	protected.POST("/checkout", handler.Checkout)
	protected.POST("/direct", handler.DirectOrder)
	protected.POST("/:id/confirm-payment", handler.ConfirmPayment)
	protected.GET("/:id", handler.GetOrder)
	protected.GET("", handler.GetMyOrders)

	// Admin-only routes
	protected.GET("/admin/all", handler.AdminGetAllOrders, middleware.RequireRole("ADMIN", "SUPERADMIN"))
}
