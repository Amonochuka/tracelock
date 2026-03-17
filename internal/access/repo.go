package access

import (
	"database/sql"
	"errors"
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
