package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"go_binance_bot/src/config"
	"go_binance_bot/src/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var authorizedUsers = map[int64]bool{}
var database *sql.DB

func SetAuthorizedUsers(users map[int64]bool) {
	authorizedUsers = users
}

func SetDatabase(db *sql.DB) {
	database = db
}

func isAuthorized(userID int64) bool {
	return authorizedUsers[userID]
}

func HandleMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message == nil { // ignore non-message updates
		return
	}

	userID := update.Message.From.ID
	if !isAuthorized(userID) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🚫 You are not authorized to use this bot.")
		bot.Send(msg)

		adminMessage := fmt.Sprintf("Unauthorized access attempt:\nUser ID: %d\nUsername: %s\nFirst Name: %s\nLast Name: %s",
			userID, update.Message.From.UserName, update.Message.From.FirstName, update.Message.From.LastName)
		adminMsg := tgbotapi.NewMessage(config.NotifyUserId, adminMessage)
		bot.Send(adminMsg)
		return
	}

	// Check if user already exists in the database
	user := db.GetUser(database, userID)
	if user == nil {
		// Convert DefaultBotConfig to map[string]interface{}
		var botConfigMap map[string]interface{}
		configJSON, err := json.Marshal(config.DefaultBotConfig)
		if err != nil {
			log.Fatalf("Failed to marshal default bot config: %v", err)
		}
		err = json.Unmarshal(configJSON, &botConfigMap)
		if err != nil {
			log.Fatalf("Failed to unmarshal default bot config: %v", err)
		}

		// Insert new user data
		user = &db.User{
			UserID:    userID,
			FirstName: update.Message.From.FirstName,
			LastName:  update.Message.From.LastName,
			BotConfig: botConfigMap,
		}
		db.InsertUser(database, *user)
	}

	switch update.Message.Command() {
	case "start":
		startHandler(bot, update)
	case "stop":
		stopHandler(bot, update)
	case "run":
		runHandler(bot, update, user)
	case "status":
		statusHandler(bot, update)
	case "get_config":
		getConfigHandler(bot, update, user)
	case "set_config":
		setConfigHandler(bot, update, user)
	case "reset":
		resetHandler(bot, update, database)
	case "help":
		helpHandler(bot, update)
	case "about":
		aboutHandler(bot, update)
	case "admin_run_job":
		adminRunJobHandler(bot, update)
	case "admin_stop_job":
		adminStopJobHandler(bot, update)
	case "admin_job_status":
		adminJobStatusHandler(bot, update)
	case "admin_help":
		adminHelpHandler(bot, update)
	default:
		defaultHandler(bot, update)
	}
}

func defaultHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	response := "Unknown command. Use /help to see available commands."
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
	bot.Send(msg)
}
