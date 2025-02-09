package handlers

import (
	"database/sql"
	"encoding/json"
	"go_binance_bot/src/config"
	"go_binance_bot/src/db"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
)

// CallbackHandler processes the inline button callback queries.
// The database parameter must be passed to handle the reset update.
func CallbackHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	callback := update.CallbackQuery

	// Acknowledge the callback to remove the loading spinner on the client.
	answer := tgbotapi.NewCallback(callback.ID, "")
	if _, err := bot.Request(answer); err != nil {
		log.Printf("Failed to answer callback query: %v", err)
	}

	// Route the callback based on its data.
	switch callback.Data {
	case "confirm_reset":
		ResetUserConfig(bot, callback, database)
	case "cancel_reset":
		CancelReset(bot, callback)
	default:
		log.Printf("Unknown callback data: %s", callback.Data)
	}
}

// ResetUserConfig resets the user's configuration to the default
// and updates the original message with the result.
func ResetUserConfig(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, database *sql.DB) {
	userID := callback.From.ID

	// Convert the default configuration to a map.
	var botConfigMap map[string]interface{}
	configJSON, err := json.Marshal(config.DefaultBotConfig)
	if err != nil {
		log.Printf("Failed to marshal default bot config: %v", err)
		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID,
			"❌ Failed to reset configuration (marshal error).")
		bot.Send(editMsg)
		return
	}
	if err = json.Unmarshal(configJSON, &botConfigMap); err != nil {
		log.Printf("Failed to unmarshal default bot config: %v", err)
		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID,
			"❌ Failed to reset configuration (unmarshal error).")
		bot.Send(editMsg)
		return
	}

	// Create a new user configuration using the default configuration.
	user := &db.User{
		UserID:    userID,
		FirstName: callback.From.FirstName,
		LastName:  callback.From.LastName,
		BotConfig: botConfigMap,
	}

	// Update the user record in the database.
	if err := db.UpdateUser(database, *user); err != nil {
		log.Printf("Failed to update user: %v", err)
		editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID,
			"❌ Failed to reset configuration.")
		bot.Send(editMsg)
		return
	}

	// Update the original message with a success message and remove the inline keyboard.
	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID,
		"✅ Your bot configuration has been reset successfully.")
	editMsg.ReplyMarkup = nil
	if _, err := bot.Send(editMsg); err != nil {
		log.Printf("Failed to update message: %v", err)
	}
}

// CancelReset updates the original message to indicate that the reset has been cancelled.
func CancelReset(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	editMsg := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID,
		"❌ Reset action has been cancelled.")
	// Remove inline keyboard.
	editMsg.ReplyMarkup = nil
	if _, err := bot.Send(editMsg); err != nil {
		log.Printf("Failed to update message: %v", err)
	}
}
