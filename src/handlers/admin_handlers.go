package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"go_binance_bot/src/apis"
	"go_binance_bot/src/config"
	"go_binance_bot/src/db"
	"go_binance_bot/src/helpers/priority_queue"
	"go_binance_bot/src/jobs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func isAdmin(userID int64) bool {
	return userID == config.NotifyUserId
}

func adminRunJobHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("Handling /admin_run_job command")

	if !isAdmin(update.Message.From.ID) {
		defaultHandler(bot, update)
		return
	}

	args := update.Message.CommandArguments()
	userID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ Invalid user ID.")
		bot.Send(msg)
		return
	}

	user := db.GetUser(database, userID)
	if user == nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ User not found.")
		bot.Send(msg)
		return
	}

	jobManager := jobs.GetJobManager()
	if jobManager.IsJobRunning(userID) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ A job is already running for this user ID.")
		bot.Send(msg)
		return
	}

	apiKey, ok := user.BotConfig["api_key"].(string)
	if !ok {
		log.Fatalf("api_key not found or is not a string in user.BotConfig")
	}

	secretKey, ok := user.BotConfig["secret_key"].(string)
	if !ok {
		log.Fatalf("secret_key not found or is not a string in user.BotConfig")
	}

	api := apis.NewBinanceAPI(config.BinanceURL, apiKey, secretKey, user.BotConfig)
	queue := priority_queue.NewPriorityQueue(2, time.Duration(config.CreateOrderInterval)*time.Second)

	jobManager.StartJob(userID, api, queue, bot, userID)
	if jobManager.IsJobRunning(userID) {
		msg := tgbotapi.NewMessage(userID, "🤖 The bot is running in the background for the specified user!")
		bot.Send(msg)
	} else {
		msg := tgbotapi.NewMessage(userID, "⚠️ Failed to start the job for the specified user.")
		bot.Send(msg)
	}
}

func adminStopJobHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("Handling /admin_stop_job command")

	if !isAdmin(update.Message.From.ID) {
		defaultHandler(bot, update)
		return
	}

	args := update.Message.CommandArguments()
	userID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ Invalid user ID.")
		bot.Send(msg)
		return
	}

	jobManager := jobs.GetJobManager()
	jobManager.StopJob(userID)

	msg := tgbotapi.NewMessage(userID, "🚫 The bot has been stopped for the specified user.")
	bot.Send(msg)
}

func adminJobStatusHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("Handling /admin_job_status command")

	if !isAdmin(update.Message.From.ID) {
		defaultHandler(bot, update)
		return
	}

	jobManager := jobs.GetJobManager()
	users, err := db.GetAllUsers(database) // Assuming you have a function to get all users
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ Failed to retrieve users.")
		bot.Send(msg)
		return
	}

	var statusMessage strings.Builder
	statusMessage.WriteString("✅ Job status for all users:\n\n")

	for _, user := range users {
		status := jobManager.GetJobStatus(user.UserID)
		statusMessage.WriteString(fmt.Sprintf("User ID: %d\nName: %s \n%s\n\n", user.UserID, user.FirstName, status))
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, statusMessage.String())
	bot.Send(msg)
}

// adminHelpHandler handles the /admin_help command for admins
func adminHelpHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	log.Println("🔹 Handling /admin_help command")

	// Check if the user is an admin
	if !isAdmin(update.Message.From.ID) {
		log.Printf("🚫 User %d is not an admin. Redirecting to default handler.", update.Message.From.ID)
		defaultHandler(bot, update)
		return
	}

	// Define the help message
	helpMessage := "*🔧 Admin Commands:*\n\n" +
		"📌 `/admin_run_job <user_id>`  Start a job for the specified user.\n" +
		"📌 `/admin_stop_job <user_id>` Stop the job for the specified user.\n" +
		"📌 `/admin_job_status`  Get the status of all users' jobs.\n" +
		"📌 `/admin_help` Show this help message. \n"

	// Create a new Telegram message
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpMessage)
	msg.ParseMode = "Markdown"

	// Send the message and handle errors
	if _, err := bot.Send(msg); err != nil {
		log.Printf("❌ Failed to send admin help message: %v", err)
	}
}
