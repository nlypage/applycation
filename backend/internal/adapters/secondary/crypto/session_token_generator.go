package crypto

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	secondaryports "github.com/nlypage/applycation/backend/internal/ports/secondary"
)

const sessionTokenByteLength = 32

// SessionTokenGenerator creates cryptographically secure opaque session tokens.
type SessionTokenGenerator struct{}

var _ secondaryports.SessionTokenGenerator = (*SessionTokenGenerator)(nil)

// NewSessionTokenGenerator creates a secure owner session token generator.
func NewSessionTokenGenerator() *SessionTokenGenerator {
	return &SessionTokenGenerator{}
}

func (g *SessionTokenGenerator) GenerateSessionToken(ctx context.Context) (secondaryports.SessionToken, error) {
	if err := ctx.Err(); err != nil {
		return secondaryports.SessionToken{}, fmt.Errorf("generate session token: %w", err)
	}

	raw := make([]byte, sessionTokenByteLength)
	if _, err := rand.Read(raw); err != nil {
		return secondaryports.SessionToken{}, fmt.Errorf("generate session token: %w", err)
	}

	value := base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256.Sum256([]byte(value))

	return secondaryports.SessionToken{
		Value: value,
		Hash:  hex.EncodeToString(hash[:]),
	}, nil
}
