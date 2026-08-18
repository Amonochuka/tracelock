package access

import (
	"errors"
	"testing"
	"time"

	"tracelock/internal/models"
)

type mockCredentialRepo struct {
	getCredentialByHashFunc func(hash string) (*models.BiometricCredential, error)
}

func (m *mockCredentialRepo) EnrollCredential(userID int, entryMethod, credentialHash string) (*models.BiometricCredential, error) {
	return nil, nil
}
func (m *mockCredentialRepo) GetCredential(userID int, entryMethod string) (*models.BiometricCredential, error) {
	return nil, nil
}
func (m *mockCredentialRepo) RevokeCredential(userID int, entryMethod string) error {
	return nil
}
func (m *mockCredentialRepo) ListUserCredentials(userID int) ([]*models.BiometricCredential, error) {
	return nil, nil
}
func (m *mockCredentialRepo) GetCredentialByHash(hash string) (*models.BiometricCredential, error) {
	if m.getCredentialByHashFunc != nil {
		return m.getCredentialByHashFunc(hash)
	}
	return nil, nil
}

type mockDeviceRepo struct {
	getDeviceFunc func(deviceID int) (*models.Device, error)
}

func (m *mockDeviceRepo) CreateDevice(zoneID int, name, deviceType, serial string) (*models.Device, error) {
	return nil, nil
}
func (m *mockDeviceRepo) GetDevice(deviceID int) (*models.Device, error) {
	if m.getDeviceFunc != nil {
		return m.getDeviceFunc(deviceID)
	}
	return nil, nil
}
func (m *mockDeviceRepo) ListZoneDevices(zoneID int) ([]*models.Device, error) {
	return nil, nil
}
func (m *mockDeviceRepo) UpdateDevice(deviceID int, name, deviceType, serial string) (*models.Device, error) {
	return nil, nil
}
func (m *mockDeviceRepo) DeactivateDevice(deviceID int) error {
	return nil
}
func (m *mockDeviceRepo) DeleteDevice(deviceID int) error {
	return nil
}

type mockUserResolver struct {
	verifyUserFunc func(id int) (*models.User, error)
}

func (m *mockUserResolver) VerifyUser(id int) (*models.User, error) {
	if m.verifyUserFunc != nil {
		return m.verifyUserFunc(id)
	}
	return nil, nil
}

type mockJWTIssuer struct{}

func (m *mockJWTIssuer) GenerateToken(user *models.User) (string, error) {
	return "", nil
}

func TestAuthenticateBiometric_DeviceTypeMismatchLogsDenied(t *testing.T) {
	var loggedReason *string
	var loggedEntryMethod string

	mockCredRepo := &mockCredentialRepo{
		getCredentialByHashFunc: func(hash string) (*models.BiometricCredential, error) {
			return &models.BiometricCredential{
				UserID:         1,
				EntryMethod:    "fingerprint",
				CredentialHash: "hash123",
				Revoked:        false,
			}, nil
		},
	}

	mockDevRepo := &mockDeviceRepo{
		getDeviceFunc: func(deviceID int) (*models.Device, error) {
			return &models.Device{
				ID:     deviceID,
				ZoneID: 2,
				Type:   "iris", // mismatches fingerprint
				Active: true,
			}, nil
		},
	}

	mockZoneRep := &mockZoneRepo{
		createChainedEventFunc: func(u, z int, act, stat string, reason *string, ts time.Time, dev *int, em string, us bool) error {
			loggedReason = reason
			loggedEntryMethod = em
			return nil
		},
	}

	zoneService := NewZoneService(mockZoneRep, nil)
	resolver := &mockUserResolver{}
	jwt := &mockJWTIssuer{}

	service := NewBiometricService(mockCredRepo, mockDevRepo, zoneService, resolver, jwt)

	_, err := service.AuthenticateBiometric(10, "hash123", "enter")
	if !errors.Is(err, ErrTypeMismatch) {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}

	if loggedReason == nil || *loggedReason != "device_type_mismatch" {
		t.Errorf("expected logged reason to be 'device_type_mismatch', got %v", loggedReason)
	}
	if loggedEntryMethod != "fingerprint" {
		t.Errorf("expected logged entry method to be 'fingerprint', got %s", loggedEntryMethod)
	}
}

func TestAuthenticateBiometric_RevokedCredentialLogsDenied(t *testing.T) {
	var loggedReason *string

	mockCredRepo := &mockCredentialRepo{
		getCredentialByHashFunc: func(hash string) (*models.BiometricCredential, error) {
			return &models.BiometricCredential{
				UserID:         1,
				EntryMethod:    "fingerprint",
				CredentialHash: "hash123",
				Revoked:        true, // revoked!
			}, nil
		},
	}

	mockDevRepo := &mockDeviceRepo{
		getDeviceFunc: func(deviceID int) (*models.Device, error) {
			return &models.Device{
				ID:     deviceID,
				ZoneID: 2,
				Type:   "fingerprint",
				Active: true,
			}, nil
		},
	}

	mockZoneRep := &mockZoneRepo{
		createChainedEventFunc: func(u, z int, act, stat string, reason *string, ts time.Time, dev *int, em string, us bool) error {
			loggedReason = reason
			return nil
		},
	}

	zoneService := NewZoneService(mockZoneRep, nil)
	resolver := &mockUserResolver{}
	jwt := &mockJWTIssuer{}

	service := NewBiometricService(mockCredRepo, mockDevRepo, zoneService, resolver, jwt)

	_, err := service.AuthenticateBiometric(10, "hash123", "enter")
	if !errors.Is(err, ErrCredentialRevoked) {
		t.Errorf("expected ErrCredentialRevoked, got %v", err)
	}

	if loggedReason == nil || *loggedReason != "credential_revoked" {
		t.Errorf("expected logged reason to be 'credential_revoked', got %v", loggedReason)
	}
}
