package entity

import "time"

// Owner represents the single local application owner.
type Owner struct {
	ID                string
	PasswordHash      string
	PasswordChangedAt time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
