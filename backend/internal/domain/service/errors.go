package service

import "errors"

var (
	ErrValidation            = errors.New("validation error")
	ErrSetupAlreadyCompleted = errors.New("setup already completed")
	ErrSetupRequired         = errors.New("setup required")
	ErrAuthFailed            = errors.New("auth failed")
)
