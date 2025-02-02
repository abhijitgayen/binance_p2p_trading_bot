package utils

import (
	"log"
	"os"
)

// LogInfo logs informational messages to the console
func LogInfo(message string) {
	log.Println("INFO:", message)
}

// LogError logs error messages to the console and exits the application
func LogError(err error) {
	log.Println("ERROR:", err)
	os.Exit(1)
}

// FormatMessage formats a message for sending to Telegram
func FormatMessage(chatID int64, text string) string {
	return fmt.Sprintf("Chat ID: %d, Message: %s", chatID, text)
}