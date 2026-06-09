package access

import (
	"errors"
	"fmt"
	"log"
	"time"

	"tracelock/internal/models"
)

type ZoneService struct {
	repo *ZoneRepo
	hub  *Hub
}

func NewZoneService(repo *ZoneRepo, hub *Hub) *ZoneService {
	return &ZoneService{repo: repo, hub: hub}
}

// --zone management--
// list all existing zones
func (s *ZoneService) ListZones() ([]*models.Zone, error) {
	return s.repo.ListZones()
}

// get a particular zone
func (s *ZoneService) GetZone(zoneID int) (*models.ZoneOccupancy, error) {
	zone, err := s.repo.GetZone(zoneID)
	if err != nil {
		return nil, err
	}
	count, err := s.repo.CountActiveUsers(zoneID)
	if err != nil {
		return nil, err
	}
	users, err := s.repo.GetActiveUsersInZone(zoneID)
	if err != nil {
		return nil, err
	}
	return &models.ZoneOccupancy{Zone: *zone, ActiveCount: count, ActiveUsers: users}, nil
}

// create a new zone
func (s *ZoneService) CreateZone(name, description string, maxCapacity int) (*models.Zone, error) {
	return s.repo.CreateZone(name, description, maxCapacity)
}

// update a zone's details
func (s *ZoneService) UpdateZone(zoneID int, name, description string, maxCapacity int) (*models.Zone, error) {
	return s.repo.UpdateZone(zoneID, name, description, maxCapacity)
}

// delete a zone
func (s *ZoneService) DeleteZone(zoneID int) error {
	count, err := s.repo.CountActiveUsers(zoneID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrZoneHasActivity
	}
	return s.repo.DeleteZone(zoneID)
}

// --zone access permissions--
// grant access
func (s *ZoneService) GrantAccess(userID, zoneID, grantedBy int) error {
	// verify if zone exists
	if _, err := s.repo.GetZone(zoneID); err != nil {
		return err
	}
	return s.repo.GrantZoneAccess(userID, zoneID, grantedBy)
}

// revoke access
func (s *ZoneService) RevokeZoneAccess(userID, zoneID int) error {
	return s.repo.RevokeZoneAccess(userID, zoneID)
}

// list user access
func (s *ZoneService) ListUserAccess(userID int) ([]*models.Zone, error) {
	return s.repo.ListUserZoneAccess(userID)
}

// list zone users
func (s *ZoneService) ListZoneUsers(zoneID int) ([]*models.User, error) {
	// verify if zone exists
	if _, err := s.repo.GetZone(zoneID); err != nil {
		return nil, err
	}
	return s.repo.ListZoneUsers(zoneID)
}

// --access events--
func (s *ZoneService) HandleZoneEvent(userID, zoneID int, role, action string, timestamp time.Time, 
	deviceID *int, entryMethod string) error {
	if action == "enter" {
		// 1. Check permission
		allowed, err := s.repo.HasZoneAccess(userID, zoneID, role)
		if err != nil {
			return err
		}

		if !allowed {
			s.logDeniedEvent(userID, zoneID, action, timestamp, "no_access", deviceID, entryMethod)
			return ErrAccessDenied
		}

		// 2. Check if user is already in another zone (Auto-Exit Logic)
		// Use a new variable 'activeZoneID' so we don't overwrite the target 'zoneID'
		activeZoneID, err := s.repo.GetActiveSessionForUser(userID)
		if err != nil && !errors.Is(err, ErrNoActiveSession) {
			return err // Return real database errors
		}

		// If they have an active session in a DIFFERENT zone, auto-exit them first
		if err == nil && activeZoneID != zoneID {
			// Delete the old session
			if err := s.repo.DeleteSession(userID, activeZoneID); err != nil {
				return fmt.Errorf("auto-exit delete session failed: %w", err)
			}

			// Generate hash for the auto-exit event
			prevExitHash, err := s.repo.GetLastHash(activeZoneID)
			if err != nil && !errors.Is(err, ErrNoHashFound) {
				return fmt.Errorf("auto-exit get last hash failed: %w", err)
			}
			if errors.Is(err, ErrNoHashFound) {
				prevExitHash = ""
			}

			exitHash := GenerateHash(userID, activeZoneID, "exit", timestamp, prevExitHash, entryMethod)

			// Create the audit trail for the auto-exit
			err = s.repo.CreateEvent(userID, activeZoneID, "exit", "allowed", exitHash, prevExitHash, deviceID, entryMethod)
			if err != nil {
				return fmt.Errorf("auto-exit create event failed: %w", err)
			}
		}

		// 3. Check capacity of the target zone
		capacity, err := s.repo.GetMaximumCapacity(zoneID)
		if err != nil {
			return err
		}

		count, err := s.repo.CountActiveUsers(zoneID)
		if err != nil {
			return err
		}

		if capacity > 0 && count >= capacity {
			s.logDeniedEvent(userID, zoneID, action, timestamp, "zone_full", deviceID, entryMethod)
			return ErrZoneFull
		}
	}

	// 4. Mutate sessions based on incoming action
	switch action {
	case "enter":
		if err := s.repo.CreateSession(userID, zoneID); err != nil {
			if errors.Is(err, ErrUserAlreadyInZone) {
				s.logDeniedEvent(userID, zoneID, action, timestamp, "already_in_zone", deviceID, entryMethod)
			}
			return err
		}
	case "exit":
		if err := s.repo.DeleteSession(userID, zoneID); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid action: %s", action)
	}

	// 5. Log the main event (with secure cryptographic hash chain)
	previousHash, err := s.repo.GetLastHash(zoneID)
	if err != nil {
		if errors.Is(err, ErrNoHashFound) {
			previousHash = ""
		} else {
			return fmt.Errorf("get last hash: %w", err)
		}
	}

	hash := GenerateHash(userID, zoneID, action, timestamp, previousHash, entryMethod)
	if err := s.repo.CreateEvent(userID, zoneID, action, "allowed", hash, previousHash, deviceID, entryMethod); err != nil {
		return err
	}

	// broadcast zone state change to all WebSocket clients
	go s.broadcastZoneState(zoneID)

	// if auto-exit happened, broadcast the old zone too
	if action == "enter" {
		activeZoneID, _ := s.repo.GetActiveSessionForUser(userID)
		if activeZoneID != 0 && activeZoneID != zoneID {
			go s.broadcastZoneState(activeZoneID)
		}
	}

	return nil
}

// log denied entries
func (s *ZoneService) logDeniedEvent(userID, zoneID int, action string, timestamp time.Time, reason string, deviceID *int, entryMethod string) {
	_ = reason
	previousHash, err := s.repo.GetLastHash(zoneID)
	if err != nil {
		previousHash = ""
	}
	hash := GenerateHash(userID, zoneID, action+":denied", timestamp, previousHash, entryMethod)
	_ = s.repo.CreateEvent(userID, zoneID, action, "denied", hash, previousHash, deviceID, entryMethod)
}

// --event queries--
// list all events of a particular zone
func (s *ZoneService) ListZoneEvents(zoneID, limit, offset int) ([]*models.AccessEvent, int, error) {
	if _, err := s.repo.GetZone(zoneID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListZoneEvents(zoneID, limit, offset)
}

// list a user's activities across all zones
func (s *ZoneService) ListUserEvents(userID, limit, offset int) ([]*models.AccessEvent, int, error) {
	return s.repo.ListUserEvents(userID, limit, offset)
}

func (s *ZoneService) VerifyChain(zoneID int) (bool, int, error) {
	if _, err := s.repo.GetZone(zoneID); err != nil {
		return false, 0, err
	}
	return s.repo.VerifyChain(zoneID)
}

// broadcastZoneState fetches current zone state and broadcasts to all WebSocket clients.
func (s *ZoneService) broadcastZoneState(zoneID int) {
	zone, err := s.repo.GetZone(zoneID)
	if err != nil {
		log.Printf("broadcast skipped: could not fetch zone %d: %v", zoneID, err)
		return
	}

	count, err := s.repo.CountActiveUsers(zoneID)
	if err != nil {
		log.Printf("broadcast skipped: could not count users in zone %d: %v", zoneID, err)
		return
	}

	users, err := s.repo.GetActiveUsersInZone(zoneID)
	if err != nil {
		log.Printf("broadcast skipped: could not fetch users in zone %d: %v", zoneID, err)
		return
	}

	payload := models.ZoneOccupancy{
		Zone:        *zone,
		ActiveCount: count,
		ActiveUsers: users,
	}

	s.hub.BroadcastPayload(payload)
}

func (s *ZoneService) GetHub() *Hub {
    return s.hub
}