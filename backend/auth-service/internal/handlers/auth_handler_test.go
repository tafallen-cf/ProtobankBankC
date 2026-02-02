package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/protobankbankc/auth-service/internal/models"
	appErrors "github.com/protobankbankc/auth-service/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthService mocks the auth service interface
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (*models.LoginResponse, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.RefreshTokenResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RefreshTokenResponse), args.Error(1)
}

func (m *MockAuthService) ValidateAccessToken(ctx context.Context, accessToken string) (*models.User, error) {
	args := m.Called(ctx, accessToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// setupTestRouter creates a test router with Gin
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestRegisterHandler tests the register endpoint
func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful registration",
			requestBody: models.RegisterRequest{
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
			setupMock: func(m *MockAuthService) {
				user := &models.User{
					ID:        uuid.New(),
					Email:     "john.doe@example.com",
					FirstName: "John",
					LastName:  "Doe",
					IsActive:  true,
					KYCStatus: "pending",
				}
				m.On("Register", mock.Anything, mock.AnythingOfType("*models.RegisterRequest")).Return(user, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, "user registered successfully", response["message"])
				assert.NotNil(t, response["user"])
			},
		},
		{
			name: "invalid request body",
			requestBody: map[string]string{
				"invalid": "data",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "invalid")
			},
		},
		{
			name: "user already exists",
			requestBody: models.RegisterRequest{
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
			setupMock: func(m *MockAuthService) {
				m.On("Register", mock.Anything, mock.AnythingOfType("*models.RegisterRequest")).
					Return(nil, appErrors.NewConflict("user already exists"))
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "already exists")
			},
		},
		{
			name: "weak password",
			requestBody: models.RegisterRequest{
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
			setupMock: func(m *MockAuthService) {
				m.On("Register", mock.Anything, mock.AnythingOfType("*models.RegisterRequest")).
					Return(nil, appErrors.NewBadRequest("password does not meet strength requirements"))
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "password")
			},
		},
		{
			name:           "malformed JSON",
			requestBody:    `{invalid json}`,
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockAuthService)
			tt.setupMock(mockService)
			handler := NewAuthHandler(mockService)
			router := setupTestRouter()
			router.POST("/auth/register", handler.Register)

			// Create request
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			tt.checkResponse(t, rec)
			mockService.AssertExpectations(t)
		})
	}
}

// TestLoginHandler tests the login endpoint
func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful login",
			requestBody: models.LoginRequest{
				Email:    "john.doe@example.com",
				Password: "SecurePass123!",
			},
			setupMock: func(m *MockAuthService) {
				response := &models.LoginResponse{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					TokenType:    "Bearer",
					ExpiresIn:    900,
					User: &models.User{
						ID:        uuid.New(),
						Email:     "john.doe@example.com",
						FirstName: "John",
						LastName:  "Doe",
					},
				}
				m.On("Login", mock.Anything, "john.doe@example.com", "SecurePass123!").Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.LoginResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Equal(t, "Bearer", response.TokenType)
				assert.NotNil(t, response.User)
			},
		},
		{
			name: "invalid credentials",
			requestBody: models.LoginRequest{
				Email:    "john.doe@example.com",
				Password: "WrongPassword",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Login", mock.Anything, "john.doe@example.com", "WrongPassword").
					Return(nil, appErrors.NewUnauthorized("invalid email or password"))
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "invalid")
			},
		},
		{
			name: "missing email",
			requestBody: models.LoginRequest{
				Password: "SecurePass123!",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response["error"])
			},
		},
		{
			name: "inactive user",
			requestBody: models.LoginRequest{
				Email:    "inactive@example.com",
				Password: "SecurePass123!",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Login", mock.Anything, "inactive@example.com", "SecurePass123!").
					Return(nil, appErrors.NewForbidden("account is inactive"))
			},
			expectedStatus: http.StatusForbidden,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "inactive")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockAuthService)
			tt.setupMock(mockService)
			handler := NewAuthHandler(mockService)
			router := setupTestRouter()
			router.POST("/auth/login", handler.Login)

			// Create request
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			tt.checkResponse(t, rec)
			mockService.AssertExpectations(t)
		})
	}
}

// TestRefreshTokenHandler tests the refresh token endpoint
func TestRefreshTokenHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful token refresh",
			requestBody: models.RefreshTokenRequest{
				RefreshToken: "valid-refresh-token",
			},
			setupMock: func(m *MockAuthService) {
				response := &models.RefreshTokenResponse{
					AccessToken: "new-access-token",
					TokenType:   "Bearer",
					ExpiresIn:   900,
				}
				m.On("RefreshToken", mock.Anything, "valid-refresh-token").Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response models.RefreshTokenResponse
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.AccessToken)
				assert.Equal(t, "Bearer", response.TokenType)
				assert.Greater(t, response.ExpiresIn, 0)
			},
		},
		{
			name: "invalid refresh token",
			requestBody: models.RefreshTokenRequest{
				RefreshToken: "invalid-token",
			},
			setupMock: func(m *MockAuthService) {
				m.On("RefreshToken", mock.Anything, "invalid-token").
					Return(nil, appErrors.NewUnauthorized("invalid or expired refresh token"))
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "invalid")
			},
		},
		{
			name: "missing refresh token",
			requestBody: models.RefreshTokenRequest{
				RefreshToken: "",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockAuthService)
			tt.setupMock(mockService)
			handler := NewAuthHandler(mockService)
			router := setupTestRouter()
			router.POST("/auth/refresh", handler.RefreshToken)

			// Create request
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			tt.checkResponse(t, rec)
			mockService.AssertExpectations(t)
		})
	}
}

// TestGetMeHandler tests the /auth/me endpoint
func TestGetMeHandler(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		setupMock      func(*MockAuthService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "successful user retrieval",
			authHeader: "Bearer valid-access-token",
			setupMock: func(m *MockAuthService) {
				user := &models.User{
					ID:        uuid.New(),
					Email:     "john.doe@example.com",
					FirstName: "John",
					LastName:  "Doe",
					IsActive:  true,
				}
				m.On("ValidateAccessToken", mock.Anything, "valid-access-token").Return(user, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var user models.User
				err := json.Unmarshal(rec.Body.Bytes(), &user)
				require.NoError(t, err)
				assert.NotEmpty(t, user.ID)
				assert.NotEmpty(t, user.Email)
			},
		},
		{
			name:       "missing authorization header",
			authHeader: "",
			setupMock:  func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "authorization")
			},
		},
		{
			name:       "invalid token format",
			authHeader: "InvalidFormat token",
			setupMock:  func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response["error"])
			},
		},
		{
			name:       "expired token",
			authHeader: "Bearer expired-token",
			setupMock: func(m *MockAuthService) {
				m.On("ValidateAccessToken", mock.Anything, "expired-token").
					Return(nil, appErrors.NewUnauthorized("token has expired"))
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "expired")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockAuthService)
			tt.setupMock(mockService)
			handler := NewAuthHandler(mockService)
			router := setupTestRouter()
			router.GET("/auth/me", handler.GetMe)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
			tt.checkResponse(t, rec)
			mockService.AssertExpectations(t)
		})
	}
}

// TestErrorHandling tests error response formatting
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		serviceError error
		expectedCode int
	}{
		{
			name:         "bad request error",
			serviceError: appErrors.NewBadRequest("invalid input"),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "unauthorized error",
			serviceError: appErrors.NewUnauthorized("invalid credentials"),
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "forbidden error",
			serviceError: appErrors.NewForbidden("access denied"),
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "not found error",
			serviceError: appErrors.NewNotFound("user not found"),
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "conflict error",
			serviceError: appErrors.NewConflict("user already exists"),
			expectedCode: http.StatusConflict,
		},
		{
			name:         "generic error",
			serviceError: errors.New("some internal error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockAuthService)
			mockService.On("Login", mock.Anything, mock.Anything, mock.Anything).Return(nil, tt.serviceError)
			handler := NewAuthHandler(mockService)
			router := setupTestRouter()
			router.POST("/auth/login", handler.Login)

			// Create request
			body, _ := json.Marshal(models.LoginRequest{
				Email:    "test@example.com",
				Password: "password",
			})
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tt.expectedCode, rec.Code)
			var response map[string]interface{}
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.NotEmpty(t, response["error"])
		})
	}
}
