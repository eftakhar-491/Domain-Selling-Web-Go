package find_domain

import (
	"github.com/labstack/echo/v5"
)

// DomainRoutes registers find_domain routes
func DomainRoutes(r *echo.Group) {
	service := NewDomainService()
	handler := NewDomainHandler(service)

	r.GET("/search", handler.Search)
	r.GET("/bulk-search", handler.BulkSearch)
}
