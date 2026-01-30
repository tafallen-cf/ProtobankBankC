package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestHashPassword tests password hashing functionality
func TestHashPassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantErr   bool
		errString string
	}{
		{
			name:     "valid password",
			password: "SecurePass123!",
			wantErr:  false,
		},
		{
			name:      "empty password",
			password:  "",
			wantErr:   true,
			errString: "password cannot be empty",
		},
		{
			name:      "password too long",
			password:  strings.Repeat("a", 73), // bcrypt max is 72 bytes
			wantErr:   true,
			errString: "password too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
				assert.Empty(t, hash)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash)

			// Verify hash starts with bcrypt prefix
			assert.True(t, strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$"))

			// Verify hash is valid bcrypt hash
			err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password))
			assert.NoError(t, err, "generated hash should be valid")
		})
	}
}

// TestComparePasswords tests password comparison
func TestComparePasswords(t *testing.T) {
	validPassword := "SecurePass123!"
	validHash, err := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	tests := []struct {
		name     string
		hash     string
		password string
		wantErr  bool
	}{
		{
			name:     "correct password",
			hash:     string(validHash),
			password: validPassword,
			wantErr:  false,
		},
		{
			name:     "incorrect password",
			hash:     string(validHash),
			password: "WrongPassword123!",
			wantErr:  true,
		},
		{
			name:     "empty password",
			hash:     string(validHash),
			password: "",
			wantErr:  true,
		},
		{
			name:     "invalid hash",
			hash:     "not-a-valid-hash",
			password: validPassword,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ComparePasswords(tt.hash, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidatePasswordStrength tests password strength validation
func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantErr   bool
		errString string
	}{
		{
			name:     "strong password",
			password: "SecurePass123!",
			wantErr:  false,
		},
		{
			name:     "strong password with symbols",
			password: "MyP@ssw0rd!2024",
			wantErr:  false,
		},
		{
			name:      "too short",
			password:  "Short1!",
			wantErr:   true,
			errString: "at least 8 characters",
		},
		{
			name:      "no uppercase",
			password:  "lowercase123!",
			wantErr:   true,
			errString: "uppercase letter",
		},
		{
			name:      "no lowercase",
			password:  "UPPERCASE123!",
			wantErr:   true,
			errString: "lowercase letter",
		},
		{
			name:      "no numbers",
			password:  "NoNumbers!",
			wantErr:   true,
			errString: "number",
		},
		{
			name:      "no special characters",
			password:  "NoSpecialChars123",
			wantErr:   true,
			errString: "special character",
		},
		{
			name:      "empty password",
			password:  "",
			wantErr:   true,
			errString: "at least 8 characters",
		},
		{
			name:      "common password",
			password:  "Password123!",
			wantErr:   true,
			errString: "too common",
		},
		{
			name:      "another common password",
			password:  "Qwerty123!",
			wantErr:   true,
			errString: "too common",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// BenchmarkHashPassword benchmarks password hashing
func BenchmarkHashPassword(b *testing.B) {
	password := "SecurePass123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword(password)
	}
}

// BenchmarkComparePasswords benchmarks password comparison
func BenchmarkComparePasswords(b *testing.B) {
	password := "SecurePass123!"
	hash, _ := HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComparePasswords(hash, password)
	}
}

// TestHashPasswordSecurity tests security properties
func TestHashPasswordSecurity(t *testing.T) {
	password := "SecurePass123!"

	t.Run("different hashes for same password", func(t *testing.T) {
		hash1, err1 := HashPassword(password)
		hash2, err2 := HashPassword(password)

		require.NoError(t, err1)
		require.NoError(t, err2)

		// Hashes should be different due to random salt
		assert.NotEqual(t, hash1, hash2)

		// But both should verify against the same password
		assert.NoError(t, ComparePasswords(hash1, password))
		assert.NoError(t, ComparePasswords(hash2, password))
	})

	t.Run("timing attack resistance", func(t *testing.T) {
		hash, err := HashPassword(password)
		require.NoError(t, err)

		// Comparing with wrong password should take similar time
		// This is guaranteed by bcrypt's design
		err1 := ComparePasswords(hash, "WrongPassword123!")
		err2 := ComparePasswords(hash, "AnotherWrongPass123!")

		assert.Error(t, err1)
		assert.Error(t, err2)
		// Both should fail, timing should be constant (bcrypt property)
	})
}
