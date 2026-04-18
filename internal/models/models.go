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
	ID           int
	UserID       int
	ZoneID       int
	Action       string
	Status       string
	Timestamp    time.Time
	Hash         string
	PreviousHash string
}

type ZoneOccupancy struct {
	Zone
	ActiveCount int
	ActiveUsers []User
}
