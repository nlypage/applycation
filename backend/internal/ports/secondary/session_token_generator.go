package secondary

import "context"

// SessionToken contains a plaintext session token and its storage-safe hash.
type SessionToken struct {
	Value string
	Hash  string
}

// SessionTokenGenerator creates opaque owner session tokens.
type SessionTokenGenerator interface {
	GenerateSessionToken(ctx context.Context) (SessionToken, error)
}
