package middleware

import (
	"net/http"

	"project-setup/internal/pkg/utils"

	"github.com/labstack/echo/v5"
)

// RequireRole returns a middleware that restricts access to users with specific roles.
// It must be used AFTER JWTMiddleware so that "role" is available in context.
//
// Usage:
//
//	r.GET("/admin-only", handler, middleware.RequireRole("ADMIN", "SUPERADMIN"))
func RequireRole(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			userRole, ok := c.Get("role").(string)
			if !ok || userRole == "" {
				return utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
			}

			for _, role := range allowedRoles {
				if userRole == role {
					return next(c)
				}
			}

			return utils.ErrorResponse(c, http.StatusForbidden, "Access denied. Required role(s): "+joinRoles(allowedRoles))
		}
	}
}

// joinRoles joins role names with commas for error messages
func joinRoles(roles []string) string {
	result := ""
	for i, role := range roles {
		if i > 0 {
			result += ", "
		}
		result += role
	}
	return result
}
