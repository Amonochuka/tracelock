package access

import (
	"errors"
	"fmt"
	"log"
	"time"

	"tracelock/internal/models"
)

type ZoneService struct {
	repo ZoneRepository
	hub  *Hub
}

func NewZoneService(repo ZoneRepository, hub *Hub) *ZoneService {
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
func (s *ZoneService) CreateZone(name, description string, maxCapacity int, requiresExitScan bool) (*models.Zone, error) {
	return s.repo.CreateZone(name, description, maxCapacity, requiresExitScan)
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

	// declared here (not inside the "enter" block) so it's still visible
	// at the bottom when broadcasting — without this, the auto-exit zone
	// becomes unreachable by the time we need to broadcast its updated state
	var activeZoneID int

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
		var sessionErr error
		activeZoneID, sessionErr = s.repo.GetActiveSessionForUser(userID)
		if sessionErr != nil && !errors.Is(sessionErr, ErrNoActiveSession) {
			return sessionErr // real database errors
		}

		// if they have an active session in a DIFFERENT zone, check if auto-exit is allowed
		if sessionErr == nil && activeZoneID != zoneID {
			// Check if the current zone requires an explicit exit scan
			requiresExit, err := s.repo.GetRequiresExitScan(activeZoneID)
			if err != nil {
				return err
			}
			if requiresExit {
				s.logDeniedEvent(userID, zoneID, action, timestamp, "requires_exit_scan", deviceID, entryMethod)
				return ErrRequiresExitScan
			}

			// delete the old session
			if err := s.repo.DeleteSession(userID, activeZoneID); err != nil {
				return fmt.Errorf("auto-exit delete session failed: %w", err)
			}

			// Append the auto-exit event atomically so it cannot fork the hash chain.
			if err := s.repo.CreateChainedEvent(userID, activeZoneID, "exit", "allowed", nil, timestamp, deviceID, entryMethod); err != nil {
				return fmt.Errorf("auto-exit create event failed: %w", err)
			}
		} else {
			// no auto-exit happened (user wasn't in another zone, or was
			// already in this same zone) — reset so the broadcast logic
			// at the bottom knows there's nothing extra to notify
			activeZoneID = 0
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
			// user tried to exit a zone they weren't actually in —
			// duplicate exit scan, misconfigured device, replayed request,
			// or a race condition. Rare in honest use, but a security
			// system should log the unusual case, not just the happy path
			if errors.Is(err, ErrNoActiveSession) {
				s.logDeniedEvent(userID, zoneID, action, timestamp, "not_in_zone", deviceID, entryMethod)
			}
			return err
		}
	default:
		return fmt.Errorf("invalid action: %s", action)
	}

	// 5. Log the main event atomically with its hash-chain predecessor.
	if err := s.repo.CreateChainedEvent(userID, zoneID, action, "allowed", nil, timestamp, deviceID, entryMethod); err != nil {
		return err
	}

	// broadcast zone state change to all WebSocket clients
	go s.broadcastZoneState(zoneID)

	// if auto-exit happened earlier, activeZoneID still holds that old zone's
	// ID (captured before its session was deleted) — broadcast it too, so its
	// WebSocket clients see the updated (lower) occupancy. We reuse the same
	// variable from step 2 instead of querying again, since by now the
	// session is already gone and a fresh query would just return nothing
	if activeZoneID != 0 {
		go s.broadcastZoneState(activeZoneID)
	}

	return nil
}

// log denied entries
func (s *ZoneService) logDeniedEvent(userID, zoneID int, action string, timestamp time.Time, reason string, deviceID *int, entryMethod string) {
	_ = s.repo.CreateChainedEvent(userID, zoneID, action, "denied", &reason, timestamp, deviceID, entryMethod)
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
	if err != nil || zone == nil {
		log.Printf("broadcast skipped: could not fetch zone %d: %v", zoneID, err)
		return
	}

	count, err := s.repo.CountActiveUsers(zoneID)
	if err != nil {
		log.Printf("broadcast skipped: could not count users in zone %d: %v", zoneID, err)
		return
	}

	payload := models.ZoneOccupancy{
		Zone:        *zone,
		ActiveCount: count,
	}

	s.hub.BroadcastPayload(payload)
}

func (s *ZoneService) GetHub() *Hub {
	return s.hub
}

func (s *ZoneService) ListZoneOccupancy() ([]*models.ZoneOccupancySnapshot, error) {
	return s.repo.ListZoneOccupancy()
}

func (s *ZoneService) GetZoneAnalytics(zoneID int) ([]*models.ZoneAnalytics, error) {
	return s.repo.GetZoneAnalytics(zoneID)
}
