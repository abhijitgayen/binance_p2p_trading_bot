package handlers

import (
    "log"
    "strings"
    "fmt"
    "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "go_binance_bot/src/db"
    "database/sql"
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
    existingUser := db.GetUser(database, userID)
    if existingUser == nil {
        user := db.User{
            UserID:    userID,
            FirstName: update.Message.From.FirstName,
            LastName:  update.Message.From.LastName,
            ExtraInfo: "Some extra info",
        }
        db.InsertUser(database, user)
    }

    switch update.Message.Command() {
    case "start":
        startHandler(bot, update)
    case "stop":
        stopHandler(bot, update)
    case "run":
        runHandler(bot, update)
    case "status":
        statusHandler(bot, update)
    case "get_config":
        getConfigHandler(bot, update)
    case "set_config":
        setConfigHandler(bot, update)
    case "reset":
        resetHandler(bot, update)
    case "help":
        helpHandler(bot, update)
    case "about":
        aboutHandler(bot, update)
    default:
        response := "Unknown command. Use /help to see available commands."
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, response)
        bot.Send(msg)
    }
}

func startHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
    log.Println("Handling /start command")
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🤖 Welcome to the Binance C2C bot! 🚀\nUse /help to see available commands.")
    bot.Send(msg)
}

func stopHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
    log.Println("Handling /stop command")
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🚫 The bot has been stopped.")
    bot.Send(msg)
}

func runHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
    log.Println("Handling /run command")
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🤖 The bot is running in the background!")
    bot.Send(msg)
}

func statusHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
    log.Println("Handling /status command")
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "✅ The bot is up and running! 🚀")
    bot.Send(msg)
}

func getConfigHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
    log.Println("Handling /get_config command")
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🔧 Here is the current configuration.")
    bot.Send(msg)
}

func setConfigHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
    log.Println("Handling /set_config command")
    args := strings.Split(update.Message.Text, " ")
    if len(args) < 3 {
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Usage: /set_config <key> <value>")
        bot.Send(msg)
        return
    }

    key := args[1]
    value := strings.Join(args[2:], " ")

    msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Configuration updated: %s = %s", key, value))
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
    helpText += "  /get_config - Get bot config\n"
    helpText += "  /set_config - Set bot config\n"
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

func defaultHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Invalid command! Use /help to see available commands.")
    bot.Send(msg)
}
