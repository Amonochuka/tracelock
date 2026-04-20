package access

import (
	"fmt"
	"time"

	"tracelock/internal/models"
)

type ZoneService struct {
	repo *ZoneRepo
}

func NewZoneService(repo *ZoneRepo) *ZoneService {
	return &ZoneService{repo: repo}
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
	return s.repo.GrantAccess(userID, zoneID, grantedBy)
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

//--access events--

func (s *ZoneService) HandleZoneEvent(userID, zoneID int, role, action string, timestamp time.Time) error {
	if action == "enter" {
		// check permission
		allowed, err := s.repo.HasZoneAccess(userID, zoneID, role)
		if err != nil {
			return err
		}

		if !allowed {
			s.logDeniedEvent(userID, zoneID, action, timestamp, "no_access")
			return ErrAccessDenied
		}

		// check capacity
		capacity, err := s.repo.GetMaximumCapacity(zoneID)
		if err != nil {
			return err
		}

		count, err := s.repo.CountActiveUsers(zoneID)
		if err != nil {
			return err
		}

		if capacity > 0 && count >= capacity {
			s.logDeniedEvent(userID, zoneID, action, timestamp, "zone_full")
			return ErrZoneFull
		}
	}

	switch action {
	case "enter":
		if err := s.repo.CreateSession(userID, zoneID); err != nil {
			if err == ErrUserAlreadyInZone {
				s.logDeniedEvent(userID, zoneID, action, timestamp, "already_in_zone")
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

	previousHash, err := s.repo.GetLastHash(zoneID)
	if err != nil {
		if err == ErrNoHashFound {
			previousHash = ""
		} else {
			return fmt.Errorf("get last hash: %w", err)
		}
	}

	hash := GenerateHash(userID, zoneID, action, timestamp, previousHash)
	return s.repo.CreateEvent(userID, zoneID, action, "allowed", hash, previousHash)
}

// log denied entries
func (s *ZoneService) logDeniedEvent(userID, zoneID int, action string, timestamp time.Time, reason string) {
	_ = reason
	previousHash, err := s.repo.GetLastHash(zoneID)
	if err != nil {
		previousHash = ""
	}
	hash := GenerateHash(userID, zoneID, action+":denied", timestamp, previousHash)
	_ = s.repo.CreateEvent(userID, zoneID, action, "denied", hash, previousHash)
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

// check if a user can enter a zone
func (s *ZoneService) CanEnterRoom(userID, zoneID int, role string) error {
	ok, err := s.repo.HasZoneAccess(userID, zoneID, role)
	if err != nil {
		return err
	}
	if !ok {
		return ErrAccessDenied
	}
	return nil
}
