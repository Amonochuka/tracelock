package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"tracelock/migrations"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func runMigrations(db *sql.DB) error {
	// 1. Create the golang-migrate database driver instance for Postgres
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres migration driver: %w", err)
	}

	// 2. Create the iofs driver with the embedded filesystem
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("could not create iofs driver: %w", err)
	}

	// 3. Initialize the migration runner instance for postgres using iofs
	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to initialize migration instance: %w", err)
	}

	// 4. Apply all pending up migrations sequentially
	log.Println("checking for database migrations...")
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("database schema is completely up to date")
			return nil
		}
		return fmt.Errorf("migration pipeline aborted: %w", err)
	}

	log.Println("all migrations applied successfully to postgres!")
	return nil
}
