package access

import (
	"fmt"
	"time"
	"tracelock/internal/models"
)

// UserResolver abstracts user lookup so BiometricService
// does not depend directly on the auth package
type UserResolver interface {
	VerifyUser(id int) (*models.User, error)
}

type BiometricService struct {
	credentials  *CredentialRepo
	devices      *DeviceRepo
	zones        *ZoneRepo
	userResolver UserResolver
}

func NewBiometricService(credentials *CredentialRepo, devices *DeviceRepo, zones *ZoneRepo, userResolver UserResolver) *BiometricService {
	return &BiometricService{
		credentials:  credentials,
		devices:      devices,
		zones:        zones,
		userResolver: userResolver,
	}
}

func (s *BiometricService) AuthenticateBiometric(deviceID int, credentialHash string) error {
	// validate device exists and is active
	device, err := s.devices.GetDevice(deviceID)
	if err != nil {
		return err
	}
	if !device.Active {
		return ErrDeviceInactive
	}

	//validate credential exists and is not revoked
	credential, err := s.credentials.GetCredentialByHash(credentialHash)
	if err != nil {
		return err
	}
	if credential.Revoked {
		return ErrCredentialRevoked
	}

	// resolve user from credential
	user, err := s.userResolver.VerifyUser(credential.UserID)
	if err != nil {
		return err
	}

	// verify user has access to device's zone
	allowed, err := s.zones.HasZoneAccess(credential.UserID, device.ZoneID, user.Role)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrAccessDenied
	}

	// create session
	if err := s.zones.CreateSession(credential.UserID, device.ZoneID); err != nil {
		return err
	}

	//  write audit event with device attribution
	previousHash, err := s.zones.GetLastHash(device.ZoneID)
	if err != nil {
		if err == ErrNoHashFound {
			previousHash = ""
		} else {
			return fmt.Errorf("get last hash: %w", err)
		}
	}

	hash := GenerateHash(credential.UserID, device.ZoneID, "enter", time.Now(), previousHash, credential.EntryMethod)
	return s.zones.CreateEvent(credential.UserID, device.ZoneID, "enter", "allowed", hash, previousHash, &deviceID, credential.EntryMethod)
}

