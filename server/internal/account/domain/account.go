package domain

import "time"

type Account struct {
	ID        string
	OwnerName string
	Balance   int64
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Status string

const (
	StatusActive  Status = "ACTIVE"
	StatusBlocked Status = "BLOCKED"
	StatusClosed  Status = "CLOSED"
)
