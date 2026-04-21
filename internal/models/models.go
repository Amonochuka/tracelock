package models

import "time"

type User struct {
	ID           int
	Name         string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

type Zone struct {
	ID          int
	Name        string
	Description string
	MaxCapacity int
	CreatedAt   time.Time
}

type AccessEvent struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	ZoneID       int       `json:"zone_id"`
	Action       string    `json:"action"`
	Status       string    `json:"status"`
	Timestamp    time.Time `json:"timestamp"`
	Hash         string    `json:"hash"`
	PreviousHash string    `json:"previous_hash"`
}

type ZoneOccupancy struct {
	Zone
	ActiveCount int
	ActiveUsers []*User `json:"active_users"`
}
