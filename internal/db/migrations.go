package db

import (
	"database/sql"
	"fmt"
	"os"
)

func runMigrations(db *sql.DB) error {
	sqlBytes, err := os.ReadFile("migrations/tables.sql")
	if err != nil {
		return fmt.Errorf("reading migrations: %w", err)
	}

	sqlStatement := string(sqlBytes)
	_, err = db.Exec(sqlStatement)
	if err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	return nil
}
