package secondary

import "context"

// CryptoPort provides auth-related cryptographic operations.
type CryptoPort interface {
	HashPassword(ctx context.Context, password string) (string, error)
	ComparePassword(ctx context.Context, passwordHash string, password string) error
	GenerateSessionToken(ctx context.Context) (SessionToken, error)
	HashSessionToken(token string) string
}
