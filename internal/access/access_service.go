package access

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
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
			if errors.Is(err, sql.ErrNoRows) {
				return ErrZoneNotFound
			}
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
			//detect duplicate
			//use postgre code 23505 for unique violation
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code == "23505" {
					return ErrUserAlreadyInZone
				}
			}
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
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoActiveSession
		}
		return fmt.Errorf("cannot get last hash: %w", err)
	}
	hash := GenerateHash(userID, zoneID, action, timestamp, previousHash)
	return s.repo.CreateEvent(userID, zoneID, action, "success", hash, previousHash)
}
