package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"project-setup/internal/models"
	"project-setup/internal/pkg/mail"
	"project-setup/internal/pkg/token"
	"project-setup/internal/pkg/utils"

	"github.com/redis/go-redis/v9"
)

// AuthService handles business logic for authentication
type AuthService struct {
	repo  *AuthRepository
	redis *redis.Client
}

// NewAuthService creates a new AuthService instance
func NewAuthService(repo *AuthRepository, redisClient *redis.Client) *AuthService {
	return &AuthService{repo: repo, redis: redisClient}
}

// Register creates a new user account with the USER role
func (s *AuthService) Register(req RegisterRequest) (*AuthResponse, error) {
	// Check if email already exists
	existingUser, _ := s.repo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user with default USER role
	user := &models.User{
		Name:        req.Name,
		Email:       req.Email,
		Password:    hashedPassword,
		PhoneNumber: req.PhoneNumber,
		Role:        models.RoleUser,
		IsActive:    true,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, errors.New("failed to create user: " + err.Error())
	}

	// Generate JWT token
	tokenString, err := token.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
		Token: tokenString,
		User: UserResponse{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			Role:        string(user.Role),
			IsActive:    user.IsActive,
		},
	}, nil
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
	// Find user by email
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated. Please contact an administrator")
	}

	// Verify password
	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	tokenString, err := token.GenerateToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &AuthResponse{
		Token: tokenString,
		User: UserResponse{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			Role:        string(user.Role),
			IsActive:    user.IsActive,
		},
	}, nil
}

// ForgotPassword generates a 6-digit OTP, stores it in Redis, and sends it via email
func (s *AuthService) ForgotPassword(req ForgotPasswordRequest) error {
	// Check if user exists
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil || user == nil {
		// Don't reveal if email exists or not (security best practice)
		return nil
	}

	// Generate 6-digit OTP code
	code, err := generateOTP(6)
	if err != nil {
		return errors.New("failed to generate reset code")
	}

	// Store OTP in Redis with 5-minute TTL
	if s.redis != nil {
		ctx := context.Background()
		redisKey := fmt.Sprintf("reset:%s", req.Email)
		if err := s.redis.Set(ctx, redisKey, code, 5*time.Minute).Err(); err != nil {
			return errors.New("failed to store reset code")
		}
	}

	// Send email with OTP
	if err := mail.SendResetCode(req.Email, code); err != nil {
		return errors.New("failed to send reset email: " + err.Error())
	}

	return nil
}

// ResetPassword validates the OTP from Redis and updates the user's password
func (s *AuthService) ResetPassword(req ResetPasswordRequest) error {
	if s.redis == nil {
		return errors.New("redis service not available")
	}

	ctx := context.Background()
	redisKey := fmt.Sprintf("reset:%s", req.Email)

	// Get OTP from Redis
	storedCode, err := s.redis.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		return errors.New("reset code has expired or is invalid")
	}
	if err != nil {
		return errors.New("failed to verify reset code")
	}

	// Validate the code
	if storedCode != req.Code {
		return errors.New("invalid reset code")
	}

	// Find the user
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return errors.New("user not found")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	// Update password
	user.Password = hashedPassword
	if err := s.repo.UpdateUser(user); err != nil {
		return errors.New("failed to update password")
	}

	// Delete OTP from Redis (one-time use)
	s.redis.Del(ctx, redisKey)

	return nil
}

// generateOTP creates a cryptographically secure N-digit numeric OTP
func generateOTP(length int) (string, error) {
	code := ""
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code, nil
}
