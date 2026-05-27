package crypto

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

func TestSessionTokenGeneratorGeneratesUniqueHashedTokens(t *testing.T) {
	t.Parallel()

	generator := NewSessionTokenGenerator()
	first, err := generator.GenerateSessionToken(context.Background())
	if err != nil {
		t.Fatalf("GenerateSessionToken() first unexpected error: %v", err)
	}
	second, err := generator.GenerateSessionToken(context.Background())
	if err != nil {
		t.Fatalf("GenerateSessionToken() second unexpected error: %v", err)
	}

	if first.Value == "" {
		t.Fatal("GenerateSessionToken().Value is empty")
	}
	if first.Hash == "" {
		t.Fatal("GenerateSessionToken().Hash is empty")
	}
	if first.Value == second.Value {
		t.Fatal("GenerateSessionToken() generated duplicate token values")
	}
	if first.Hash == second.Hash {
		t.Fatal("GenerateSessionToken() generated duplicate token hashes")
	}

	expectedHash := sha256.Sum256([]byte(first.Value))
	if first.Hash != hex.EncodeToString(expectedHash[:]) {
		t.Fatalf("GenerateSessionToken().Hash = %q, want sha256 hash of token value", first.Hash)
	}
}

func TestSessionTokenGeneratorHonorsCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewSessionTokenGenerator().GenerateSessionToken(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("GenerateSessionToken() error = %v, want %v", err, context.Canceled)
	}
}
