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

func (s *ZoneService) EnterZone(userID, zoneID int, action string, timestamp time.Time) error {
	capacity, err := s.repo.GetMaximumCapacity(zoneID)
	if err != nil {
		return err
	}

	count, err := s.repo.CountActiveUsers(zoneID)
	if err != nil {
		return err
	}

	if count >= capacity {
		return fmt.Errorf("zone is full")
	}

	err = s.repo.CreateSession(userID, zoneID)
	if err != nil {
		return err
	}

	previousHash, err := s.repo.GetLastHash(zoneID)
	if err != nil {
		return fmt.Errorf("cannot get last hash: %w", err)
	}
	hash := GenerateHash(userID, zoneID, action, timestamp, previousHash)
	return s.repo.CreateEvent(userID, zoneID, "enter", "success", hash, previousHash)
}
