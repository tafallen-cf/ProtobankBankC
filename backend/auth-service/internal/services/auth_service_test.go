package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/protobankbankc/auth-service/internal/models"
	appErrors "github.com/protobankbankc/auth-service/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock UserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateKYCStatus(ctx context.Context, id uuid.UUID, status string, verifiedAt *time.Time) error {
	args := m.Called(ctx, id, status, verifiedAt)
	return args.Error(0)
}

func (m *MockUserRepository) SetInactive(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestRegister tests user registration
func TestRegister(t *testing.T) {
	jwtSecret := "test-secret-key-at-least-32-chars-long-for-security"

	tests := []struct {
		name          string
		request       *models.RegisterRequest
		setupMock     func(*MockUserRepository)
		wantErr       bool
		errType       error
		errContains   string
	}{
		{
			name: "successful registration",
			request: &models.RegisterRequest{
				Email:        "john.doe@example.com",
				Phone:        "+447700900123",
				Password:     "SecurePass123!",
				FirstName:    "John",
				LastName:     "Doe",
				DateOfBirth:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
				AddressLine1: "123 Main St",
				City:         "London",
				Postcode:     "SW1A 1AA",
				Country:      "UK",
			},
			setupMock: func(repo *MockUserRepository) {
				repo.On("GetByEmail", mock.Anything, "john.doe@example.com").Return(nil, appErrors.NewNotFound("user not found"))
				repo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "email already exists",
			request: &models.RegisterRequest{
				Email:        "existing@example.com",
				Phone:        "+447700900123",
				Password:     "SecurePass123!",
				FirstName:    "John",
				LastName:     "Doe",
				DateOfBirth:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
				AddressLine1: "123 Main St",
				City:         "London",
				Postcode:     "SW1A 1AA",
				Country:      "UK",
			},
			setupMock: func(repo *MockUserRepository) {
				existingUser := &models.User{
					ID:    uuid.New(),
					Email: "existing@example.com",
				}
				repo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			wantErr:     true,
			errType:     appErrors.ErrUserAlreadyExists,
			errContains: "already exists",
		},
		{
			name: "invalid email format",
			request: &models.RegisterRequest{
				Email:        "not-an-email",
				Phone:        "+447700900123",
				Password:     "SecurePass123!",
				FirstName:    "John",
				LastName:     "Doe",
				DateOfBirth:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
				AddressLine1: "123 Main St",
				City:         "London",
				Postcode:     "SW1A 1AA",
				Country:      "UK",
			},
			setupMock:   func(repo *MockUserRepository) {},
			wantErr:     true,
			errType:     appErrors.ErrInvalidEmail,
			errContains: "invalid email",
		},
		{
			name: "weak password",
			request: &models.RegisterRequest{
				Email:        "john.doe@example.com",
				Phone:        "+447700900123",
				Password:     "weak",
				FirstName:    "John",
				LastName:     "Doe",
				DateOfBirth:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
				AddressLine1: "123 Main St",
				City:         "London",
				Postcode:     "SW1A 1AA",
				Country:      "UK",
			},
			setupMock:   func(repo *MockUserRepository) {},
			wantErr:     true,
			errType:     appErrors.ErrWeakPassword,
			errContains: "password",
		},
		{
			name: "missing required fields",
			request: &models.RegisterRequest{
				Email:    "john.doe@example.com",
				Password: "SecurePass123!",
				// Missing other required fields
			},
			setupMock:   func(repo *MockUserRepository) {},
			wantErr:     true,
			errType:     appErrors.ErrInvalidInput,
			errContains: "required",
		},
		{
			name: "underage user",
			request: &models.RegisterRequest{
				Email:        "teen@example.com",
				Phone:        "+447700900123",
				Password:     "SecurePass123!",
				FirstName:    "John",
				LastName:     "Doe",
				DateOfBirth:  time.Now().AddDate(-15, 0, 0), // 15 years old
				AddressLine1: "123 Main St",
				City:         "London",
				Postcode:     "SW1A 1AA",
				Country:      "UK",
			},
			setupMock:   func(repo *MockUserRepository) {},
			wantErr:     true,
			errType:     appErrors.ErrInvalidInput,
			errContains: "18 years",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)

			service := NewAuthService(mockRepo, jwtSecret, 15*time.Minute, 7*24*time.Hour)
			ctx := context.Background()

			user, err := service.Register(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				if tt.errType != nil {
					assert.True(t, errors.Is(err, tt.errType))
				}
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.request.Email, user.Email)
				assert.Equal(t, tt.request.FirstName, user.FirstName)
				assert.Equal(t, tt.request.LastName, user.LastName)
				assert.NotEmpty(t, user.ID)
				assert.True(t, user.IsActive)
				assert.Equal(t, "pending", user.KYCStatus)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestLogin tests user login
func TestLogin(t *testing.T) {
	jwtSecret := "test-secret-key-at-least-32-chars-long-for-security"

	tests := []struct {
		name        string
		email       string
		password    string
		setupMock   func(*MockUserRepository)
		wantErr     bool
		errType     error
		errContains string
	}{
		{
			name:     "successful login",
			email:    "john.doe@example.com",
			password: "SecurePass123!",
			setupMock: func(repo *MockUserRepository) {
				// Password hash for "SecurePass123!"
				user := &models.User{
					ID:           uuid.New(),
					Email:        "john.doe@example.com",
					PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYFJ5NQjeFi", // bcrypt hash
					FirstName:    "John",
					LastName:     "Doe",
					IsActive:     true,
					KYCStatus:    "verified",
				}
				repo.On("GetByEmail", mock.Anything, "john.doe@example.com").Return(user, nil)
			},
			wantErr: false,
		},
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "SecurePass123!",
			setupMock: func(repo *MockUserRepository) {
				repo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, appErrors.NewNotFound("user not found"))
			},
			wantErr:     true,
			errType:     appErrors.ErrInvalidCredentials,
			errContains: "invalid email or password",
		},
		{
			name:     "incorrect password",
			email:    "john.doe@example.com",
			password: "WrongPassword123!",
			setupMock: func(repo *MockUserRepository) {
				user := &models.User{
					ID:           uuid.New(),
					Email:        "john.doe@example.com",
					PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYFJ5NQjeFi",
					IsActive:     true,
				}
				repo.On("GetByEmail", mock.Anything, "john.doe@example.com").Return(user, nil)
			},
			wantErr:     true,
			errType:     appErrors.ErrInvalidCredentials,
			errContains: "invalid email or password",
		},
		{
			name:     "inactive user account",
			email:    "inactive@example.com",
			password: "SecurePass123!",
			setupMock: func(repo *MockUserRepository) {
				user := &models.User{
					ID:           uuid.New(),
					Email:        "inactive@example.com",
					PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYFJ5NQjeFi",
					IsActive:     false,
				}
				repo.On("GetByEmail", mock.Anything, "inactive@example.com").Return(user, nil)
			},
			wantErr:     true,
			errType:     appErrors.ErrUserInactive,
			errContains: "inactive",
		},
		{
			name:        "empty email",
			email:       "",
			password:    "SecurePass123!",
			setupMock:   func(repo *MockUserRepository) {},
			wantErr:     true,
			errType:     appErrors.ErrInvalidInput,
			errContains: "email",
		},
		{
			name:        "empty password",
			email:       "john.doe@example.com",
			password:    "",
			setupMock:   func(repo *MockUserRepository) {},
			wantErr:     true,
			errType:     appErrors.ErrInvalidInput,
			errContains: "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)

			service := NewAuthService(mockRepo, jwtSecret, 15*time.Minute, 7*24*time.Hour)
			ctx := context.Background()

			response, err := service.Login(ctx, tt.email, tt.password)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				if tt.errType != nil {
					assert.True(t, errors.Is(err, tt.errType))
				}
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Equal(t, "Bearer", response.TokenType)
				assert.Greater(t, response.ExpiresIn, 0)
				assert.NotNil(t, response.User)
				assert.Equal(t, tt.email, response.User.Email)
				// Password hash should not be exposed
				assert.Empty(t, response.User.PasswordHash)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestRefreshToken tests token refresh
func TestRefreshToken(t *testing.T) {
	jwtSecret := "test-secret-key-at-least-32-chars-long-for-security"

	tests := []struct {
		name         string
		refreshToken string
		setupMock    func(*MockUserRepository)
		wantErr      bool
		errType      error
		errContains  string
	}{
		{
			name:         "successful token refresh",
			refreshToken: "valid-refresh-token",
			setupMock: func(repo *MockUserRepository) {
				user := &models.User{
					ID:        uuid.New(),
					Email:     "john.doe@example.com",
					IsActive:  true,
					KYCStatus: "verified",
				}
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(user, nil)
			},
			wantErr: false,
		},
		{
			name:         "empty refresh token",
			refreshToken: "",
			setupMock:    func(repo *MockUserRepository) {},
			wantErr:      true,
			errType:      appErrors.ErrInvalidInput,
			errContains:  "token",
		},
		{
			name:         "invalid refresh token",
			refreshToken: "invalid-token",
			setupMock:    func(repo *MockUserRepository) {},
			wantErr:      true,
			errType:      appErrors.ErrTokenInvalid,
			errContains:  "invalid",
		},
		{
			name:         "user not found",
			refreshToken: "valid-refresh-token",
			setupMock: func(repo *MockUserRepository) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, appErrors.NewNotFound("user not found"))
			},
			wantErr:     true,
			errType:     appErrors.ErrUserNotFound,
			errContains: "not found",
		},
		{
			name:         "inactive user",
			refreshToken: "valid-refresh-token",
			setupMock: func(repo *MockUserRepository) {
				user := &models.User{
					ID:       uuid.New(),
					Email:    "john.doe@example.com",
					IsActive: false,
				}
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(user, nil)
			},
			wantErr:     true,
			errType:     appErrors.ErrUserInactive,
			errContains: "inactive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)

			service := NewAuthService(mockRepo, jwtSecret, 15*time.Minute, 7*24*time.Hour)
			ctx := context.Background()

			// For successful test, generate a real refresh token
			if !tt.wantErr && tt.name == "successful token refresh" {
				// We need to get the user ID from the mock
				mockRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(func(_ context.Context, id uuid.UUID) *models.User {
					return &models.User{
						ID:       id,
						Email:    "john.doe@example.com",
						IsActive: true,
					}
				}, nil)

				// Generate a valid refresh token for testing
				testUserID := uuid.New()
				testToken, err := service.generateRefreshToken(testUserID.String(), "john.doe@example.com")
				require.NoError(t, err)
				tt.refreshToken = testToken
			}

			response, err := service.RefreshToken(ctx, tt.refreshToken)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				if tt.errType != nil {
					assert.True(t, errors.Is(err, tt.errType))
				}
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.NotEmpty(t, response.AccessToken)
				assert.Equal(t, "Bearer", response.TokenType)
				assert.Greater(t, response.ExpiresIn, 0)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestValidateAccessToken tests access token validation
func TestValidateAccessToken(t *testing.T) {
	jwtSecret := "test-secret-key-at-least-32-chars-long-for-security"

	tests := []struct {
		name        string
		token       string
		setupMock   func(*MockUserRepository)
		setupToken  func(*AuthService) string
		wantErr     bool
		errType     error
		errContains string
	}{
		{
			name: "valid access token",
			setupToken: func(service *AuthService) string {
				userID := uuid.New()
				token, _ := service.generateAccessToken(userID.String(), "john.doe@example.com")
				return token
			},
			setupMock: func(repo *MockUserRepository) {
				user := &models.User{
					ID:       uuid.New(),
					Email:    "john.doe@example.com",
					IsActive: true,
				}
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(user, nil)
			},
			wantErr: false,
		},
		{
			name:        "empty token",
			token:       "",
			setupMock:   func(repo *MockUserRepository) {},
			wantErr:     true,
			errType:     appErrors.ErrInvalidInput,
			errContains: "token",
		},
		{
			name:        "invalid token format",
			token:       "invalid-token",
			setupMock:   func(repo *MockUserRepository) {},
			wantErr:     true,
			errType:     appErrors.ErrTokenInvalid,
			errContains: "invalid",
		},
		{
			name: "user not found",
			setupToken: func(service *AuthService) string {
				userID := uuid.New()
				token, _ := service.generateAccessToken(userID.String(), "john.doe@example.com")
				return token
			},
			setupMock: func(repo *MockUserRepository) {
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, appErrors.NewNotFound("user not found"))
			},
			wantErr:     true,
			errType:     appErrors.ErrUserNotFound,
			errContains: "not found",
		},
		{
			name: "inactive user",
			setupToken: func(service *AuthService) string {
				userID := uuid.New()
				token, _ := service.generateAccessToken(userID.String(), "john.doe@example.com")
				return token
			},
			setupMock: func(repo *MockUserRepository) {
				user := &models.User{
					ID:       uuid.New(),
					Email:    "john.doe@example.com",
					IsActive: false,
				}
				repo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(user, nil)
			},
			wantErr:     true,
			errType:     appErrors.ErrUserInactive,
			errContains: "inactive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.setupMock(mockRepo)

			service := NewAuthService(mockRepo, jwtSecret, 15*time.Minute, 7*24*time.Hour)
			ctx := context.Background()

			token := tt.token
			if tt.setupToken != nil {
				token = tt.setupToken(service)
			}

			user, err := service.ValidateAccessToken(ctx, token)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				if tt.errType != nil {
					assert.True(t, errors.Is(err, tt.errType))
				}
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.NotEmpty(t, user.ID)
				assert.NotEmpty(t, user.Email)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestPasswordValidation tests password validation logic
func TestPasswordValidation(t *testing.T) {
	jwtSecret := "test-secret-key-at-least-32-chars-long-for-security"
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, jwtSecret, 15*time.Minute, 7*24*time.Hour)

	tests := []struct {
		name        string
		password    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid strong password",
			password: "SecurePass123!",
			wantErr:  false,
		},
		{
			name:        "too short",
			password:    "Short1!",
			wantErr:     true,
			errContains: "at least 8 characters",
		},
		{
			name:        "no uppercase",
			password:    "securepass123!",
			wantErr:     true,
			errContains: "uppercase",
		},
		{
			name:        "no lowercase",
			password:    "SECUREPASS123!",
			wantErr:     true,
			errContains: "lowercase",
		},
		{
			name:        "no numbers",
			password:    "SecurePassword!",
			wantErr:     true,
			errContains: "number",
		},
		{
			name:        "no special characters",
			password:    "SecurePass123",
			wantErr:     true,
			errContains: "special character",
		},
		{
			name:        "common password",
			password:    "Password123!",
			wantErr:     true,
			errContains: "common",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validatePassword(tt.password)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestEmailValidation tests email validation logic
func TestEmailValidation(t *testing.T) {
	jwtSecret := "test-secret-key-at-least-32-chars-long-for-security"
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, jwtSecret, 15*time.Minute, 7*24*time.Hour)

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "valid email",
			email:   "john.doe@example.com",
			wantErr: false,
		},
		{
			name:    "valid email with subdomain",
			email:   "user@mail.example.com",
			wantErr: false,
		},
		{
			name:    "valid email with plus",
			email:   "user+tag@example.com",
			wantErr: false,
		},
		{
			name:    "missing @",
			email:   "notanemail.com",
			wantErr: true,
		},
		{
			name:    "missing domain",
			email:   "user@",
			wantErr: true,
		},
		{
			name:    "missing local part",
			email:   "@example.com",
			wantErr: true,
		},
		{
			name:    "empty email",
			email:   "",
			wantErr: true,
		},
		{
			name:    "spaces in email",
			email:   "user name@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateEmail(tt.email)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
