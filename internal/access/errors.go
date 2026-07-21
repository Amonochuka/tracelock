package access

import "errors"

var (
	ErrZoneFull            = errors.New("zone is full")
	ErrUserAlreadyInZone   = errors.New("user already in zone")
	ErrNoActiveSession     = errors.New("no active session found")
	ErrZoneNotFound        = errors.New("zone not found")
	ErrNoHashFound         = errors.New("hash does not exist")
	ErrAccessNotFound      = errors.New("access grant not found")
	ErrAccessDenied        = errors.New("user does not have access to this zone")
	ErrZoneNameExists      = errors.New("zone name already exists")
	ErrZoneHasActivity     = errors.New("zone has active sessions and cannot be deleted")
	ErrDeviceNotFound      = errors.New("device not found")
	ErrDeviceSerialExists  = errors.New("device serial already exists")
	ErrCredentialExists    = errors.New("credential already exists for this method")
	ErrCredentialNotFound  = errors.New("credential not found")
	ErrDeviceInactive      = errors.New("device not active")
	ErrCredentialRevoked   = errors.New("credentials have been revoked")
	ErrAccountLocked       = errors.New("account is temporarily locked")
	ErrRequiresExitScan    = errors.New("must explicitly exit current zone first")
)
