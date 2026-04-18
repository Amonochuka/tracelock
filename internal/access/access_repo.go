package access

import (
	"database/sql"
	"errors"
	"fmt"

	"tracelock/internal/models"

	"github.com/lib/pq"
)

type ZoneRepo struct {
	db *sql.DB
}

func NewZoneRepo(db *sql.DB) *ZoneRepo {
	return &ZoneRepo{db: db}
}

// --zone creation--
// create a new zone
func (z *ZoneRepo) CreateZone(name, description string, maxCapacity int) (*models.Zone, error) {
	zone := &models.Zone{}
	err := z.db.QueryRow(`INSERT INTO zones(name, description, max_capacity)
		VALUES($1,$2,$3) RETURNING id, name, description, max_capacity, created_at`,
		name, description, maxCapacity).
		Scan(&zone.ID, &zone.Name, &zone.Description, &zone.MaxCapacity, &zone.CreatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrZoneNameExists
		}
		return nil, fmt.Errorf("create zone: %w", err)
	}
	return zone, nil
}

// delete a zonehow methe code thats missing from my orginal ac
func (z *ZoneRepo) DeleteZone(zoneID int) error {
	res, err := z.db.Exec(`DELETE FROM zones WHERE zone_id = $1`, zoneID)
	if err != nil {
		return fmt.Errorf("delete zone: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrZoneNotFound
	}
	return nil
}

// get a specific zone
func (z *ZoneRepo) GetZone(zoneID int) (*models.Zone, error) {
	zone := &models.Zone{}
	err := z.db.QueryRow(`SELECT id, name, description, max_capacity, created_at
		FROM zones WHERE id = $1`, zoneID).
		Scan(&zone.ID, &zone.Name, &zone.Description, &zone.MaxCapacity, &zone.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrZoneNotFound
		}
		return nil, fmt.Errorf("get zone: %w", err)
	}
	return zone, nil
}

// check a zone's max capacity
func (z *ZoneRepo) GetMaximumCapacity(zoneID int) (int, error) {
	var capacity int
	err := z.db.QueryRow("SELECT max_capacity FROM zones WHERE id = $1", zoneID).Scan(&capacity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrZoneNotFound
		}
		return 0, fmt.Errorf("get max_capacity: %w", err)
	}
	return capacity, nil
}

// CreateEvent writes an access event to the audit log.
func (z *ZoneRepo) CreateEvent(userID, zoneID int, action, status, hash, previousHash string) error {
	_, err := z.db.Exec(`
        INSERT INTO access_events (user_id, zone_id, action, status, hash, previous_hash)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, userID, zoneID, action, status, hash, previousHash)
	if err != nil {
		return fmt.Errorf("create event: %w", err)
	}
	return nil
}

// GetLastHash retrieves the most recent event hash for a zone to chain the next event.
func (z *ZoneRepo) GetLastHash(zoneID int) (string, error) {
	var hash string
	err := z.db.QueryRow(`
        SELECT hash FROM access_events WHERE zone_id = $1
        ORDER BY timestamp DESC LIMIT 1`, zoneID).Scan(&hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNoHashFound
		}
		return "", fmt.Errorf("get last hash: %w", err)
	}
	return hash, nil
}

// CreateSession registers a user as actively inside a zone.
func (z *ZoneRepo) CreateSession(userID, zoneID int) error {
	_, err := z.db.Exec(`
        INSERT INTO active_sessions (user_id, zone_id)
        VALUES ($1, $2)
    `, userID, zoneID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrUserAlreadyInZone
		}
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// DeleteSession removes a user's active session when they exit a zone.
func (z *ZoneRepo) DeleteSession(userID, zoneID int) error {
	res, err := z.db.Exec(`
        DELETE FROM active_sessions WHERE user_id = $1 AND zone_id = $2`, userID, zoneID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNoActiveSession
	}
	return nil
}

// check current users in a certain zone
func (z *ZoneRepo) CountActiveUsers(zoneID int) (int, error) {
	var count int
	err := z.db.QueryRow(`SELECT COUNT(*) FROM active_sessions WHERE zone_id = $1`, zoneID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active users: %w", err)
	}
	return count, nil
}

// ---zone access permissions--
// check if a user has been granted permission to enter a certain zone
func (z *ZoneRepo) HasZoneAccess(userID, zoneID int, role string) (bool, error) {
	if role == "admin" {
		return true, nil
	}

	var exists bool
	err := z.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM user_zone_access 
		WHERE user_id = $1 AND zone_id = $2)`, userID, zoneID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check zone access: %w", err)
	}
	return exists, nil
}

// grant zone access
func (z *ZoneRepo) GrantAccess(userID, zoneID, granted_by int) error {
	_, err := z.db.Exec(`INSERT INTO user_zone_access(user_id, zone_id, granted_by)VALUES($1, $2, $3)
						ON CONFLICT DO NOTHING`, userID, zoneID, granted_by)
	if err != nil {
		return fmt.Errorf("grant zone access: %w", err)
	}
	return nil
}

// revoke access to a room
func (z *ZoneRepo) RevokeZoneAccess(userID, zoneID int) error {
	res, err := z.db.Exec(`DELETE FROM user_zone_access WHERE user_id = $1 AND zone_id = $2`, userID, zoneID)
	if err != nil {
		return fmt.Errorf("revoke zone access:%w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrAccessNotFound
	}
	return nil
}

// lists all zones a user has been granted access to
func (z *ZoneRepo) ListUserZoneAccess(userID int) ([]*models.Zone, error) {
	rows, err := z.db.Query(`SELECT zo.id, zo.name, zo.description, zo.max_capacity, zo.created_at
							FROM zones zo INNER JOIN user_zone_access uza ON uza.zone_id = zo.id
							WHERE uza.user_id = $1 ORDER BY zo.id`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user zone access: %w", err)
	}
	defer rows.Close()

	var zones []*models.Zone
	for rows.Next() {
		zo := &models.Zone{}
		if err := rows.Scan(&zo.ID, &zo.Name, &zo.Description, &zo.MaxCapacity, &zo.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan zone: %w", err)
		}
		zones = append(zones, zo)
	}
	return zones, nil
}

// list all users granted access to a zone
func (z *ZoneRepo) ListZoneUsers(zoneID int) ([]*models.User, error) {
	rows, err := z.db.Query(`SELECT u.id, u.name, u.email, u.role, u.created_at
							FROM users u INNER JOIN user_zone_access uza ON uza.user_id = u.id
							WHERE uza.zone_id = $1 ORDER BY u.name`, zoneID)
	if err != nil {
		return nil, fmt.Errorf("list zone users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan zone: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}

// list all zones
func (z *ZoneRepo) ListZones() ([]*models.Zone, error) {
	rows, err := z.db.Query(`SELECT id, name, description, max_capacity, created_at
		FROM zones ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list zones: %w", err)
	}
	defer rows.Close()

	var zones []*models.Zone
	for rows.Next() {
		zo := &models.Zone{}
		if err := rows.Scan(&zo.ID, &zo.Name, &zo.Description, &zo.MaxCapacity, &zo.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan zone: %w", err)
		}
		zones = append(zones, zo)
	}
	return zones, nil
}

// UpdateZone updates a zone's details.
func (z *ZoneRepo) UpdateZone(zoneID int, name, description string, maxCapacity int) (*models.Zone, error) {
	zone := &models.Zone{}
	err := z.db.QueryRow(`UPDATE zones SET name=$1, description=$2, max_capacity=$3
		WHERE id=$4 RETURNING id, name, description, max_capacity, created_at`,
		name, description, maxCapacity, zoneID).
		Scan(&zone.ID, &zone.Name, &zone.Description, &zone.MaxCapacity, &zone.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrZoneNotFound
		}
		return nil, fmt.Errorf("update zone: %w", err)
	}
	return zone, nil
}

// GetActiveUsersInZone returns all users currently inside a zone.
func (z *ZoneRepo) GetActiveUsersInZone(zoneID int) ([]models.User, error) {
	rows, err := z.db.Query(`SELECT u.id, u.name, u.email, u.role, u.created_at
		FROM users u INNER JOIN active_sessions s ON s.user_id = u.id
		WHERE s.zone_id = $1 ORDER BY s.entered_at`, zoneID)
	if err != nil {
		return nil, fmt.Errorf("get active users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}
