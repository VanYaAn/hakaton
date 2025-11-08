package domain

import "time"

type User struct {
	ID        int64
	MaxUserID string // MAX messenger user ID
	Username  string
	CreatedAt time.Time
}
