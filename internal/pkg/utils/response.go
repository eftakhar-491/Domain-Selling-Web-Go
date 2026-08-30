package utils

import "github.com/labstack/echo/v5"

// APIResponse is a standardized JSON response structure
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// SuccessResponse sends a standardized success JSON response
func SuccessResponse(c *echo.Context, statusCode int, message string, data any) error {
	return c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse sends a standardized error JSON response
func ErrorResponse(c *echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
	})
}
