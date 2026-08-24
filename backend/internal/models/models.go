package models

import "time"

type User struct {
	ID             int        `json:"id"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	PasswordHash   string     `json:"-"`
	Role           string     `json:"role"`
	FailedAttempts int        `json:"failed_attempts"`
	LockedUntil    *time.Time `json:"locked_until,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type Zone struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	MaxCapacity      int       `json:"max_capacity"`
	RequiresExitScan bool      `json:"requires_exit_scan"`
	CreatedAt        time.Time `json:"created_at"`
}

type AccessEvent struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	ZoneID       int       `json:"zone_id"`
	Action       string    `json:"action"`
	Status       string    `json:"status"`
	Reason       *string   `json:"reason,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Hash         string    `json:"hash"`
	PreviousHash string    `json:"previous_hash"`
	DeviceID     *int      `json:"device_id,omitempty"`
	EntryMethod  string    `json:"entry_method,omitempty"`
}

type ZoneOccupancy struct {
	Zone
	ActiveCount int     `json:"active_count"`
	ActiveUsers []*User `json:"active_users,omitempty"`
}

type Device struct {
	ID           int       `json:"id"`
	ZoneID       int       `json:"zone_id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Serial       string    `json:"serial"`
	Active       bool      `json:"active"`
	IsEntryPoint bool      `json:"is_entry_point"`
	CreatedAt    time.Time `json:"created_at"`
}

type BiometricCredential struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	EntryMethod    string    `json:"entry_method"`
	CredentialHash string    `json:"credential_hash"`
	EnrolledAt     time.Time `json:"enrolled_at"`
	Revoked        bool      `json:"revoked"`
}

type ZoneAnalytics struct {
	DayOfWeek  int `json:"day_of_week"` // 0=Sunday, 6=Saturday
	Hour       int `json:"hour"`        // 0-23
	EntryCount int `json:"entry_count"`
}

// UserZoneBreakdown summarises one zone's share of a user's audited activity.
type UserZoneBreakdown struct {
	ZoneID   int       `json:"zone_id"`
	ZoneName string    `json:"zone_name"`
	Entries  int       `json:"entries"`
	Denied   int       `json:"denied"`
	LastSeen time.Time `json:"last_seen"`
}

// UserAnalytics aggregates every access event recorded against one user.
type UserAnalytics struct {
	TotalEvents  int                  `json:"total_events"`
	Entries      int                  `json:"entries"`
	Exits        int                  `json:"exits"`
	Denied       int                  `json:"denied"`
	ZonesVisited int                  `json:"zones_visited"`
	LastEventAt  *time.Time           `json:"last_event_at,omitempty"`
	Zones        []*UserZoneBreakdown `json:"zones"`
}

type ZoneOccupancySnapshot struct {
	Zone
	ActiveCount      int     `json:"active_count"`
	OccupancyPercent float64 `json:"occupancy_percent"`
}
