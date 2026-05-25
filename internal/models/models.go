package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Zone struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	MaxCapacity int       `json:"max_capacity"`
	CreatedAt   time.Time `json:"created_at"`
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
	DeviceID     *int      `json:"device_id,omitempty"`
	EntryMethod  string    `json:"entry_method,omitempty"`
}

type ZoneOccupancy struct {
	Zone
	ActiveCount int     `json:"active_count"`
	ActiveUsers []*User `json:"active_users"`
}

type Device struct {
	ID        int       `json:"id"`
	ZoneID    int       `json:"zone_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Serial    string    `json:"serial"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type BiometricCredential struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	EntryMethod    string    `json:"entry_method"`
	CredentialHash string    `json:"credential_hash"`
	EnrolledAt     time.Time `json:"enrolled_at"`
	Revoked        bool      `json:"revoked"`
}