package config

import (
    "log"
    "os"
    "strconv"
    "strings"

    "github.com/joho/godotenv"
)

var (
    TelegramBotToken string
    DatabasePath     string
    AuthorizedUsers  map[int64]bool
)

func LoadSettings() {
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

    TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
    if TelegramBotToken == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
    }

    DatabasePath = os.Getenv("DATABASE_PATH")
    if DatabasePath == "" {
        log.Fatal("DATABASE_PATH environment variable is required")
    }

    AuthorizedUsers = make(map[int64]bool)
    users := strings.Split(os.Getenv("AUTHORIZED_USERS"), ",")
    for _, user := range users {
        userID, err := strconv.ParseInt(user, 10, 64)
        if err != nil {
            log.Fatalf("Invalid user ID in AUTHORIZED_USERS: %v", err)
        }
        AuthorizedUsers[userID] = true
    }
}