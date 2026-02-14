package models

import "time"

type Account struct {
	ID        uint64
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
