package access

import "fmt"

type ZoneService struct {
	repo *ZoneRepo
}

func NewZoneService(repo *ZoneRepo) *ZoneService {
	return &ZoneService{repo: repo}
}

func (s *ZoneService) EnterZone(userID, zoneID int) error {
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

	previous_hash, _ := s.repo.GetLastHash(zoneID)
	hash := GenerateHash(userID, zoneID, previous_hash)
	return s.repo.CreateEvent(userID, zoneID, "enter", "success", hash, previous_hash)

}
