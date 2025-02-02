package db

import (
    "database/sql"
    "encoding/json"
    "log"
)

type User struct {
    UserID    int64
    FirstName string
    LastName  string
    BotConfig map[string]interface{}
}

func InsertUser(db *sql.DB, user User) {
    botConfigJSON, err := json.Marshal(user.BotConfig)
    if err != nil {
        log.Fatalf("Failed to marshal bot config: %v", err)
    }

    insertUserSQL := `INSERT INTO users (user_id, first_name, last_name, bot_config) VALUES (?, ?, ?, ?)`
    statement, err := db.Prepare(insertUserSQL)
    if err != nil {
        log.Fatalf("Failed to prepare statement: %v", err)
    }
    defer statement.Close()

    _, err = statement.Exec(user.UserID, user.FirstName, user.LastName, botConfigJSON)
    if err != nil {
        log.Fatalf("Failed to insert user: %v", err)
    }
}

func UpdateUser(db *sql.DB, user User) {
    botConfigJSON, err := json.Marshal(user.BotConfig)
    if err != nil {
        log.Fatalf("Failed to marshal bot config: %v", err)
    }

    updateUserSQL := `UPDATE users SET first_name = ?, last_name = ?, bot_config = ? WHERE user_id = ?`
    statement, err := db.Prepare(updateUserSQL)
    if err != nil {
        log.Fatalf("Failed to prepare statement: %v", err)
    }
    defer statement.Close()

    _, err = statement.Exec(user.FirstName, user.LastName, botConfigJSON, user.UserID)
    if err != nil {
        log.Fatalf("Failed to update user: %v", err)
    }
}

func GetUser(db *sql.DB, userID int64) *User {
    row := db.QueryRow("SELECT user_id, first_name, last_name, bot_config FROM users WHERE user_id = ?", userID)

    var user User
    var botConfigJSON string
    err := row.Scan(&user.UserID, &user.FirstName, &user.LastName, &botConfigJSON)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil
        }
        log.Fatalf("Failed to get user: %v", err)
    }

    err = json.Unmarshal([]byte(botConfigJSON), &user.BotConfig)
    if err != nil {
        log.Fatalf("Failed to unmarshal bot config: %v", err)
    }

    return &user
}