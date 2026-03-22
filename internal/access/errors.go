package access

import "errors"

var (
	ErrZoneFull          = errors.New("zone is full")
	ErrUserAlreadyInZone = errors.New("user already in zone")
	ErrNoActiveSession   = errors.New("no active session found")
	ErrZoneNotFound      = errors.New("zone not found")
	ErrAccessDenied      = errors.New("access denied")
)
