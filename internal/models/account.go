package models

import "time"

type Account struct {
	ID        uint64
	Name      string
	Type      string
	Code      string
	Currency  string
	UserID    uint64
	Balance   int64 // Cached balance
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AccountBalance struct {
	AccountID     uint64
	Balance       int64
	LockedBalance int64
}
