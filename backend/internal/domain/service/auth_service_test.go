package service

import (
	"context"
	"errors"
	"testing"
	"time"

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
	tokens := secondarymocks.NewMockSessionTokenGenerator(t)
	tx := secondarymocks.NewMockTxManager(t)
	svc := NewAuthService(owners, secondarymocks.NewMockOwnerSessionRepository(t), hasher, tokens, tx)

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
				secondarymocks.NewMockOwnerSessionRepository(t),
				secondarymocks.NewMockPasswordHasher(t),
				secondarymocks.NewMockSessionTokenGenerator(t),
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
	tokens := secondarymocks.NewMockSessionTokenGenerator(t)
	tx := secondarymocks.NewMockTxManager(t)
	svc := NewAuthService(owners, secondarymocks.NewMockOwnerSessionRepository(t), hasher, tokens, tx)

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
	tokens := secondarymocks.NewMockSessionTokenGenerator(t)
	tx := secondarymocks.NewMockTxManager(t)
	svc := NewAuthService(owners, secondarymocks.NewMockOwnerSessionRepository(t), hasher, tokens, tx)

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
	tokens := secondarymocks.NewMockSessionTokenGenerator(t)
	tx := secondarymocks.NewMockTxManager(t)
	svc := NewAuthService(owners, secondarymocks.NewMockOwnerSessionRepository(t), hasher, tokens, tx)

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
	tokens := secondarymocks.NewMockSessionTokenGenerator(t)
	tx := secondarymocks.NewMockTxManager(t)
	svc := NewAuthService(owners, secondarymocks.NewMockOwnerSessionRepository(t), hasher, tokens, tx)

	expectRunInTx(t, tx, ctx)
	owners.EXPECT().GetSingle(ctx).Return(entity.Owner{}, secondaryports.ErrNotFound).Once()
	hasher.EXPECT().HashPassword(ctx, validSetupSecret).Return("bcrypt-hash", nil).Once()
	owners.EXPECT().Create(ctx, "bcrypt-hash").Return(entity.Owner{}, createErr).Once()

	_, err := svc.Setup(ctx, setupInput(validSetupSecret))
	if !errors.Is(err, createErr) {
		t.Fatalf("Setup() error = %v, want %v", err, createErr)
	}
}

func TestAuthServiceLoginCreatesOwnerSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	owner := entity.Owner{ID: "owner-1", PasswordHash: "bcrypt-hash"}
	expectedSession := entity.OwnerSession{ID: "session-1", OwnerID: owner.ID, SessionTokenHash: "token-hash"}
	owners := secondarymocks.NewMockOwnerRepository(t)
	hasher := secondarymocks.NewMockPasswordHasher(t)
	tokens := secondarymocks.NewMockSessionTokenGenerator(t)
	tx := secondarymocks.NewMockTxManager(t)
	sessions := secondarymocks.NewMockOwnerSessionRepository(t)
	svc := NewAuthService(owners, sessions, hasher, tokens, tx)

	expectRunInTx(t, tx, ctx)
	owners.EXPECT().GetSingle(ctx).Return(owner, nil).Once()
	hasher.EXPECT().ComparePassword(ctx, owner.PasswordHash, validSetupSecret).Return(nil).Once()
	tokens.EXPECT().GenerateSessionToken(ctx).Return(secondaryports.SessionToken{Value: "raw-token", Hash: "token-hash"}, nil).Once()
	sessions.EXPECT().Create(ctx, mock.MatchedBy(func(params secondaryports.CreateOwnerSessionParams) bool {
		if params.OwnerID != owner.ID {
			t.Fatalf("OwnerID = %q, want %q", params.OwnerID, owner.ID)
		}
		if params.SessionTokenHash != "token-hash" {
			t.Fatalf("SessionTokenHash = %q, want token-hash", params.SessionTokenHash)
		}
		if params.UserAgent == nil || *params.UserAgent != "Mozilla/5.0" {
			t.Fatalf("UserAgent = %v, want Mozilla/5.0", params.UserAgent)
		}
		if params.IPAddress == nil || *params.IPAddress != "127.0.0.1" {
			t.Fatalf("IPAddress = %v, want 127.0.0.1", params.IPAddress)
		}
		if time.Until(params.ExpiresAt) <= 0 {
			t.Fatalf("ExpiresAt = %v, want future time", params.ExpiresAt)
		}
		return true
	})).Return(expectedSession, nil).Once()

	result, err := svc.Login(ctx, loginInput(validSetupSecret))
	if err != nil {
		t.Fatalf("Login() unexpected error: %v", err)
	}
	if result.Owner != owner {
		t.Fatalf("Login().Owner = %+v, want %+v", result.Owner, owner)
	}
	if result.Session != expectedSession {
		t.Fatalf("Login().Session = %+v, want %+v", result.Session, expectedSession)
	}
	if result.SessionToken != "raw-token" {
		t.Fatalf("Login().SessionToken = %q, want raw-token", result.SessionToken)
	}
}

func TestAuthServiceLoginRejectsInvalidPasswordInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		secret string
	}{
		{name: "empty", secret: ""},
		{name: "whitespace", secret: "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewAuthService(
				secondarymocks.NewMockOwnerRepository(t),
				secondarymocks.NewMockOwnerSessionRepository(t),
				secondarymocks.NewMockPasswordHasher(t),
				secondarymocks.NewMockSessionTokenGenerator(t),
				secondarymocks.NewMockTxManager(t),
			)

			_, err := svc.Login(context.Background(), loginInput(tt.secret))
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("Login() error = %v, want %v", err, ErrValidation)
			}
		})
	}
}

func TestAuthServiceLoginReturnsSetupRequired(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	owners := secondarymocks.NewMockOwnerRepository(t)
	hasher := secondarymocks.NewMockPasswordHasher(t)
	tokens := secondarymocks.NewMockSessionTokenGenerator(t)
	tx := secondarymocks.NewMockTxManager(t)
	svc := NewAuthService(owners, secondarymocks.NewMockOwnerSessionRepository(t), hasher, tokens, tx)

	expectRunInTx(t, tx, ctx)
	owners.EXPECT().GetSingle(ctx).Return(entity.Owner{}, secondaryports.ErrNotFound).Once()

	_, err := svc.Login(ctx, loginInput(validSetupSecret))
	if !errors.Is(err, ErrSetupRequired) {
		t.Fatalf("Login() error = %v, want %v", err, ErrSetupRequired)
	}
}

func TestAuthServiceLoginReturnsAuthFailedForInvalidPassword(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	owner := entity.Owner{ID: "owner-1", PasswordHash: "bcrypt-hash"}
	owners := secondarymocks.NewMockOwnerRepository(t)
	hasher := secondarymocks.NewMockPasswordHasher(t)
	tokens := secondarymocks.NewMockSessionTokenGenerator(t)
	tx := secondarymocks.NewMockTxManager(t)
	svc := NewAuthService(owners, secondarymocks.NewMockOwnerSessionRepository(t), hasher, tokens, tx)

	expectRunInTx(t, tx, ctx)
	owners.EXPECT().GetSingle(ctx).Return(owner, nil).Once()
	hasher.EXPECT().ComparePassword(ctx, owner.PasswordHash, validSetupSecret).Return(secondaryports.ErrInvalidPassword).Once()

	_, err := svc.Login(ctx, loginInput(validSetupSecret))
	if !errors.Is(err, ErrAuthFailed) {
		t.Fatalf("Login() error = %v, want %v", err, ErrAuthFailed)
	}
}

func TestAuthServiceLoginPropagatesTokenGenerationError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tokenErr := errors.New("entropy unavailable")
	owner := entity.Owner{ID: "owner-1", PasswordHash: "bcrypt-hash"}
	owners := secondarymocks.NewMockOwnerRepository(t)
	hasher := secondarymocks.NewMockPasswordHasher(t)
	tokens := secondarymocks.NewMockSessionTokenGenerator(t)
	tx := secondarymocks.NewMockTxManager(t)
	svc := NewAuthService(owners, secondarymocks.NewMockOwnerSessionRepository(t), hasher, tokens, tx)

	expectRunInTx(t, tx, ctx)
	owners.EXPECT().GetSingle(ctx).Return(owner, nil).Once()
	hasher.EXPECT().ComparePassword(ctx, owner.PasswordHash, validSetupSecret).Return(nil).Once()
	tokens.EXPECT().GenerateSessionToken(ctx).Return(secondaryports.SessionToken{}, tokenErr).Once()

	_, err := svc.Login(ctx, loginInput(validSetupSecret))
	if !errors.Is(err, tokenErr) {
		t.Fatalf("Login() error = %v, want %v", err, tokenErr)
	}
}

func TestAuthServiceLoginPropagatesCreateSessionError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	createErr := errors.New("create session failed")
	owner := entity.Owner{ID: "owner-1", PasswordHash: "bcrypt-hash"}
	owners := secondarymocks.NewMockOwnerRepository(t)
	hasher := secondarymocks.NewMockPasswordHasher(t)
	tokens := secondarymocks.NewMockSessionTokenGenerator(t)
	tx := secondarymocks.NewMockTxManager(t)
	sessions := secondarymocks.NewMockOwnerSessionRepository(t)
	svc := NewAuthService(owners, sessions, hasher, tokens, tx)

	expectRunInTx(t, tx, ctx)
	owners.EXPECT().GetSingle(ctx).Return(owner, nil).Once()
	hasher.EXPECT().ComparePassword(ctx, owner.PasswordHash, validSetupSecret).Return(nil).Once()
	tokens.EXPECT().GenerateSessionToken(ctx).Return(secondaryports.SessionToken{Value: "raw-token", Hash: "token-hash"}, nil).Once()
	sessions.EXPECT().Create(ctx, mock.Anything).Return(entity.OwnerSession{}, createErr).Once()

	_, err := svc.Login(ctx, loginInput(validSetupSecret))
	if !errors.Is(err, createErr) {
		t.Fatalf("Login() error = %v, want %v", err, createErr)
	}
}

func TestAuthServiceLogoutRevokesOwnerSession(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	hasher := secondarymocks.NewMockPasswordHasher(t)
	tokens := secondarymocks.NewMockSessionTokenGenerator(t)
	tx := secondarymocks.NewMockTxManager(t)
	sessions := secondarymocks.NewMockOwnerSessionRepository(t)
	svc := NewAuthService(secondarymocks.NewMockOwnerRepository(t), sessions, hasher, tokens, tx)

	expectRunInTx(t, tx, ctx)
	sessions.EXPECT().Revoke(ctx, hashSessionToken("raw-token")).Return(nil).Once()

	err := svc.Logout(ctx, logoutInput("raw-token"))
	if err != nil {
		t.Fatalf("Logout() unexpected error: %v", err)
	}
}

func TestAuthServiceLogoutRejectsMissingToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		token string
	}{
		{name: "empty", token: ""},
		{name: "whitespace", token: "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewAuthService(
				secondarymocks.NewMockOwnerRepository(t),
				secondarymocks.NewMockOwnerSessionRepository(t),
				secondarymocks.NewMockPasswordHasher(t),
				secondarymocks.NewMockSessionTokenGenerator(t),
				secondarymocks.NewMockTxManager(t),
			)

			err := svc.Logout(context.Background(), logoutInput(tt.token))
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("Logout() error = %v, want %v", err, ErrValidation)
			}
		})
	}
}

func TestAuthServiceLogoutReturnsAuthFailedWhenSessionNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	hasher := secondarymocks.NewMockPasswordHasher(t)
	tokens := secondarymocks.NewMockSessionTokenGenerator(t)
	tx := secondarymocks.NewMockTxManager(t)
	sessions := secondarymocks.NewMockOwnerSessionRepository(t)
	svc := NewAuthService(secondarymocks.NewMockOwnerRepository(t), sessions, hasher, tokens, tx)

	expectRunInTx(t, tx, ctx)
	sessions.EXPECT().Revoke(ctx, hashSessionToken("missing-token")).Return(secondaryports.ErrNotFound).Once()

	err := svc.Logout(ctx, logoutInput("missing-token"))
	if !errors.Is(err, ErrAuthFailed) {
		t.Fatalf("Logout() error = %v, want %v", err, ErrAuthFailed)
	}
}

func TestAuthServiceLogoutPropagatesRevokeError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	revokeErr := errors.New("revoke failed")
	hasher := secondarymocks.NewMockPasswordHasher(t)
	tokens := secondarymocks.NewMockSessionTokenGenerator(t)
	tx := secondarymocks.NewMockTxManager(t)
	sessions := secondarymocks.NewMockOwnerSessionRepository(t)
	svc := NewAuthService(secondarymocks.NewMockOwnerRepository(t), sessions, hasher, tokens, tx)

	expectRunInTx(t, tx, ctx)
	sessions.EXPECT().Revoke(ctx, hashSessionToken("raw-token")).Return(revokeErr).Once()

	err := svc.Logout(ctx, logoutInput("raw-token"))
	if !errors.Is(err, revokeErr) {
		t.Fatalf("Logout() error = %v, want %v", err, revokeErr)
	}
}

func setupInput(secret string) primaryports.SetupInput {
	return primaryports.SetupInput{Password: secret}
}

func loginInput(secret string) primaryports.LoginInput {
	return primaryports.LoginInput{
		Password:  secret,
		UserAgent: "Mozilla/5.0",
		IPAddress: "127.0.0.1",
	}
}

func logoutInput(sessionToken string) primaryports.LogoutInput {
	return primaryports.LogoutInput{SessionToken: sessionToken}
}

func expectRunInTx(t *testing.T, tx *secondarymocks.MockTxManager, ctx context.Context) {
	t.Helper()

	tx.EXPECT().RunInTx(ctx, mock.AnythingOfType("func(context.Context) error")).RunAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		},
	).Once()
}

