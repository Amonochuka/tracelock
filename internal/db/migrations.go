
package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(db *sql.DB) error {
	// 1. Create the golang-migrate database driver instance for Postgres
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres migration driver: %w", err)
	}

	// 2. Point to your local migrations directory relative to where the app runs
	// This replaces reading the single "migrations/tables.sql" file
	migrationPath := "file://migrations"

	// 3. Initialize the migration runner instance for postgres
	m, err := migrate.NewWithDatabaseInstance(migrationPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to initialize migration instance: %w", err)
	}

	// 4. Apply all pending up migrations sequentially
	log.Println("checking for database migrations...")
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("database schema is completely up to date")
			return nil
		}
		return fmt.Errorf("migration pipeline aborted: %w", err)
	}

	log.Println("all migrations applied successfully to postgres!")
	return nil
}
