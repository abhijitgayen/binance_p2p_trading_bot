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

    return db
}