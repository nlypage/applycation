package secondary

import "context"

// PasswordHasher hashes and verifies local owner passwords.
type PasswordHasher interface {
	HashPassword(ctx context.Context, password string) (string, error)
	ComparePassword(ctx context.Context, passwordHash string, password string) error
}
