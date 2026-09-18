package user

// ProfileResponse represents the user profile data
type ProfileResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Role        string  `json:"role"`
	IsActive    bool    `json:"is_active"`
}

// UpdateProfileRequest represents the request body for updating own profile
type UpdateProfileRequest struct {
	Name        string  `json:"name" validate:"omitempty,min=2,max=100"`
	PhoneNumber *string `json:"phone_number,omitempty" validate:"omitempty,min=7,max=20"`
}

// UpdateRoleRequest represents the request body for changing a user's role (SUPERADMIN only)
type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=USER ADMIN SUPERADMIN"`
}

// UserListResponse represents paginated list of users
type UserListResponse struct {
	Users      []ProfileResponse `json:"users"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

// AdminUpdateUserRequest represents request for admin editing a user
type AdminUpdateUserRequest struct {
	Name        string  `json:"name" validate:"omitempty,min=2,max=100"`
	Email       string  `json:"email" validate:"omitempty,email"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
	Role        string  `json:"role,omitempty" validate:"omitempty,oneof=USER ADMIN SUPERADMIN"`
}

// AdminStatsResponse holds aggregated stats for the admin overview
type AdminStatsResponse struct {
	TotalUsers     int64   `json:"total_users"`
	ActiveUsers    int64   `json:"active_users"`
	TotalDomains   int64   `json:"total_domains"`
	ActiveDomains  int64   `json:"active_domains"`
	TotalOrders    int64   `json:"total_orders"`
	PaidOrders     int64   `json:"paid_orders"`
	PendingOrders  int64   `json:"pending_orders"`
	TotalRevenue   float64 `json:"total_revenue"`
	TotalDiscounts int64   `json:"total_discounts"`
}
