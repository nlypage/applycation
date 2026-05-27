package crypto

import (
	"context"
	"fmt"

	secondaryports "github.com/nlypage/applycation/backend/internal/ports/secondary"
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher hashes local owner passwords with bcrypt.
type PasswordHasher struct {
	cost int
}

var _ secondaryports.PasswordHasher = (*PasswordHasher)(nil)

// NewPasswordHasher creates a bcrypt password hasher with the default cost.
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{cost: bcrypt.DefaultCost}
}

func newPasswordHasherWithCost(cost int) *PasswordHasher {
	return &PasswordHasher{cost: cost}
}

func (h *PasswordHasher) HashPassword(ctx context.Context, password string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return string(hash), nil
}

func (h *PasswordHasher) ComparePassword(ctx context.Context, passwordHash string, password string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("compare password: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return secondaryports.ErrInvalidPassword
		}
		return fmt.Errorf("compare password: %w", err)
	}

	return nil
}
