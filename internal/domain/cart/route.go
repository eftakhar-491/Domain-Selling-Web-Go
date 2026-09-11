package cart

import (
	"project-setup/internal/domain/discount"
	"project-setup/internal/middleware"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

func CartRoutes(r *echo.Group, db *gorm.DB) {
	discountRepo := discount.NewDiscountRepository(db)
	discountService := discount.NewDiscountService(discountRepo)

	repo := NewCartRepository(db)
	service := NewCartService(repo, discountService)
	handler := NewCartHandler(service)

	// All cart routes require authentication
	r.Use(middleware.JWTMiddleware)

	r.GET("", handler.GetCart)
	r.POST("/items", handler.AddToCart)
	r.PUT("/items/:id", handler.UpdateCartItem)
	r.DELETE("/items/:id", handler.RemoveCartItem)
	r.DELETE("", handler.ClearCart)
	r.POST("/coupon", handler.ApplyCoupon)
	r.DELETE("/coupon", handler.RemoveCoupon)
}
