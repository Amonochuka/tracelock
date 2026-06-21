package access

import "tracelock/internal/models"

// ZoneRepository defines everything ZoneService needs from the data layer.
// This lets tests use a mock repo instead of hitting a real database.
type ZoneRepository interface {
	CreateZone(name, description string, maxCapacity int) (*models.Zone, error)
	DeleteZone(zoneID int) error
	GetZone(zoneID int) (*models.Zone, error)
	GetMaximumCapacity(zoneID int) (int, error)
	CreateEvent(userID, zoneID int, action, status, hash, previousHash string, deviceID *int, entryMethod string) error
	GetLastHash(zoneID int) (string, error)
	CreateSession(userID, zoneID int) error
	DeleteSession(userID, zoneID int) error
	CountActiveUsers(zoneID int) (int, error)
	HasZoneAccess(userID, zoneID int, role string) (bool, error)
	GrantZoneAccess(userID, zoneID, grantedBy int) error
	RevokeZoneAccess(userID, zoneID int) error
	ListUserZoneAccess(userID int) ([]*models.Zone, error)
	ListZoneUsers(zoneID int) ([]*models.User, error)
	ListZones() ([]*models.Zone, error)
	UpdateZone(zoneID int, name, description string, maxCapacity int) (*models.Zone, error)
	GetActiveUsersInZone(zoneID int) ([]*models.User, error)
	ListZoneEvents(zoneID, limit, offset int) ([]*models.AccessEvent, int, error)
	ListUserEvents(userID, limit, offset int) ([]*models.AccessEvent, int, error)
	VerifyChain(zoneID int) (bool, int, error)
	GetActiveSessionForUser(userID int) (int, error)
	ListZoneOccupancy() ([]*models.ZoneOccupancySnapshot, error)
	GetZoneAnalytics(zoneID int) ([]*models.ZoneAnalytics, error)
}

type DeviceRepository interface {
	CreateDevice(zoneID int, name, deviceType, serial string) (*models.Device, error)
	GetDevice(deviceID int) (*models.Device, error)
	ListZoneDevices(zoneID int) ([]*models.Device, error)
	UpdateDevice(deviceID int, name, deviceType, serial string) (*models.Device, error)
	DeactivateDevice(deviceID int) error
	DeleteDevice(deviceID int) error
}

type CredentialRepository interface {
	EnrollCredential(userID int, entryMethod, credentialHash string) (*models.BiometricCredential, error)
	GetCredential(userID int, entryMethod string) (*models.BiometricCredential, error)
	RevokeCredential(userID int, entryMethod string) error
	ListUserCredentials(userID int) ([]*models.BiometricCredential, error)
	GetCredentialByHash(hash string) (*models.BiometricCredential, error)
}
