package db

import (
	"database/sql"
	"os"
)

func runMigrations(db *sql.DB) error{
	sqlBytes, err := os.ReadFile("migrations/tables.sql")
	if err != nil{
		return err
	}

	sqlStatement := string(sqlBytes)
	_, err = db.Exec(sqlStatement)
	if err != nil{
		return err
	}
	return nil
}