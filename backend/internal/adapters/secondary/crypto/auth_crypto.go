package crypto

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	secondaryports "github.com/nlypage/applycation/backend/internal/ports/secondary"
)

// AuthCrypto объединяет крипто-операции локальной авторизации.
type AuthCrypto struct {
	passwordHasher *PasswordHasher
	tokenGenerator *SessionTokenGenerator
}

var _ secondaryports.CryptoPort = (*AuthCrypto)(nil)

func NewAuthCrypto() *AuthCrypto {
	return &AuthCrypto{
		passwordHasher: NewPasswordHasher(),
		tokenGenerator: NewSessionTokenGenerator(),
	}
}

func (c *AuthCrypto) HashPassword(ctx context.Context, password string) (string, error) {
	return c.passwordHasher.HashPassword(ctx, password)
}

func (c *AuthCrypto) ComparePassword(ctx context.Context, passwordHash string, password string) error {
	return c.passwordHasher.ComparePassword(ctx, passwordHash, password)
}

func (c *AuthCrypto) GenerateSessionToken(ctx context.Context) (secondaryports.SessionToken, error) {
	return c.tokenGenerator.GenerateSessionToken(ctx)
}

func (c *AuthCrypto) HashSessionToken(token string) string {
	return hashSessionToken(token)
}

func hashSessionToken(token string) string {
	digest := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(digest[:])
}
