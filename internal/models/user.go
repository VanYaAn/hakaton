package models

import "time"

type User struct {
	ID        int64
	MaxUserID string
	Username  string
	CreatedAt time.Time
}
