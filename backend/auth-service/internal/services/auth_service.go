package services

import (
	"context"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/protobankbankc/auth-service/internal/models"
	"github.com/protobankbankc/auth-service/internal/repository"
	"github.com/protobankbankc/auth-service/internal/utils"
	appErrors "github.com/protobankbankc/auth-service/pkg/errors"
)

// Common weak passwords to block
var commonPasswords = map[string]bool{
	"password":     true,
	"password123":  true,
	"12345678":     true,
	"qwerty":       true,
	"abc123":       true,
	"password1":    true,
	"password123!": true,
	"welcome":      true,
	"welcome123":   true,
	"admin":        true,
	"admin123":     true,
	"letmein":      true,
	"monkey":       true,
	"1234567890":   true,
}

// AuthService handles authentication business logic
type AuthService struct {
	userRepo             repository.UserRepository
	jwtSecret            string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repository.UserRepository,
	jwtSecret string,
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:             userRepo,
		jwtSecret:            jwtSecret,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Validate required fields
	if err := s.validateRegistrationRequest(req); err != nil {
		return nil, err
	}

	// Validate email format
	if err := s.validateEmail(req.Email); err != nil {
		return nil, err
	}

	// Validate password strength
	if err := s.validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, appErrors.NewConflict("user with this email already exists")
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user model
	user := &models.User{
		ID:           uuid.New(),
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: passwordHash,
		Phone:        req.Phone,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		DateOfBirth:  req.DateOfBirth,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		Region:       req.Region,
		Postcode:     req.Postcode,
		Country:      req.Country,
		IsActive:     true,
		KYCStatus:    "pending",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	// Save user to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Remove password hash before returning
	user.PasswordHash = ""

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.LoginResponse, error) {
	// Validate inputs
	if email == "" {
		return nil, appErrors.NewBadRequest("email is required")
	}
	if password == "" {
		return nil, appErrors.NewBadRequest("password is required")
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		// Don't reveal if user exists or not
		return nil, appErrors.NewUnauthorized("invalid email or password")
	}

	// Check if account is active
	if !user.IsActive {
		return nil, appErrors.NewForbidden("account is inactive")
	}

	// Verify password
	if err := utils.ComparePassword(user.PasswordHash, password); err != nil {
		return nil, appErrors.NewUnauthorized("invalid email or password")
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Remove password hash before returning
	user.PasswordHash = ""

	return &models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.accessTokenDuration.Seconds()),
		User:         user,
	}, nil
}

// RefreshToken validates a refresh token and issues a new access token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.RefreshTokenResponse, error) {
	// Validate input
	if refreshToken == "" {
		return nil, appErrors.NewBadRequest("refresh token is required")
	}

	// Validate refresh token
	claims, err := utils.ValidateToken(refreshToken, s.jwtSecret)
	if err != nil {
		return nil, appErrors.NewUnauthorized("invalid or expired refresh token")
	}

	// Verify it's a refresh token
	if claims.Type != "refresh" {
		return nil, appErrors.NewUnauthorized("invalid token type")
	}

	// Parse user ID
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, appErrors.NewUnauthorized("invalid user ID in token")
	}

	// Get user from database to verify they still exist and are active
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, appErrors.NewNotFound("user not found")
	}

	// Check if account is active
	if !user.IsActive {
		return nil, appErrors.NewForbidden("account is inactive")
	}

	// Generate new access token
	accessToken, err := s.generateAccessToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &models.RefreshTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.accessTokenDuration.Seconds()),
	}, nil
}

// ValidateAccessToken validates an access token and returns the user
func (s *AuthService) ValidateAccessToken(ctx context.Context, accessToken string) (*models.User, error) {
	// Validate input
	if accessToken == "" {
		return nil, appErrors.NewBadRequest("access token is required")
	}

	// Validate token
	claims, err := utils.ValidateToken(accessToken, s.jwtSecret)
	if err != nil {
		return nil, appErrors.NewUnauthorized("invalid or expired access token")
	}

	// Verify it's an access token
	if claims.Type != "access" {
		return nil, appErrors.NewUnauthorized("invalid token type")
	}

	// Parse user ID
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, appErrors.NewUnauthorized("invalid user ID in token")
	}

	// Get user from database
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, appErrors.NewNotFound("user not found")
	}

	// Check if account is active
	if !user.IsActive {
		return nil, appErrors.NewForbidden("account is inactive")
	}

	// Remove password hash before returning
	user.PasswordHash = ""

	return user, nil
}

// validateRegistrationRequest validates all required fields
func (s *AuthService) validateRegistrationRequest(req *models.RegisterRequest) error {
	if req.Email == "" {
		return appErrors.NewBadRequest("email is required")
	}
	if req.Password == "" {
		return appErrors.NewBadRequest("password is required")
	}
	if req.FirstName == "" {
		return appErrors.NewBadRequest("first name is required")
	}
	if req.LastName == "" {
		return appErrors.NewBadRequest("last name is required")
	}
	if req.DateOfBirth.IsZero() {
		return appErrors.NewBadRequest("date of birth is required")
	}
	if req.AddressLine1 == "" {
		return appErrors.NewBadRequest("address line 1 is required")
	}
	if req.City == "" {
		return appErrors.NewBadRequest("city is required")
	}
	if req.Postcode == "" {
		return appErrors.NewBadRequest("postcode is required")
	}
	if req.Country == "" {
		return appErrors.NewBadRequest("country is required")
	}

	// Validate age (must be 18+)
	age := time.Now().Year() - req.DateOfBirth.Year()
	if age < 18 {
		return appErrors.NewBadRequest("you must be at least 18 years old to register")
	}
	// More precise age calculation
	if time.Now().YearDay() < req.DateOfBirth.YearDay() {
		age--
	}
	if age < 18 {
		return appErrors.NewBadRequest("you must be at least 18 years old to register")
	}

	return nil
}

// validateEmail validates email format
func (s *AuthService) validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return appErrors.NewBadRequest("email cannot be empty")
	}

	// Use Go's mail.ParseAddress for robust email validation
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return appErrors.NewBadRequest("invalid email format")
	}

	// Additional validation: check for spaces
	if strings.Contains(addr.Address, " ") {
		return appErrors.NewBadRequest("invalid email format")
	}

	// Check email length
	if len(addr.Address) > 254 {
		return appErrors.NewBadRequest("email too long")
	}

	return nil
}

// validatePassword validates password strength
func (s *AuthService) validatePassword(password string) error {
	if len(password) < 8 {
		return appErrors.NewBadRequest("password must be at least 8 characters long")
	}

	if len(password) > 72 {
		return appErrors.NewBadRequest("password too long: maximum 72 characters")
	}

	// Check for uppercase letter
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	if !hasUpper {
		return appErrors.NewBadRequest("password must contain at least one uppercase letter")
	}

	// Check for lowercase letter
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	if !hasLower {
		return appErrors.NewBadRequest("password must contain at least one lowercase letter")
	}

	// Check for number
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	if !hasNumber {
		return appErrors.NewBadRequest("password must contain at least one number")
	}

	// Check for special character
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	if !hasSpecial {
		return appErrors.NewBadRequest("password must contain at least one special character")
	}

	// Check against common passwords
	lowerPassword := strings.ToLower(password)
	if commonPasswords[lowerPassword] {
		return appErrors.NewBadRequest("password is too common, please choose a stronger password")
	}

	return nil
}

// generateAccessToken generates a JWT access token
func (s *AuthService) generateAccessToken(userID, email string) (string, error) {
	return utils.GenerateAccessToken(userID, email, s.accessTokenDuration, s.jwtSecret)
}

// generateRefreshToken generates a JWT refresh token
func (s *AuthService) generateRefreshToken(userID, email string) (string, error) {
	return utils.GenerateRefreshToken(userID, email, s.refreshTokenDuration, s.jwtSecret)
}
