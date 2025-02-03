package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var (
	TelegramBotToken     string
	DatabasePath         string
	AuthorizedUsers      map[int64]bool
	BinanceURL           string
	ProcessQueueInterval int
	CallJobInterval      int
	CreateOrderInterval  int
)

func LoadSettings() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	DatabasePath = os.Getenv("DATABASE_PATH")
	if DatabasePath == "" {
		log.Fatal("DATABASE_PATH environment variable is required")
	}

	AuthorizedUsers = make(map[int64]bool)
	users := strings.Split(os.Getenv("AUTHORIZED_USERS"), ",")
	for _, user := range users {
		userID, err := strconv.ParseInt(user, 10, 64)
		if err != nil {
			log.Fatalf("Invalid user ID in AUTHORIZED_USERS: %v", err)
		}
		AuthorizedUsers[userID] = true
	}

	BinanceURL = os.Getenv("BINANCE_URL")
	if BinanceURL == "" {
		log.Fatal("BINANCE_URL environment variable is required")
	}

	ProcessQueueIntervalStr := os.Getenv("PROCESS_QUEUE_INTERVAL")
	if ProcessQueueIntervalStr == "" {
		log.Fatal("PROCESS_QUEUE_INTERVAL environment variable is required")
	}
	ProcessQueueInterval, err = strconv.Atoi(ProcessQueueIntervalStr)
	if err != nil {
		log.Fatalf("Invalid PROCESS_QUEUE_INTERVAL value: %v", err)
	}

	callJobIntervalStr := os.Getenv("CALL_JOB_INTERVAL")
	if callJobIntervalStr == "" {
		log.Fatal("CALL_JOB_INTERVAL environment variable is required")
	}
	CallJobInterval, err = strconv.Atoi(callJobIntervalStr)
	if err != nil {
		log.Fatalf("Invalid CALL_JOB_INTERVAL value: %v", err)
	}

	createOrderIntervalStr := os.Getenv("CREATE_ORDER_INTERVAL")
	if createOrderIntervalStr == "" {
		log.Fatal("CREATE_ORDER_INTERVAL environment variable is required")
	}
	CreateOrderInterval, err = strconv.Atoi(createOrderIntervalStr)
	if err != nil {
		log.Fatalf("Invalid CALL_JOB_INTERVAL value: %v", err)
	}

	// Log all loaded settings for verification
	log.Printf("*************Loaded settings****************\n")
	log.Printf("TELEGRAM_BOT_TOKEN: %s\n", TelegramBotToken)
	log.Printf("DATABASE_PATH: %s\n", DatabasePath)
	log.Printf("AUTHORIZED_USERS: %v\n", AuthorizedUsers)
	log.Printf("BINANCE_URL: %s\n", BinanceURL)
	log.Printf("PROCESS_QUEUE_INTERVAL: %d\n", ProcessQueueInterval)
	log.Printf("CALL_JOB_INTERVAL: %d\n", CallJobInterval)
	log.Printf("CREATE_ORDER_INTERVAL: %d\n", CreateOrderInterval)
	log.Printf("*************Loaded settings****************\n\n")
}
