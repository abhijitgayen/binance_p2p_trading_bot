package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Create the users table if it does not exist.
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		"user_id" INTEGER NOT NULL PRIMARY KEY,
		"first_name" TEXT,
		"last_name" TEXT,
		"bot_config" TEXT
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// Migration: Add the is_active column with a default of 0 for new rows.
	addColumnSQL := `ALTER TABLE users ADD COLUMN "is_active" INTEGER NOT NULL DEFAULT 0;`
	_, err = db.Exec(addColumnSQL)
	if err != nil {
		// It's common to see an error here if the migration has already been applied.
		log.Printf("Migration note: Could not add column is_active (it may already exist): %v", err)
	} else {
		// Immediately update all existing rows to have is_active = 1.
		updateSQL := `UPDATE users SET is_active = 1;`
		_, err = db.Exec(updateSQL)
		if err != nil {
			log.Fatalf("Failed to update existing users for is_active: %v", err)
		}
	}

	return db
}
