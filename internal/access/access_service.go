package access

import (
	"fmt"
	"time"
)

type ZoneService struct {
	repo *ZoneRepo
}

func NewZoneService(repo *ZoneRepo) *ZoneService {
	return &ZoneService{repo: repo}
}

func (s *ZoneService) HandleZoneEvent(userID, zoneID int, action string, timestamp time.Time) error {
	if action == "enter" {
		capacity, err := s.repo.GetMaximumCapacity(zoneID)
		if err != nil {
			return err
		}

		count, err := s.repo.CountActiveUsers(zoneID)
		if err != nil {
			return err
		}

		if capacity > 0 && count >= capacity {
			return ErrZoneFull
		}
	}

	switch action {
	case "enter":
		err := s.repo.CreateSession(userID, zoneID)
		if err != nil {
			return err
		}
	case "exit":
		err := s.repo.DeleteSession(userID, zoneID)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid action: %s", action)
	}

	previousHash, err := s.repo.GetLastHash(zoneID)
	if err != nil {
		return fmt.Errorf("cannot get last hash: %w", err)
	}
	hash := GenerateHash(userID, zoneID, action, timestamp, previousHash)
	return s.repo.CreateEvent(userID, zoneID, action, "success", hash, previousHash)
}

func (s *ZoneService) CanEnterRoom(userID, zoneID int) error {
	ok, err := s.repo.HasZoneAccess(userID, zoneID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrAccessDenied
	}
	return nil
}
