package utils

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Common weak passwords to reject
var commonPasswords = map[string]bool{
	"password":    true,
	"password123": true,
	"123456":      true,
	"12345678":    true,
	"qwerty":      true,
	"qwerty123":   true,
	"abc123":      true,
	"monkey":      true,
	"letmein":     true,
	"welcome":     true,
	"admin":       true,
	"admin123":    true,
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	// Validate password before hashing
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	if len(password) > 72 {
		return "", fmt.Errorf("password too long: maximum 72 bytes")
	}

	// Generate hash with default cost (12)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// ComparePasswords compares a hashed password with a plain text password
func ComparePasswords(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid password")
	}
	return nil
}

// ValidatePasswordStrength validates password meets strength requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 72 {
		return fmt.Errorf("password too long: maximum 72 bytes")
	}

	// Check for uppercase
	hasUpper := false
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	// Check for lowercase
	hasLower := false
	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			hasLower = true
			break
		}
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	// Check for number
	hasNumber := false
	for _, char := range password {
		if char >= '0' && char <= '9' {
			hasNumber = true
			break
		}
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	// Check for special character
	hasSpecial := false
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, char := range password {
		if strings.ContainsRune(specialChars, char) {
			hasSpecial = true
			break
		}
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Check against common passwords
	lowerPassword := strings.ToLower(password)
	if commonPasswords[lowerPassword] {
		return fmt.Errorf("password is too common, please choose a more unique password")
	}

	return nil
}
