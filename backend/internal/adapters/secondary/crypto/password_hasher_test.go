package crypto

import (
	"context"
	"errors"
	"testing"

	secondaryports "github.com/nlypage/applycation/backend/internal/ports/secondary"
)

func TestPasswordHasher(t *testing.T) {
	t.Parallel()

	hasher := newPasswordHasherWithCost(4)
	ctx := context.Background()

	hash, err := hasher.HashPassword(ctx, "correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword() returned empty hash")
	}
	if hash == "correct horse battery staple" {
		t.Fatal("HashPassword() returned plaintext password")
	}

	if err := hasher.ComparePassword(ctx, hash, "correct horse battery staple"); err != nil {
		t.Fatalf("ComparePassword() unexpected error: %v", err)
	}

	err = hasher.ComparePassword(ctx, hash, "wrong password")
	if !errors.Is(err, secondaryports.ErrInvalidPassword) {
		t.Fatalf("ComparePassword() error = %v, want %v", err, secondaryports.ErrInvalidPassword)
	}
}

func TestPasswordHasherHonorsCancelledContext(t *testing.T) {
	t.Parallel()

	hasher := newPasswordHasherWithCost(4)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := hasher.HashPassword(ctx, "correct horse battery staple")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("HashPassword() error = %v, want %v", err, context.Canceled)
	}

	err = hasher.ComparePassword(ctx, "$2a$04$abcdefghijklmnopqrstuuVdDnG6D3SPZz0eDyGrBGrn7VX1m0sTi", "password")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ComparePassword() error = %v, want %v", err, context.Canceled)
	}
}
