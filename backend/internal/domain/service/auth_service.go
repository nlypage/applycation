package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/nlypage/applycation/backend/internal/domain/entity"
	primaryports "github.com/nlypage/applycation/backend/internal/ports/primary"
	secondaryports "github.com/nlypage/applycation/backend/internal/ports/secondary"
)

const minOwnerPasswordLength = 8

// AuthService orchestrates local owner authentication use cases.
type AuthService struct {
	owners         secondaryports.OwnerRepository
	passwordHasher secondaryports.PasswordHasher
	txManager      secondaryports.TxManager
}

var _ primaryports.AuthService = (*AuthService)(nil)

// NewAuthService creates a local authentication service.
func NewAuthService(
	owners secondaryports.OwnerRepository,
	passwordHasher secondaryports.PasswordHasher,
	txManager secondaryports.TxManager,
) *AuthService {
	return &AuthService{
		owners:         owners,
		passwordHasher: passwordHasher,
		txManager:      txManager,
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

func validateSetupInput(input primaryports.SetupInput) error {
	if strings.TrimSpace(input.Password) == "" {
		return fmt.Errorf("%w: password is required", ErrValidation)
	}
	if len(input.Password) < minOwnerPasswordLength {
		return fmt.Errorf("%w: password must contain at least %d characters", ErrValidation, minOwnerPasswordLength)
	}
	return nil
}
