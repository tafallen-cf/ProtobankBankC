package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Email           string     `json:"email" db:"email"`
	Phone           string     `json:"phone" db:"phone"`
	PasswordHash    string     `json:"-" db:"password_hash"` // Never expose in JSON
	FirstName       string     `json:"first_name" db:"first_name"`
	LastName        string     `json:"last_name" db:"last_name"`
	DateOfBirth     time.Time  `json:"date_of_birth" db:"date_of_birth"`
	AddressLine1    string     `json:"address_line1" db:"address_line1"`
	AddressLine2    string     `json:"address_line2" db:"address_line2"`
	City            string     `json:"city" db:"city"`
	Postcode        string     `json:"postcode" db:"postcode"`
	Country         string     `json:"country" db:"country"`
	KYCStatus       string     `json:"kyc_status" db:"kyc_status"`
	KYCVerifiedAt   *time.Time `json:"kyc_verified_at" db:"kyc_verified_at"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email           string    `json:"email" binding:"required,email"`
	Phone           string    `json:"phone" binding:"required"`
	Password        string    `json:"password" binding:"required,min=8"`
	FirstName       string    `json:"first_name" binding:"required"`
	LastName        string    `json:"last_name" binding:"required"`
	DateOfBirth     time.Time `json:"date_of_birth" binding:"required"`
	AddressLine1    string    `json:"address_line1" binding:"required"`
	AddressLine2    string    `json:"address_line2"`
	City            string    `json:"city" binding:"required"`
	Postcode        string    `json:"postcode" binding:"required"`
	Country         string    `json:"country" binding:"required"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	DeviceID   string `json:"device_id"`
	DeviceType string `json:"device_type"`
}

// LoginResponse represents login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	User         *User  `json:"user"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse represents refresh token response
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"` // "access" or "refresh"
}
