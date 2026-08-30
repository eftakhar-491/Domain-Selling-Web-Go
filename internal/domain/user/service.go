package user

import (
	"errors"
	"math"

	"project-setup/internal/models"
)

// UserService handles business logic for user operations
type UserService struct {
	repo *UserRepository
}

// NewUserService creates a new UserService instance
func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetProfile retrieves the authenticated user's own profile
func (s *UserService) GetProfile(userID uint) (*ProfileResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &ProfileResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        string(user.Role),
		IsActive:    user.IsActive,
	}, nil
}

// GetAllUsers returns a paginated list of all users (ADMIN & SUPERADMIN only)
func (s *UserService) GetAllUsers(page, limit int) (*UserListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := s.repo.FindAll(page, limit)
	if err != nil {
		return nil, errors.New("failed to fetch users")
	}

	var userResponses []ProfileResponse
	for _, u := range users {
		userResponses = append(userResponses, ProfileResponse{
			ID:          u.ID,
			Name:        u.Name,
			Email:       u.Email,
			PhoneNumber: u.PhoneNumber,
			Role:        string(u.Role),
			IsActive:    u.IsActive,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &UserListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// UpdateProfile updates the authenticated user's own profile
func (s *UserService) UpdateProfile(userID uint, req UpdateProfileRequest) (*ProfileResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = req.PhoneNumber
	}

	if err := s.repo.Update(user); err != nil {
		return nil, errors.New("failed to update profile")
	}

	return &ProfileResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        string(user.Role),
		IsActive:    user.IsActive,
	}, nil
}

// UpdateRole changes a user's role (SUPERADMIN only)
func (s *UserService) UpdateRole(targetUserID uint, req UpdateRoleRequest) (*ProfileResponse, error) {
	user, err := s.repo.FindByID(targetUserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.Role = models.Role(req.Role)

	if err := s.repo.Update(user); err != nil {
		return nil, errors.New("failed to update role")
	}

	return &ProfileResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        string(user.Role),
		IsActive:    user.IsActive,
	}, nil
}

// DeleteUser removes a user from the system (ADMIN & SUPERADMIN only)
func (s *UserService) DeleteUser(targetUserID uint, requestingUserID uint) error {
	// Prevent self-deletion
	if targetUserID == requestingUserID {
		return errors.New("you cannot delete your own account")
	}

	user, err := s.repo.FindByID(targetUserID)
	if err != nil {
		return errors.New("user not found")
	}

	// Prevent deleting SUPERADMIN accounts
	if user.Role == models.RoleSuperAdmin {
		return errors.New("cannot delete a SUPERADMIN account")
	}

	return s.repo.Delete(targetUserID)
}
