package access

import "errors"

var (
	ErrZoneFull          = errors.New("zone is full")
	ErrUserAlreadyInZone = errors.New("user already in zone")
	ErrNoActiveSession   = errors.New("no active session found")
	ErrZoneNotFound      = errors.New("zone not found")
	ErrNoHashFound       = errors.New("hash does not exist")
	ErrAccessNotFound    = errors.New("access grant ot found")
	ErrAccessDenied      = errors.New("user does not have access to this zone")
	ErrZoneNameExists    = errors.New("zone name already exists")
	ErrZoneHasActivity   = errors.New("zone has active sessions and cannot be deleted")
)
