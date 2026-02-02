package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/protobankbankc/auth-service/internal/models"
	appErrors "github.com/protobankbankc/auth-service/pkg/errors"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *models.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// GetByPhone retrieves a user by phone
	GetByPhone(ctx context.Context, phone string) (*models.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *models.User) error

	// Delete deletes a user by ID
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateKYCStatus updates the KYC status for a user
	UpdateKYCStatus(ctx context.Context, id uuid.UUID, status string, verifiedAt *time.Time) error

	// SetInactive sets a user as inactive
	SetInactive(ctx context.Context, id uuid.UUID) error
}

// userRepository implements UserRepository
type userRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, email, phone, password_hash, first_name, last_name,
			date_of_birth, address_line1, address_line2, city, postcode, country,
			kyc_status, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	now := time.Now()
	user.ID = uuid.New()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.IsActive = true
	user.KYCStatus = "pending"

	_, err := r.db.Exec(ctx, query,
		user.ID, user.Email, user.Phone, user.PasswordHash,
		user.FirstName, user.LastName, user.DateOfBirth,
		user.AddressLine1, user.AddressLine2, user.City, user.Postcode, user.Country,
		user.KYCStatus, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)

	if err != nil {
		if isPgError(err, "23505") { // Unique violation
			return appErrors.NewConflict("user with this email or phone already exists")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, phone, password_hash, first_name, last_name,
			   date_of_birth, address_line1, address_line2, city, postcode, country,
			   kyc_status, kyc_verified_at, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.DateOfBirth,
		&user.AddressLine1, &user.AddressLine2, &user.City, &user.Postcode, &user.Country,
		&user.KYCStatus, &user.KYCVerifiedAt, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, appErrors.NewNotFound("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, phone, password_hash, first_name, last_name,
			   date_of_birth, address_line1, address_line2, city, postcode, country,
			   kyc_status, kyc_verified_at, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.DateOfBirth,
		&user.AddressLine1, &user.AddressLine2, &user.City, &user.Postcode, &user.Country,
		&user.KYCStatus, &user.KYCVerifiedAt, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, appErrors.NewNotFound("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetByPhone retrieves a user by phone
func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	query := `
		SELECT id, email, phone, password_hash, first_name, last_name,
			   date_of_birth, address_line1, address_line2, city, postcode, country,
			   kyc_status, kyc_verified_at, is_active, created_at, updated_at
		FROM users
		WHERE phone = $1
	`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, phone).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.DateOfBirth,
		&user.AddressLine1, &user.AddressLine2, &user.City, &user.Postcode, &user.Country,
		&user.KYCStatus, &user.KYCVerifiedAt, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, appErrors.NewNotFound("user not found")
		}
		return nil, fmt.Errorf("failed to get user by phone: %w", err)
	}

	return user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET first_name = $2, last_name = $3, phone = $4,
			address_line1 = $5, address_line2 = $6, city = $7,
			postcode = $8, country = $9, updated_at = $10
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		user.ID, user.FirstName, user.LastName, user.Phone,
		user.AddressLine1, user.AddressLine2, user.City,
		user.Postcode, user.Country, user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return appErrors.NewNotFound("user not found")
	}

	return nil
}

// Delete deletes a user by ID
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return appErrors.NewNotFound("user not found")
	}

	return nil
}

// UpdateKYCStatus updates the KYC status for a user
func (r *userRepository) UpdateKYCStatus(ctx context.Context, id uuid.UUID, status string, verifiedAt *time.Time) error {
	query := `
		UPDATE users
		SET kyc_status = $2, kyc_verified_at = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, status, verifiedAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update KYC status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return appErrors.NewNotFound("user not found")
	}

	return nil
}

// SetInactive sets a user as inactive
func (r *userRepository) SetInactive(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET is_active = false, updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to set user inactive: %w", err)
	}

	if result.RowsAffected() == 0 {
		return appErrors.NewNotFound("user not found")
	}

	return nil
}

// isPgError checks if an error is a PostgreSQL error with a specific code
func isPgError(err error, code string) bool {
	if err == nil {
		return false
	}
	// Check if error message contains the code
	// This is a simplified check - in production use proper pgx error handling
	return contains(err.Error(), code)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || s[0:len(substr)] == substr || contains(s[1:], substr))
}
