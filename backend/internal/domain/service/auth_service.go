package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nlypage/applycation/backend/internal/domain/entity"
	primaryports "github.com/nlypage/applycation/backend/internal/ports/primary"
	secondaryports "github.com/nlypage/applycation/backend/internal/ports/secondary"
)

const (
	minOwnerPasswordLength = 8
	defaultOwnerSessionTTL = 30 * 24 * time.Hour
)

// AuthService orchestrates local owner authentication use cases.
type AuthService struct {
	owners                secondaryports.OwnerRepository
	ownerSessions         secondaryports.OwnerSessionRepository
	passwordHasher        secondaryports.PasswordHasher
	sessionTokenGenerator secondaryports.SessionTokenGenerator
	txManager             secondaryports.TxManager
}

var _ primaryports.AuthService = (*AuthService)(nil)

// NewAuthService creates a local authentication service.
func NewAuthService(
	owners secondaryports.OwnerRepository,
	ownerSessions secondaryports.OwnerSessionRepository,
	passwordHasher secondaryports.PasswordHasher,
	sessionTokenGenerator secondaryports.SessionTokenGenerator,
	txManager secondaryports.TxManager,
) *AuthService {
	return &AuthService{
		owners:                owners,
		ownerSessions:         ownerSessions,
		passwordHasher:        passwordHasher,
		sessionTokenGenerator: sessionTokenGenerator,
		txManager:             txManager,
	}
}

func (s *AuthService) Setup(ctx context.Context, input primaryports.SetupInput) (entity.Owner, error) {
	if err := validateSetupInput(input); err != nil {
		return entity.Owner{}, err
	}

	var created entity.Owner
	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		_, err := s.owners.GetSingle(txCtx)
		if err == nil {
			return ErrSetupAlreadyCompleted
		}
		if !errors.Is(err, secondaryports.ErrNotFound) {
			return fmt.Errorf("check existing owner: %w", err)
		}

		passwordHash, err := s.passwordHasher.HashPassword(txCtx, input.Password)
		if err != nil {
			return fmt.Errorf("hash setup password: %w", err)
		}

		owner, err := s.owners.Create(txCtx, passwordHash)
		if err != nil {
			return fmt.Errorf("create owner: %w", err)
		}

		created = owner
		return nil
	})
	if err != nil {
		return entity.Owner{}, err
	}

	return created, nil
}

func (s *AuthService) Login(ctx context.Context, input primaryports.LoginInput) (primaryports.LoginResult, error) {
	if err := validateLoginInput(input); err != nil {
		return primaryports.LoginResult{}, err
	}

	var result primaryports.LoginResult
	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		owner, err := s.owners.GetSingle(txCtx)
		if errors.Is(err, secondaryports.ErrNotFound) {
			return ErrSetupRequired
		}
		if err != nil {
			return fmt.Errorf("get owner for login: %w", err)
		}

		if err := s.passwordHasher.ComparePassword(txCtx, owner.PasswordHash, input.Password); err != nil {
			if errors.Is(err, secondaryports.ErrInvalidPassword) {
				return ErrAuthFailed
			}
			return fmt.Errorf("compare owner password: %w", err)
		}

		token, err := s.sessionTokenGenerator.GenerateSessionToken(txCtx)
		if err != nil {
			return fmt.Errorf("generate owner session token: %w", err)
		}

		session, err := s.ownerSessions.Create(txCtx, secondaryports.CreateOwnerSessionParams{
			OwnerID:          owner.ID,
			SessionTokenHash: token.Hash,
			UserAgent:        optionalString(input.UserAgent),
			IPAddress:        optionalString(input.IPAddress),
			ExpiresAt:        time.Now().UTC().Add(defaultOwnerSessionTTL),
		})
		if err != nil {
			return fmt.Errorf("create owner session: %w", err)
		}

		result = primaryports.LoginResult{
			Owner:        owner,
			Session:      session,
			SessionToken: token.Value,
		}
		return nil
	})
	if err != nil {
		return primaryports.LoginResult{}, err
	}

	return result, nil
}

func (s *AuthService) Logout(ctx context.Context, input primaryports.LogoutInput) error {
	if err := validateLogoutInput(input); err != nil {
		return err
	}

	tokenHash := hashSessionToken(input.SessionToken)
	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		err := s.ownerSessions.Revoke(txCtx, tokenHash)
		if errors.Is(err, secondaryports.ErrNotFound) {
			return ErrAuthFailed
		}
		if err != nil {
			return fmt.Errorf("revoke owner session: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func validateSetupInput(input primaryports.SetupInput) error {
	if strings.TrimSpace(input.Password) == "" {
		return fmt.Errorf("%w: password is required", ErrValidation)
	}
	if len(input.Password) < minOwnerPasswordLength {
		return fmt.Errorf("%w: password must contain at least %d characters", ErrValidation, minOwnerPasswordLength)
	}
	return nil
}

func validateLoginInput(input primaryports.LoginInput) error {
	if strings.TrimSpace(input.Password) == "" {
		return fmt.Errorf("%w: password is required", ErrValidation)
	}
	return nil
}

func validateLogoutInput(input primaryports.LogoutInput) error {
	if strings.TrimSpace(input.SessionToken) == "" {
		return fmt.Errorf("%w: session token is required", ErrValidation)
	}
	return nil
}

func hashSessionToken(token string) string {
	digest := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(digest[:])
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
