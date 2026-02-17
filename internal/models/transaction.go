package models

import "time"

type Transaction struct {
	ID        uint64
	Reference string
	Type      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LedgerEntry struct {
	ID                 uint64
	TransactionID      uint64
	Type               string
	Reference          string
	Status             string
	AccountID          uint64
	Amount             int64
	Direction          string // "debit" or "credit"
	BalanceAfter       int64
	LockedBalanceAfter int64
	CreatedAt          time.Time
}
