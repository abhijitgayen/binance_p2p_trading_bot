package handlers

import (
	"fmt"

	"go_binance_bot/src/apis"
	"go_binance_bot/src/config"
	"go_binance_bot/src/db"
	"go_binance_bot/src/helpers/msg_gen"
	"go_binance_bot/src/helpers/priority_queue"
	"go_binance_bot/src/jobs"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func startHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("Handling /start command")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🤖 Welcome to the Binance C2C bot! 🚀\nUse /help to see available commands.")
	bot.Send(msg)
}

func runHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update, user *db.User) {
	log.Println("Handling /run command")

	userID := update.Message.From.ID

	jobManager := jobs.GetJobManager()
	if jobManager.IsJobRunning(userID) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ A job is already running for your user ID.")
		bot.Send(msg)
		return
	}

	apiKey, ok := user.BotConfig["api_key"].(string)
	if !ok || apiKey == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "api_key not found or is not a string in user.BotConfig")
		bot.Send(msg)
		log.Printf("api_key not found or is not a string in user.BotConfig")
		return
	}

	secretKey, ok := user.BotConfig["secret_key"].(string)
	if !ok || secretKey == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "secret_key not found or is not a string in user.BotConfig")
		bot.Send(msg)
		log.Printf("secret_key not found or is not a string in user.BotConfig")
		return
	}

	api := apis.NewBinanceAPI(config.BinanceURL, apiKey, secretKey, user.BotConfig)
	queue := priority_queue.NewPriorityQueue(2, time.Duration(config.CreateOrderInterval)*time.Second)

	jobManager.StartJob(userID, api, queue, bot, update.Message.Chat.ID)
	if jobManager.IsJobRunning(userID) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🤖 The bot is running in the background!")
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ Failed to start the job.")
		bot.Send(msg)
	}
}

func stopHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("Handling /stop command")

	userID := update.Message.From.ID
	jobManager := jobs.GetJobManager() // Assuming GetJobManager returns *JobManager

	// Attempt to stop the job and handle the case where no job exists.
	if err := jobManager.StopJob(userID); err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "No job is currently running.\n\n Use /run to start a job.")
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🚫 The bot has been stopped.\n\nUse the command /run to run the job.")
	bot.Send(msg)
}

func statusHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("Handling /status command")

	userID := update.Message.From.ID
	jobManager := jobs.GetJobManager()
	status := jobManager.GetJobStatus(userID)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("✅ The bot is up and running Status! 🚀\n\n%s", status))
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

	fmt.Println("key:", key, "value:", value)

	// Validate the key and type
	if !validateKeyAndType(key, value) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("⚠️ *Invalid configuration key or type: %s*", key))
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return
	}

	// Convert value to the appropriate data type
	convertedValue := convertValue(value)

	// Check if the key refers to a nested configuration
	if strings.Contains(key, ".") {
		keys := strings.Split(key, ".")
		if len(keys) == 2 {
			nestedKey := keys[0]
			subKey := keys[1]

			if nestedConfig, ok := user.BotConfig[nestedKey].(map[string]interface{}); ok {
				// Check if the value is the same
				if existingValue, exists := nestedConfig[subKey]; exists && existingValue == convertedValue {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("⚠️ *The configuration for %s.%s is already set to the given value.*", nestedKey, subKey))
					msg.ParseMode = "Markdown"
					bot.Send(msg)
					return
				}

				nestedConfig[subKey] = convertedValue
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
		// Check if the value is the same
		if existingValue, exists := user.BotConfig[key]; exists && existingValue == convertedValue {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("⚠️ *The configuration for %s is already set to the given value.*", key))
			msg.ParseMode = "Markdown"
			bot.Send(msg)
			return
		}

		user.BotConfig[key] = convertedValue
	}

	// Update the user configuration in the database
	db.UpdateUser(database, *user)

	// Send success message
	successMessage := fmt.Sprintf(
		"✅ *Configuration Updated!*\n\n🛠 *%s* → `%v`\n\n📌 To check the configuration, use the /get\\_config command.",
		key, convertedValue,
	)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, successMessage)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}

func validateKeyAndType(key, value string) bool {
	keys := strings.Split(key, ".")
	var field reflect.StructField
	var ok bool

	if len(keys) == 2 {
		field, ok = getFieldByTag(reflect.TypeOf(config.DefaultBotConfig.ExtraFilter), keys[1])
	} else {
		field, ok = getFieldByTag(reflect.TypeOf(config.DefaultBotConfig), keys[0])
	}

	if !ok {
		return false
	}

	// Validate the type
	switch field.Type.Kind() {
	case reflect.String:
		return true
	case reflect.Int:
		_, err := strconv.Atoi(value)
		return err == nil
	case reflect.Float64:
		_, err := strconv.ParseFloat(value, 64)
		return err == nil
	default:
		return false
	}
}

func getFieldByTag(t reflect.Type, tag string) (reflect.StructField, bool) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("json") == tag {
			return field, true
		}
	}

	return reflect.StructField{}, false
}

// convertValue attempts to convert a string value to an int, float64, or bool
func convertValue(value string) interface{} {
	// Try to convert to int
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}

	// Try to convert to float64
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}

	// Try to convert to bool
	if boolValue, err := strconv.ParseBool(value); err == nil {
		return boolValue
	}

	// Return the original string if no conversion was successful
	return value
}

// resetHandler sends an inline keyboard asking for confirmation.
func resetHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("Handling /reset command")

	// Create inline keyboard buttons for confirmation and cancellation
	confirmButton := tgbotapi.NewInlineKeyboardButtonData("✅ Confirm", "confirm_reset")
	cancelButton := tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "cancel_reset")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(confirmButton, cancelButton),
	)

	// Send confirmation message
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"⚠️ Are you sure you want to reset your bot configuration? This action cannot be undone.")
	msg.ReplyMarkup = keyboard

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

	aboutText := "🤖 *About This Trading Bot*\n\n" +
		"This bot is built exclusively for executing Binance trading orders. " +
		"To get started, please set up your configuration using the `/set_config` command. " +
		"Once configured, the bot will run in the background, automatically processing orders " +
		"based on your setup.\n\n" +
		"If any unrecognized activity or unexpected commands are detected, you will immediately " +
		"receive a notification so that you can review and take appropriate action.\n\n" +
		"Enjoy seamless, automated trading with our secure bot!"

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, aboutText)
	// Using Markdown for formatting.
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}
