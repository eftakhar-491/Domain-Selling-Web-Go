package middleware

import (
	"net/http"
	"strings"

	"project-setup/internal/pkg/token"
	"project-setup/internal/pkg/utils"

	"github.com/labstack/echo/v5"
)

// JWTMiddleware validates the JWT token from the Authorization header
// and sets user claims (user_id, email, role) into the Echo context
func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization header is required")
		}

		// Expect format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid authorization header format. Use: Bearer <token>")
		}

		tokenString := parts[1]
		claims, err := token.ValidateToken(tokenString)
		if err != nil {
			return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired token: "+err.Error())
		}

		// Set user information in context for downstream handlers
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		return next(c)
	}
}
