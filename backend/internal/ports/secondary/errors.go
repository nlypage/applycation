package secondary

import "errors"

var (
	ErrNotImplemented  = errors.New("secondary adapter is not implemented")
	ErrNotFound        = errors.New("resource not found")
	ErrInvalidPassword = errors.New("invalid password")
)
