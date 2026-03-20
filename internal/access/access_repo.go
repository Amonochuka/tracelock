package access

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

type ZoneRepo struct {
	db *sql.DB
}

func NewZoneRepo(db *sql.DB) *ZoneRepo {
	return &ZoneRepo{db: db}
}

func (z *ZoneRepo) GetMaximumCapacity(zoneID int) (int, error) {
	var capacity int
	err := z.db.QueryRow("SELECT max_capacity FROM zones WHERE id = $1", zoneID).Scan(&capacity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrZoneNotFound
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

func (z *ZoneRepo) GetLastHash(zoneID int) (string, error) {
	var hash string
	err := z.db.QueryRow(`SELECT hash FROM access_events WHERE zone_id = $1
	ORDER BY timestamp DESC LIMIT 1`, zoneID).Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return hash, nil
}

func (z *ZoneRepo) CreateSession(userID, zoneID int) error {
	_, err := z.db.Exec(`
		INSERT INTO active_sessions (user_id, zone_id)
		VALUES ($1, $2)
	`, userID, zoneID)

	if err != nil {
		//detect duplicate
		//use postgre code 23505 for unique violation
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return ErrUserAlreadyInZone
			}
		}
		return fmt.Errorf("CreateSession insert failed: %w", err)
	}
	return nil
}

func (z *ZoneRepo) DeleteSession(userID, zoneID int) error {
	res, err := z.db.Exec(`
		DELETE FROM active_sessions WHERE user_id = $1 AND zone_id = $2`, userID, zoneID)

	if err != nil {
		return fmt.Errorf("delete session failed: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNoActiveSession
	}

	return nil
}

func (z *ZoneRepo) CountActiveUsers(zoneID int) (int, error) {
	var count int
	err := z.db.QueryRow(`SELECT COUNT(*) FROM active_sessions WHERE zone_id = $1`, zoneID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active users failed: %w", err)
	}
	return count, nil
}
