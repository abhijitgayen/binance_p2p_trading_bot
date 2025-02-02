package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"go_binance_bot/src/apis"
	"go_binance_bot/src/config"
	"go_binance_bot/src/db"
	"go_binance_bot/src/helpers/msg_gen"
	"go_binance_bot/src/helpers/priority_queue"
	"go_binance_bot/src/jobs"

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

	var response string

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
		response = "Unknown command. Use /help to see available commands."
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
		bot.Send(msg)
	}
}

func startHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("Handling /start command")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🤖 Welcome to the Binance C2C bot! 🚀\nUse /help to see available commands.")
	bot.Send(msg)
}

func runHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, user *db.User) {
	log.Println("Handling /run command")

	userID := update.Message.From.ID

	apiKey, ok := user.BotConfig["api_key"].(string)
	if !ok {
		log.Fatalf("api_key not found or is not a string in user.BotConfig")
	}

	secretKey, ok := user.BotConfig["secret_key"].(string)
	if !ok {
		log.Fatalf("secret_key not found or is not a string in user.BotConfig")
	}

	api := apis.NewBinanceAPI("https://api.binance.com", apiKey, secretKey, user.BotConfig)
	queue := priority_queue.NewPriorityQueue(2, 5*time.Second)

	jobManager := jobs.GetJobManager()
	jobManager.StartJob(userID, api, queue)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🤖 The bot is running in the background!")
	bot.Send(msg)
}

func stopHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("Handling /stop command")

	userID := update.Message.From.ID
	jobManager := jobs.GetJobManager()
	jobManager.StopJob(userID)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🚫 The bot has been stopped.")
	bot.Send(msg)
}

func statusHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("Handling /status command")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅ The bot is up and running! 🚀")
	bot.Send(msg)
}

func getConfigHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, user *db.User) {
	log.Println("Handling /get_config command")
	if user == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "User not found.")
		bot.Send(msg)
		return
	}

	message := msg_gen.GenerateConfigMessage(user.BotConfig)
	message += "\n To update the configuration, use the /set\\_config command."
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

// setConfigHandler processes the /set_config command and updates the user's bot configuration.
func setConfigHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, user *db.User) {
	log.Println("⚙️ Handling /set_config command")

	if user == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ *User not found.*")
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return
	}

	args := strings.SplitN(update.Message.Text, " ", 3)
	if len(args) < 3 {
		helpMessage := "⚠️ *Incorrect Usage!*\n📌 Use the command in the following format:\n\n\t `/set_config <key> <value>`\n\n📝 *Examples:*\n\t🔹 `/set_config asset USDT`\n\t🔹 `/set_config extra_filter.price 90`\n\t🔹 `/set_config extra_filter.error_codes 83999,84000`"

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpMessage)
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return
	}

	key := args[1]
	value := args[2]

	// Check if the key refers to a nested configuration
	if strings.Contains(key, ".") {
		keys := strings.Split(key, ".")
		if len(keys) == 2 {
			nestedKey := keys[0]
			subKey := keys[1]

			if nestedConfig, ok := user.BotConfig[nestedKey].(map[string]interface{}); ok {
				nestedConfig[subKey] = value
				user.BotConfig[nestedKey] = nestedConfig
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ *Invalid nested configuration key!*")
				msg.ParseMode = "Markdown"
				bot.Send(msg)
				return
			}
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ *Invalid nested configuration key format!*")
			msg.ParseMode = "Markdown"
			bot.Send(msg)
			return
		}
	} else {
		user.BotConfig[key] = value
	}

	// Update the user configuration in the database
	db.UpdateUser(database, *user)

	// Send success message
	successMessage := fmt.Sprintf(
		"✅ *Configuration Updated!*\n\n🛠 *%s* → `%s`\n\n📌 To check the configuration, use the /get\\_config command.",
		key, value,
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, successMessage)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func resetHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("Handling /reset command")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🔄 The bot configuration has been reset.")
	bot.Send(msg)
}

func helpHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("Handling /help command")
	helpText := "🤖 *Here are some commands you can use:*\n\n"

	helpText += "📌 *Bot Control*\n"
	helpText += "  /start - Start the bot\n"
	helpText += "  /run - Run the bot\n"
	helpText += "  /stop - Stop the bot\n"
	helpText += "  /status - Check bot status\n\n"

	helpText += "⚙️ *Config Management*\n"
	helpText += "  /get\\_config - Get bot config\n"
	helpText += "  /set\\_config - Set bot config\n"
	helpText += "  /reset - Reset bot config\n\n"

	helpText += "ℹ️ *Other*\n"
	helpText += "  /help - Get help\n"
	helpText += "  /about - About this bot\n"

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func aboutHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("Handling /about command")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "A secure bot for seamless peer-to-peer trading on Binance, allowing authorized users to easily execute buy and sell orders.")
	bot.Send(msg)
}
