package access

import (
	"time"
	"tracelock/internal/models"
)

// UserResolver abstracts user lookup so BiometricService
// does not depend directly on the auth package
type UserResolver interface {
	VerifyUser(id int) (*models.User, error)
}

type JWTIssuer interface {
	GenerateToken(user *models.User) (string, error)
}

type BiometricService struct {
	credentials  *CredentialRepo
	devices      *DeviceRepo
	zones        *ZoneService
	userResolver UserResolver
	jwtService   JWTIssuer
}

func NewBiometricService(credentials *CredentialRepo, devices *DeviceRepo, zones *ZoneService, userResolver UserResolver, jwtService JWTIssuer) *BiometricService {
	return &BiometricService{
		credentials:  credentials,
		devices:      devices,
		zones:        zones,
		userResolver: userResolver,
		jwtService:   jwtService,
	}
}

func (s *BiometricService) AuthenticateBiometric(deviceID int, credentialHash string) (string, error) {
	// validate device exists and is active
	device, err := s.devices.GetDevice(deviceID)
	if err != nil {
		return "", err
	}
	if !device.Active {
		return "", ErrDeviceInactive
	}

	// validate credential exists and is not revoked
	credential, err := s.credentials.GetCredentialByHash(credentialHash)
	if err != nil {
		return "", err
	}
	if credential.Revoked {
		return "", ErrCredentialRevoked
	}

	// resolve user
	user, err := s.userResolver.VerifyUser(credential.UserID)
	if err != nil {
		return "", err
	}

	// delegate to HandleZoneEvent for session + event creation
	if err := s.zones.HandleZoneEvent(credential.UserID, device.ZoneID, user.Role, "enter", time.Now(), &deviceID, credential.EntryMethod); err != nil {
		return "", err
	}

	// generate JWT for the authenticated user
	// generate JWT only if device is the main entry point
	if device.IsEntryPoint {
		return s.jwtService.GenerateToken(user)
	}
	return "", nil
}
