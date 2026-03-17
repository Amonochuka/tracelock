package access

import (
	"database/sql"
	"errors"
	"fmt"
)

type ZoneRepo struct {
	db *sql.DB
}

func NewZoneRepo(db *sql.DB) *ZoneRepo {
	return &ZoneRepo{db: db}
}

func (z *ZoneRepo) GetMaximumCapacity(ZoneID int) (int, error) {
	var capacity int
	err := z.db.QueryRow("SELECT max_capacity FROM zones WHERE id = $1", ZoneID).Scan(&capacity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("zone not found")
		}
		return 0, err
	}
	return capacity, nil
}

func (z *ZoneRepo) CreateEvent(userID, zoneID int, action, status, hash, previousHash string) error {
	_, err := z.db.Exec(`
		INSERT INTO access_events (user_id, zone_id, action, status, hash, previous_hash)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, zoneID, action, status, hash, previousHash)

	if err != nil {
		return fmt.Errorf("CreateEvent insert failed: %w", err)
	}
	return nil
}
