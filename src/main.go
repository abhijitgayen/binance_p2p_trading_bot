package main

import (
    "log"
    "os"
    "strings"
    "strconv"

    "github.com/joho/godotenv"
    "go_binance_bot/src/bot"
    "go_binance_bot/src/handlers"
    "go_binance_bot/src/db"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }
    
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
    }

    dbPath := os.Getenv("DATABASE_PATH")
    if dbPath == "" {
        log.Fatal("DATABASE_PATH environment variable is required")
    }
    database := db.InitDB(dbPath)
    defer database.Close()  // Ensure the database is closed when main exits
    handlers.SetDatabase(database)

    authorizedUsers := make(map[int64]bool)
    users := strings.Split(os.Getenv("AUTHORIZED_USERS"), ",")
    for _, user := range users {
        userID, err := strconv.ParseInt(user, 10, 64)
        if err != nil {
            log.Fatalf("Invalid user ID in AUTHORIZED_USERS: %v", err)
        }
        authorizedUsers[userID] = true
    }
    handlers.SetAuthorizedUsers(authorizedUsers)

    b, err := bot.NewBot(token)
    if err != nil {
        log.Fatalf("Failed to create bot: %v", err)
    }

    log.Println("Starting bot...")
    b.Start()
    log.Println("Bot started successfully")
}