package service

import (
	"context"
	"errors"
	"testing"

	"github.com/nlypage/applycation/backend/internal/domain/entity"
	primaryports "github.com/nlypage/applycation/backend/internal/ports/primary"
	secondaryports "github.com/nlypage/applycation/backend/internal/ports/secondary"
	secondarymocks "github.com/nlypage/applycation/backend/internal/testutil/mocks/secondary"
	"github.com/stretchr/testify/mock"
)

const validSetupSecret = "correct horse battery staple"

func TestAuthServiceSetupCreatesFirstOwner(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	expectedOwner := entity.Owner{ID: "owner-1", PasswordHash: "bcrypt-hash"}
	owners := secondarymocks.NewMockOwnerRepository(t)
	hasher := secondarymocks.NewMockPasswordHasher(t)
	tx := secondarymocks.NewMockTxManager(t)
	svc := NewAuthService(owners, hasher, tx)

	expectRunInTx(t, tx, ctx)
	owners.EXPECT().GetSingle(ctx).Return(entity.Owner{}, secondaryports.ErrNotFound).Once()
	hasher.EXPECT().HashPassword(ctx, validSetupSecret).Return("bcrypt-hash", nil).Once()
	owners.EXPECT().Create(ctx, "bcrypt-hash").Return(expectedOwner, nil).Once()

	owner, err := svc.Setup(ctx, setupInput(validSetupSecret))
	if err != nil {
		t.Fatalf("Setup() unexpected error: %v", err)
	}
	if owner != expectedOwner {
		t.Fatalf("Setup() = %+v, want %+v", owner, expectedOwner)
	}
}

func TestAuthServiceSetupRejectsInvalidPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		secret string
	}{
		{name: "empty", secret: ""},
		{name: "whitespace", secret: "   "},
		{name: "too short", secret: "short"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewAuthService(
				secondarymocks.NewMockOwnerRepository(t),
				secondarymocks.NewMockPasswordHasher(t),
				secondarymocks.NewMockTxManager(t),
			)

			_, err := svc.Setup(context.Background(), setupInput(tt.secret))
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("Setup() error = %v, want %v", err, ErrValidation)
			}
		})
	}
}

func TestAuthServiceSetupReturnsAlreadyCompleted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	owners := secondarymocks.NewMockOwnerRepository(t)
	hasher := secondarymocks.NewMockPasswordHasher(t)
	tx := secondarymocks.NewMockTxManager(t)
	svc := NewAuthService(owners, hasher, tx)

	expectRunInTx(t, tx, ctx)
	owners.EXPECT().GetSingle(ctx).Return(entity.Owner{ID: "existing-owner"}, nil).Once()

	_, err := svc.Setup(ctx, setupInput(validSetupSecret))
	if !errors.Is(err, ErrSetupAlreadyCompleted) {
		t.Fatalf("Setup() error = %v, want %v", err, ErrSetupAlreadyCompleted)
	}
}

func TestAuthServiceSetupPropagatesRepositoryCheckError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	checkErr := errors.New("database unavailable")
	owners := secondarymocks.NewMockOwnerRepository(t)
	hasher := secondarymocks.NewMockPasswordHasher(t)
	tx := secondarymocks.NewMockTxManager(t)
	svc := NewAuthService(owners, hasher, tx)

	expectRunInTx(t, tx, ctx)
	owners.EXPECT().GetSingle(ctx).Return(entity.Owner{}, checkErr).Once()

	_, err := svc.Setup(ctx, setupInput(validSetupSecret))
	if !errors.Is(err, checkErr) {
		t.Fatalf("Setup() error = %v, want %v", err, checkErr)
	}
}

func TestAuthServiceSetupPropagatesHashError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	hashErr := errors.New("hash failed")
	owners := secondarymocks.NewMockOwnerRepository(t)
	hasher := secondarymocks.NewMockPasswordHasher(t)
	tx := secondarymocks.NewMockTxManager(t)
	svc := NewAuthService(owners, hasher, tx)

	expectRunInTx(t, tx, ctx)
	owners.EXPECT().GetSingle(ctx).Return(entity.Owner{}, secondaryports.ErrNotFound).Once()
	hasher.EXPECT().HashPassword(ctx, validSetupSecret).Return("", hashErr).Once()

	_, err := svc.Setup(ctx, setupInput(validSetupSecret))
	if !errors.Is(err, hashErr) {
		t.Fatalf("Setup() error = %v, want %v", err, hashErr)
	}
}

func TestAuthServiceSetupPropagatesCreateError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	createErr := errors.New("create failed")
	owners := secondarymocks.NewMockOwnerRepository(t)
	hasher := secondarymocks.NewMockPasswordHasher(t)
	tx := secondarymocks.NewMockTxManager(t)
	svc := NewAuthService(owners, hasher, tx)

	expectRunInTx(t, tx, ctx)
	owners.EXPECT().GetSingle(ctx).Return(entity.Owner{}, secondaryports.ErrNotFound).Once()
	hasher.EXPECT().HashPassword(ctx, validSetupSecret).Return("bcrypt-hash", nil).Once()
	owners.EXPECT().Create(ctx, "bcrypt-hash").Return(entity.Owner{}, createErr).Once()

	_, err := svc.Setup(ctx, setupInput(validSetupSecret))
	if !errors.Is(err, createErr) {
		t.Fatalf("Setup() error = %v, want %v", err, createErr)
	}
}

func setupInput(secret string) primaryports.SetupInput {
	return primaryports.SetupInput{Password: secret}
}

func expectRunInTx(t *testing.T, tx *secondarymocks.MockTxManager, ctx context.Context) {
	t.Helper()

	tx.EXPECT().RunInTx(ctx, mock.AnythingOfType("func(context.Context) error")).RunAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	).Once()
}
