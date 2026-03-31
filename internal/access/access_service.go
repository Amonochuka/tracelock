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
		if err := s.repo.CreateSession(userID, zoneID); err != nil {
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
