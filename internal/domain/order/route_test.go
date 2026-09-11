package order

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"project-setup/internal/middleware"
	stripepkg "project-setup/internal/pkg/stripe"
	"project-setup/internal/pkg/token"
	customValidator "project-setup/internal/validator"

	"github.com/labstack/echo/v5"
)

func init() {
	_ = os.Setenv("JWT_SECRET", "test_jwt_secret_for_unit_testing_12345")
}

func setupTestRouter() (*echo.Echo, *OrderHandler) {
	e := echo.New()
	e.Validator = customValidator.New()

	stripeService := stripepkg.NewStripeService()
	service := &OrderService{
		stripeService: stripeService,
	}
	handler := NewOrderHandler(service)

	r := e.Group("/api/v1/orders")

	// Webhook
	r.POST("/webhook", handler.HandleWebhook)

	// Protected routes
	protected := r.Group("")
	protected.Use(middleware.JWTMiddleware)

	protected.POST("/checkout", handler.Checkout)
	protected.POST("/direct", handler.DirectOrder)
	protected.POST("/:id/confirm-payment", handler.ConfirmPayment)
	protected.GET("/:id", handler.GetOrder)
	protected.GET("", handler.GetMyOrders)
	protected.GET("/admin/all", handler.AdminGetAllOrders, middleware.RequireRole("ADMIN", "SUPERADMIN"))

	return e, handler
}

func TestRoutes_AuthenticationProtection(t *testing.T) {
	e, _ := setupTestRouter()

	endpoints := []struct {
		name   string
		method string
		path   string
	}{
		{"Checkout without token", http.MethodPost, "/api/v1/orders/checkout"},
		{"Direct order without token", http.MethodPost, "/api/v1/orders/direct"},
		{"Get orders without token", http.MethodGet, "/api/v1/orders"},
		{"Get order by id without token", http.MethodGet, "/api/v1/orders/1"},
		{"Confirm payment without token", http.MethodPost, "/api/v1/orders/1/confirm-payment"},
		{"Admin all without token", http.MethodGet, "/api/v1/orders/admin/all"},
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			req := httptest.NewRequest(ep.method, ep.path, bytes.NewBufferString("{}"))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("%s: expected 401 Unauthorized, got %d", ep.name, rec.Code)
			}
		})
	}
}

func TestRoutes_RoleBasedAccessControl(t *testing.T) {
	e, _ := setupTestRouter()

	// Generate real token using the project's token generator
	userToken, err := token.GenerateToken(10, "user@example.com", "USER")
	if err != nil {
		t.Fatalf("failed to generate user token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/admin/all", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for USER role accessing /admin/all, got %d. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestRoutes_WebhookRequiresSignature(t *testing.T) {
	e, _ := setupTestRouter()

	// Webhook request without Stripe-Signature header should return 400 Bad Request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/webhook", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for webhook without signature, got %d", rec.Code)
	}
}
