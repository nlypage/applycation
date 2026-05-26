package entity

import "time"

// OwnerSession represents an authenticated owner session.
type OwnerSession struct {
	ID               string
	OwnerID          string
	SessionTokenHash string
	UserAgent        *string
	IPAddress        *string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	LastSeenAt       time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
