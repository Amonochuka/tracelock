package access

import (
	"errors"
	"testing"
	"time"

	"tracelock/internal/models"
)

// mockZoneRepo implements ZoneRepository for testing.
// Each function field can be overridden per test to control behavior.
// Unset fields panic if called — that's intentional, it tells you
// your test exercised a code path you didn't expect.
type mockZoneRepo struct {
	hasZoneAccessFunc           func(userID, zoneID int, role string) (bool, error)
	getMaximumCapacityFunc      func(zoneID int) (int, error)
	countActiveUsersFunc        func(zoneID int) (int, error)
	createSessionFunc           func(userID, zoneID int) error
	deleteSessionFunc           func(userID, zoneID int) error
	getLastHashFunc             func(zoneID int) (string, error)
	createEventFunc             func(userID, zoneID int, action, status, hash, previousHash string, deviceID *int, entryMethod string) error
	getActiveSessionForUserFunc func(userID int) (int, error)
}

func (m *mockZoneRepo) HasZoneAccess(userID, zoneID int, role string) (bool, error) {
	if m.hasZoneAccessFunc != nil {
		return m.hasZoneAccessFunc(userID, zoneID, role)
	}
	return false, nil
}

func (m *mockZoneRepo) GetMaximumCapacity(zoneID int) (int, error) {
	if m.getMaximumCapacityFunc != nil {
		return m.getMaximumCapacityFunc(zoneID)
	}
	return 0, nil
}

func (m *mockZoneRepo) CountActiveUsers(zoneID int) (int, error) {
	if m.countActiveUsersFunc != nil {
		return m.countActiveUsersFunc(zoneID)
	}
	return 0, nil
}

func (m *mockZoneRepo) CreateSession(userID, zoneID int) error {
	if m.createSessionFunc != nil {
		return m.createSessionFunc(userID, zoneID)
	}
	return nil
}

func (m *mockZoneRepo) DeleteSession(userID, zoneID int) error {
	if m.deleteSessionFunc != nil {
		return m.deleteSessionFunc(userID, zoneID)
	}
	return nil
}

func (m *mockZoneRepo) GetLastHash(zoneID int) (string, error) {
	if m.getLastHashFunc != nil {
		return m.getLastHashFunc(zoneID)
	}
	return "", ErrNoHashFound
}

func (m *mockZoneRepo) CreateEvent(userID, zoneID int, action, status, hash, previousHash string, deviceID *int, entryMethod string) error {
	if m.createEventFunc != nil {
		return m.createEventFunc(userID, zoneID, action, status, hash, previousHash, deviceID, entryMethod)
	}
	return nil
}

func (m *mockZoneRepo) GetActiveSessionForUser(userID int) (int, error) {
	if m.getActiveSessionForUserFunc != nil {
		return m.getActiveSessionForUserFunc(userID)
	}
	return 0, ErrNoActiveSession
}

func (m *mockZoneRepo) GetZone(zoneID int) (*models.Zone, error) {
	return &models.Zone{ID: zoneID, MaxCapacity: 0}, nil
}

// --- unused-in-these-tests methods, stubbed to satisfy the ZoneRepository interface ---

func (m *mockZoneRepo) CreateZone(name, description string, maxCapacity int) (*models.Zone, error) {
	return nil, nil
}
func (m *mockZoneRepo) DeleteZone(zoneID int) error { return nil }

func (m *mockZoneRepo) GrantZoneAccess(userID, zoneID, grantedBy int) error   { return nil }
func (m *mockZoneRepo) RevokeZoneAccess(userID, zoneID int) error             { return nil }
func (m *mockZoneRepo) ListUserZoneAccess(userID int) ([]*models.Zone, error) { return nil, nil }
func (m *mockZoneRepo) ListZoneUsers(zoneID int) ([]*models.User, error)      { return nil, nil }
func (m *mockZoneRepo) ListZones() ([]*models.Zone, error)                    { return nil, nil }
func (m *mockZoneRepo) UpdateZone(zoneID int, name, description string, maxCapacity int) (*models.Zone, error) {
	return nil, nil
}
func (m *mockZoneRepo) GetActiveUsersInZone(zoneID int) ([]*models.User, error) { return nil, nil }
func (m *mockZoneRepo) ListZoneEvents(zoneID, limit, offset int) ([]*models.AccessEvent, int, error) {
	return nil, 0, nil
}
func (m *mockZoneRepo) ListUserEvents(userID, limit, offset int) ([]*models.AccessEvent, int, error) {
	return nil, 0, nil
}
func (m *mockZoneRepo) VerifyChain(zoneID int) (bool, int, error)                    { return true, 0, nil }
func (m *mockZoneRepo) ListZoneOccupancy() ([]*models.ZoneOccupancySnapshot, error)  { return nil, nil }
func (m *mockZoneRepo) GetZoneAnalytics(zoneID int) ([]*models.ZoneAnalytics, error) { return nil, nil }

// ============================================================
// Tests
// ============================================================

func TestHandleZoneEvent_AccessDenied(t *testing.T) {
	mockRepo := &mockZoneRepo{
		hasZoneAccessFunc: func(userID, zoneID int, role string) (bool, error) {
			return false, nil // user has no access
		},
		// access-denied path logs a denied event, so CreateEvent IS called here
		createEventFunc: func(u, z int, act, stat, h, ph string, d *int, em string) error {
			return nil
		},
		getLastHashFunc: func(zoneID int) (string, error) {
			return "", ErrNoHashFound
		},
	}

	service := NewZoneService(mockRepo, NewHub("*"))

	err := service.HandleZoneEvent(1, 1, "user", "enter", time.Now(), nil, "fingerprint")

	if !errors.Is(err, ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestHandleZoneEvent_AdminAlwaysAllowed(t *testing.T) {
	mockRepo := &mockZoneRepo{
		hasZoneAccessFunc: func(userID, zoneID int, role string) (bool, error) {
			return true, nil // admin bypasses access check
		},
		getActiveSessionForUserFunc: func(userID int) (int, error) {
			return 0, ErrNoActiveSession // not in any other zone — no auto-exit
		},
		getMaximumCapacityFunc: func(zoneID int) (int, error) {
			return 0, nil // unlimited capacity
		},
		countActiveUsersFunc: func(zoneID int) (int, error) {
			return 0, nil
		},
		createSessionFunc: func(userID, zoneID int) error {
			return nil
		},
		getLastHashFunc: func(zoneID int) (string, error) {
			return "", ErrNoHashFound // first event in this zone
		},
		createEventFunc: func(u, z int, act, stat, h, ph string, d *int, em string) error {
			return nil
		},
	}

	service := NewZoneService(mockRepo, NewHub("*"))

	err := service.HandleZoneEvent(1, 1, "admin", "enter", time.Now(), nil, "fingerprint")

	// explicit nil check — not just "is this not ErrAccessDenied"
	// catches ANY unexpected error on this success path, not just one specific case
	if err != nil {
		t.Errorf("expected admin to enter successfully with no error, got: %v", err)
	}
}

func TestHandleZoneEvent_ZoneFull(t *testing.T) {
	mockRepo := &mockZoneRepo{
		hasZoneAccessFunc: func(userID, zoneID int, role string) (bool, error) {
			return true, nil
		},
		getActiveSessionForUserFunc: func(userID int) (int, error) {
			return 0, ErrNoActiveSession
		},
		getMaximumCapacityFunc: func(zoneID int) (int, error) {
			return 5, nil
		},
		countActiveUsersFunc: func(zoneID int) (int, error) {
			return 5, nil // already at capacity
		},
		createEventFunc: func(u, z int, act, stat, h, ph string, d *int, em string) error {
			return nil // zone-full path also logs a denied event
		},
		getLastHashFunc: func(zoneID int) (string, error) {
			return "", ErrNoHashFound
		},
	}

	service := NewZoneService(mockRepo, NewHub("*"))

	err := service.HandleZoneEvent(1, 1, "user", "enter", time.Now(), nil, "fingerprint")

	if !errors.Is(err, ErrZoneFull) {
		t.Errorf("expected ErrZoneFull, got %v", err)
	}
}

func TestHandleZoneEvent_ExitAccepted(t *testing.T) {
	// real HandleZoneEvent skips HasZoneAccess entirely for action=="exit"
	// (exits are intentionally unconditional for physical safety reasons)
	// so we only mock what the exit branch actually calls
	mockRepo := &mockZoneRepo{
		deleteSessionFunc: func(userID, zoneID int) error {
			return nil // session removed cleanly
		},
		getLastHashFunc: func(zoneID int) (string, error) {
			return "", ErrNoHashFound // first event for this zone — valid happy path
		},
		createEventFunc: func(u, z int, act, stat, h, ph string, d *int, em string) error {
			return nil // exit event logged
		},
	}

	service := NewZoneService(mockRepo, NewHub("*"))

	err := service.HandleZoneEvent(1, 1, "user", "exit", time.Now(), nil, "fingerprint")

	if err != nil {
		t.Errorf("expected user to exit successfully with no error, got: %v", err)
	}
}

func TestHandleZoneEvent_ExitDeniedWhenNotInZone(t *testing.T) {
	// covers the bug fix: exit on a zone the user isn't actually in
	// should log a denied event and return ErrNoActiveSession, not
	// fail silently
	mockRepo := &mockZoneRepo{
		deleteSessionFunc: func(userID, zoneID int) error {
			return ErrNoActiveSession // user was never in this zone
		},
		getLastHashFunc: func(zoneID int) (string, error) {
			return "", ErrNoHashFound
		},
		createEventFunc: func(u, z int, act, stat, h, ph string, d *int, em string) error {
			return nil // denied event still gets logged
		},
	}

	service := NewZoneService(mockRepo, NewHub("*"))

	err := service.HandleZoneEvent(1, 1, "user", "exit", time.Now(), nil, "fingerprint")

	if !errors.Is(err, ErrNoActiveSession) {
		t.Errorf("expected ErrNoActiveSession, got %v", err)
	}
}

func TestHandleZoneEvent_AutoExitOnZoneSwap(t *testing.T) {
	var deletedZoneID int
	var deletedUserID int
	var createdEventZones []int

	mockRepo := &mockZoneRepo{
		hasZoneAccessFunc: func(userID, zoneID int, role string) (bool, error) {
			return true, nil
		},
		// user is currently in zone 1, trying to enter zone 2
		getActiveSessionForUserFunc: func(userID int) (int, error) {
			return 1, nil
		},
		deleteSessionFunc: func(userID, zoneID int) error {
			deletedUserID = userID
			deletedZoneID = zoneID
			return nil
		},
		getMaximumCapacityFunc: func(zoneID int) (int, error) {
			return 0, nil
		},
		countActiveUsersFunc: func(zoneID int) (int, error) {
			return 0, nil
		},
		createSessionFunc: func(userID, zoneID int) error {
			return nil
		},
		getLastHashFunc: func(zoneID int) (string, error) {
			return "", ErrNoHashFound
		},
		createEventFunc: func(u, z int, act, stat, h, ph string, d *int, em string) error {
			createdEventZones = append(createdEventZones, z)
			return nil
		},
	}

	service := NewZoneService(mockRepo, NewHub("*"))

	// user enters zone 2 while still active in zone 1
	err := service.HandleZoneEvent(99, 2, "user", "enter", time.Now(), nil, "fingerprint")

	if err != nil {
		t.Fatalf("expected zone swap to succeed with no error, got: %v", err)
	}

	if deletedUserID != 99 || deletedZoneID != 1 {
		t.Errorf("expected auto-exit to delete session for user 99 in zone 1, got user %d zone %d", deletedUserID, deletedZoneID)
	}

	// two events should be written: the auto-exit from zone 1, and the entry into zone 2
	if len(createdEventZones) != 2 {
		t.Fatalf("expected 2 events to be created (auto-exit + entry), got %d", len(createdEventZones))
	}
	if createdEventZones[0] != 1 {
		t.Errorf("expected first event to be the auto-exit from zone 1, got zone %d", createdEventZones[0])
	}
	if createdEventZones[1] != 2 {
		t.Errorf("expected second event to be entry into zone 2, got zone %d", createdEventZones[1])
	}
}

func TestHandleZoneEvent_NoAutoExitWhenEnteringSameZone(t *testing.T) {
	deleteSessionCalled := false

	mockRepo := &mockZoneRepo{
		hasZoneAccessFunc: func(userID, zoneID int, role string) (bool, error) {
			return true, nil
		},
		// user is already in zone 1, and is trying to "enter" zone 1 again
		getActiveSessionForUserFunc: func(userID int) (int, error) {
			return 1, nil
		},
		deleteSessionFunc: func(userID, zoneID int) error {
			deleteSessionCalled = true
			return nil
		},
		getMaximumCapacityFunc: func(zoneID int) (int, error) {
			return 0, nil
		},
		countActiveUsersFunc: func(zoneID int) (int, error) {
			return 0, nil
		},
		createSessionFunc: func(userID, zoneID int) error {
			return ErrUserAlreadyInZone // DB unique constraint catches the real duplicate
		},
		getLastHashFunc: func(zoneID int) (string, error) {
			return "", ErrNoHashFound
		},
		createEventFunc: func(u, z int, act, stat, h, ph string, d *int, em string) error {
			return nil // the denied event from ErrUserAlreadyInZone
		},
	}

	service := NewZoneService(mockRepo, NewHub("*"))

	err := service.HandleZoneEvent(1, 1, "user", "enter", time.Now(), nil, "fingerprint")

	if !errors.Is(err, ErrUserAlreadyInZone) {
		t.Errorf("expected ErrUserAlreadyInZone, got %v", err)
	}

	// auto-exit logic should never fire when the active zone equals the target zone
	if deleteSessionCalled {
		t.Error("expected no auto-exit delete when entering the same zone already occupied, but DeleteSession was called")
	}
}
