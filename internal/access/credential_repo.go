package access

import (
	"database/sql"
	"errors"
	"fmt"

	"tracelock/internal/models"

	"github.com/lib/pq"
)

type CredentialRepo struct {
	db *sql.DB
}

func NewCredentialRepo(db *sql.DB) *CredentialRepo {
	return &CredentialRepo{db: db}
}

func (c *CredentialRepo) EnrollCredential(userID int, entryMethod, credentialHash string) (*models.BiometricCredential, error) {
	credential := &models.BiometricCredential{}
	err := c.db.QueryRow(`INSERT INTO biometric_credentials(user_id, entry_method, credential_hash)VALUES($1,$2,$3)
	RETURNING id, user_id, entry_method, credential_hash, enrolled_at, revoked`, userID, entryMethod,credentialHash ).
		Scan(&credential.ID, &credential.UserID, &credential.EntryMethod, &credential.CredentialHash, &credential.EnrolledAt, &credential.Revoked)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrCredentialExists
		}
		return nil, fmt.Errorf("create credential: %w", err)
	}
	return credential, nil
}

func (c *CredentialRepo) GetCredential(userID int, entryMethod string) (*models.BiometricCredential, error) {
	cdl := &models.BiometricCredential{}
	err := c.db.QueryRow(`SELECT id, user_id, entry_method, credential_hash, enrolled_at, revoked FROM biometric_credentials WHERE user_id = $1 AND entry_method = $2`, userID, entryMethod).
		Scan(&cdl.ID, &cdl.UserID, &cdl.EntryMethod, &cdl.CredentialHash, &cdl.EnrolledAt, &cdl.Revoked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCredentialNotFound
		}
		return nil, fmt.Errorf("get credentials: %w", err)
	}
	return cdl, nil
}

func (c *CredentialRepo) RevokeCredential(userID int, entryMethod string) error {
	res, err := c.db.Exec(`UPDATE biometric_credentials SET revoked = true WHERE user_id = $1 AND entry_method = $2`, userID, entryMethod)
	if err != nil {
		return fmt.Errorf("revoke credential: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrCredentialNotFound
	}
	return nil
}

func (c *CredentialRepo) ListUserCredentials(userID int) ([]*models.BiometricCredential, error) {
	rows, err := c.db.Query("SELECT id, user_id, entry_method, credential_hash, enrolled_at, revoked FROM biometric_credentials WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("list user credentials: %w", err)
	}
	defer rows.Close()

	var credentials []*models.BiometricCredential
	for rows.Next() {
		cdl := &models.BiometricCredential{}
		if err := rows.Scan(&cdl.ID, &cdl.UserID, &cdl.EntryMethod, &cdl.CredentialHash, &cdl.EnrolledAt, &cdl.Revoked); err != nil {
			return nil, fmt.Errorf("scan credentials: %w", err)
		}
		credentials = append(credentials, cdl)
	}
	return credentials, nil
}

// GetCredentialByHash finds a credential by its hash — used during device authentication
func (c *CredentialRepo) GetCredentialByHash(hash string) (*models.BiometricCredential, error) {
	cdl := &models.BiometricCredential{}
	err := c.db.QueryRow(`SELECT id, user_id, entry_method, credential_hash, enrolled_at, revoked 
		FROM biometric_credentials WHERE credential_hash = $1`, hash).
		Scan(&cdl.ID, &cdl.UserID, &cdl.EntryMethod, &cdl.CredentialHash, &cdl.EnrolledAt, &cdl.Revoked)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCredentialNotFound
		}
		return nil, fmt.Errorf("get credential by hash: %w", err)
	}
	return cdl, nil
}