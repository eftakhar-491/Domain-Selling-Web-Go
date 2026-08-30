package user

import (
	"net/http"
	"strconv"

	"project-setup/internal/pkg/utils"

	"github.com/labstack/echo/v5"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	service *UserService
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{service: service}
}

// GetProfile returns the authenticated user's own profile
// GET /api/v1/user/profile
func (h *UserHandler) GetProfile(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	profile, err := h.service.GetProfile(userID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusNotFound, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Profile retrieved successfully", profile)
}

// GetUsers returns a paginated list of all users (ADMIN & SUPERADMIN only)
// GET /api/v1/user?page=1&limit=10
func (h *UserHandler) GetUsers(c *echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	result, err := h.service.GetAllUsers(page, limit)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Users retrieved successfully", result)
}

// UpdateProfile updates the authenticated user's own profile
// PUT /api/v1/user/profile
func (h *UserHandler) UpdateProfile(c *echo.Context) error {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	profile, err := h.service.UpdateProfile(userID, req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Profile updated successfully", profile)
}

// UpdateRole changes a user's role (SUPERADMIN only)
// PUT /api/v1/user/:id/role
func (h *UserHandler) UpdateRole(c *echo.Context) error {
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
	}

	var req UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	profile, err := h.service.UpdateRole(uint(targetID), req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "Role updated successfully", profile)
}

// DeleteUser removes a user (ADMIN & SUPERADMIN only)
// DELETE /api/v1/user/:id
func (h *UserHandler) DeleteUser(c *echo.Context) error {
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
	}

	requestingUserID, ok := c.Get("user_id").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user session")
	}

	if err := h.service.DeleteUser(uint(targetID), requestingUserID); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
}
