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
