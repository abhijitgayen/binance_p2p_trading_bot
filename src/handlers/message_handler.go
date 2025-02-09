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
			log.Printf("Failed to marshal default bot config: %v", err)
		}
		err = json.Unmarshal(configJSON, &botConfigMap)
		if err != nil {
			log.Printf("Failed to unmarshal default bot config: %v", err)
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

	// Get the command from the message.
	command := update.Message.Command()

	// First, check if the command is one of the simple user commands.
	if isSimpleUserCommand(command) {
		handleSimpleUser(bot, update, user)
		return
	}

	// Next, check if the command is an admin command.
	if isAdminCommand(command) {
		handleAdminUser(bot, update)
		return
	}

	// If it doesn’t match any known command, use the default (unknown) handler.
	defaultHandler(bot, update)
}

// Helper function: returns true if the command is for a simple user.
func isSimpleUserCommand(cmd string) bool {
	switch cmd {
	case "start", "stop", "run", "status", "get_config", "set_config", "reset", "help", "about":
		return true
	}
	return false
}

// Helper function: returns true if the command is for an admin user.
func isAdminCommand(cmd string) bool {
	switch cmd {
	case "admin_run_job", "admin_stop_job", "admin_job_status", "admin_help":
		return true
	}
	return false
}

// Handling simple user commands
func handleSimpleUser(bot *tgbotapi.BotAPI, update tgbotapi.Update, user *db.User) {
	if user.IsActive == 0 {
		inActiveHandler(bot, update)
		return
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
		resetHandler(bot, update)
	case "help":
		helpHandler(bot, update)
	case "about":
		aboutHandler(bot, update)
	default:
		defaultHandler(bot, update)
	}
}

// Handling admin commands
func handleAdminUser(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if !isAdmin(update.Message.From.ID) {
		defaultHandler(bot, update)
		return
	}

	switch update.Message.Command() {
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

// Fallback handler for unknown commands.
func defaultHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	response := "Unknown command. Use /help to see available commands."
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
	bot.Send(msg)
}

func inActiveHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ You are not active. Please contact with the admin.")
	bot.Send(msg)
}
