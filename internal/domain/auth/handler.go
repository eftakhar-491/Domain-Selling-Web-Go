package auth

import (
	"net/http"

	"project-setup/internal/pkg/utils"

	"github.com/labstack/echo/v5"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	service *AuthService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	resp, err := h.service.Register(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusConflict, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", resp)
}

// Login handles user login
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	resp, err := h.service.Login(req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusUnauthorized, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Login successful", resp)
}

// ForgotPassword requests a 6-digit OTP code to be sent via email
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *echo.Context) error {
	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	if err := h.service.ForgotPassword(req); err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "If this email is registered, a password reset code has been sent", nil)
}

// ResetPassword verifies the OTP code and sets a new password
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *echo.Context) error {
	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	if err := h.service.ResetPassword(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Password reset successfully. You can now login with your new password", nil)
}
