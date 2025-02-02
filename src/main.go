package main

import (
	"log"

	"go_binance_bot/src/bot"
	"go_binance_bot/src/config"
	"go_binance_bot/src/db"
	"go_binance_bot/src/handlers"
)

func main() {
	config.LoadSettings()

	database := db.InitDB(config.DatabasePath)
	defer database.Close()
	handlers.SetDatabase(database)

	handlers.SetAuthorizedUsers(config.AuthorizedUsers)

	b, err := bot.NewBot(config.TelegramBotToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Println("Starting bot...")
	b.Start()
	log.Println("Bot started successfully")
}
