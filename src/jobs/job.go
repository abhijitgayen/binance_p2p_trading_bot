package jobs

import (
	"fmt"
	"go_binance_bot/src/apis"
	"go_binance_bot/src/helpers/priority_queue"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AdInfo struct {
	Ad       map[string]interface{}
	HasError bool
	Err      map[string]interface{}
}

type Job struct {
	BinanceAPI *apis.BinanceAPI
	Queue      *priority_queue.PriorityQueue
	stopChan   chan struct{}
	adsTracker map[string]*AdInfo
	wg         sync.WaitGroup
	bot        *tgbotapi.BotAPI
	chatID     int64
}

func NewJob(api *apis.BinanceAPI, queue *priority_queue.PriorityQueue, bot *tgbotapi.BotAPI, chatID int64) *Job {
	return &Job{
		BinanceAPI: api,
		Queue:      queue,
		stopChan:   make(chan struct{}),
		adsTracker: make(map[string]*AdInfo),
		bot:        bot,
		chatID:     chatID,
	}
}

func (j *Job) Run() {
	j.wg.Add(1)
	go func() {
		defer j.wg.Done()
		for {
			select {
			case <-j.stopChan:
				return
			default:
				j.Queue.ProcessTasks()
				time.Sleep(1 * time.Second) // Adjust the sleep time as needed
			}
		}
	}()

	for {
		select {
		case <-j.stopChan:
			return
		default:
			j.ListAdsAndCreateOrders("USDT", "INR", 1, 2, "BUY")
			time.Sleep(1 * time.Second) // Adjust the sleep time as needed
		}
	}
}

func (j *Job) Stop() {
	close(j.stopChan)
}

func (j *Job) ListAdsAndCreateOrders(asset, fiat string, page, rows int, tradeType string) {
	if j.BinanceAPI.Config == nil {
		log.Fatalf("Binance API config not found")
	}

	// Extract error codes to ignore from the config
	extraFilter, ok := j.BinanceAPI.Config["extra_filter"].(map[string]interface{})
	if !ok {
		log.Fatalf("extra_filter not found or is not a map in config")
	}

	errorCodesStr, ok := extraFilter["error_codes"].(string)
	if !ok {
		log.Fatalf("error_codes not found or is not a string in extra_filter")
	}

	errorCodesList := strings.Split(errorCodesStr, ",")

	errorIgnore := make(map[int]bool)
	for _, codeStr := range errorCodesList {
		code, err := strconv.Atoi(strings.TrimSpace(codeStr))
		if err != nil {
			log.Fatalf("Failed to parse error code: %v", err)
		}
		errorIgnore[code] = true
	}

	// Fetch ads from Binance API
	adsResponse, err := j.BinanceAPI.SearchAds(asset, fiat, page, rows, tradeType)
	if err != nil {
		log.Fatalf("Failed to fetch ads: %v", err)
	}

	// Extract ads from response
	ads, ok := adsResponse["data"].([]interface{})
	if !ok {
		log.Fatalf("Invalid ads response format")
	}

	// Add ads to the priority queue
	for _, ad := range ads {
		adMap, ok := ad.(map[string]interface{})
		if !ok {
			continue
		}

		adv, ok := adMap["adv"].(map[string]interface{})
		if !ok {
			log.Println("adv not found or is not a map in adMap")
			continue
		}

		priceStr, ok := adv["price"].(string)
		if !ok || priceStr == "" {
			log.Println("price not found or is not a string in adMap")
			continue
		}

		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			log.Printf("Failed to parse price: %v", err)
			continue
		}

		taskName := fmt.Sprintf("Order %s", adv["advNo"].(string))

		// Check if the ad has been processed before and had an error
		if adInfo, exists := j.adsTracker[taskName]; exists {
			if adInfo.HasError {
				if adInfo.Err != nil && !errorIgnore[int(adInfo.Err["code"].(float64))] {
					continue
				}

				if j.Queue.ContainsTask(taskName) {
					continue
				}

				// log.Printf("Retrying task %s\n", taskName)
			} else {
				continue
			}
		}

		// fmt.Printf("Not Skipping ad %s \n", taskName)
		// Store the ad information
		j.adsTracker[taskName] = &AdInfo{
			Ad:       adv,
			HasError: false,
		}

		// we do not want to add the same task multiple times and also need to select some perticular type of ads.
		j.Queue.AddTask(priority_queue.Task{
			Name: taskName,
			PriorityFn: func(price float64) func() int {
				return func() int { return calculatePriority(price) }
			}(price),
			Timestamp: time.Now(),
			Exec: func(taskName string, adv map[string]interface{}) func() {
				return func() {
					err := j.createOrder(taskName, adv)
					if err != nil {
						// log.Printf("Error creating order for %s: %v\n", taskName, err)
						j.adsTracker[taskName].HasError = true
					} else {
						// log.Printf("Successfully executed task %s\n", taskName)
					}
				}
			}(taskName, adv),
		})
	}
}

func calculatePriority(price float64) int {
	basePriority := int(100.0 - price)
	return basePriority
}

func (j *Job) createOrder(taskName string, adv map[string]interface{}) error {
	// Ensure adv contains the necessary fields
	if adv == nil {
		log.Println("adv is nil")
		return fmt.Errorf("adv is nil")
	}

	advOrderNumber, ok := adv["advNo"].(string)
	if !ok {
		return fmt.Errorf("advOrderNumber not found or is not a string in adv")
	}

	asset, ok := adv["asset"].(string)
	if !ok {
		return fmt.Errorf("asset not found or is not a string in adv")
	}

	fiatUnit, ok := adv["fiatUnit"].(string)
	if !ok {
		return fmt.Errorf("fiatUnit not found or is not a string in adv")
	}

	tradeType, ok := adv["tradeType"].(string)
	if !ok {
		return fmt.Errorf("tradeType not found or is not a string in adv")
	}

	buyType := "BY_MONEY"

	matchPriceStr, ok := adv["price"].(string)
	if !ok {
		return fmt.Errorf("price not found or is not a string in adv")
	}
	matchPrice, err := strconv.ParseFloat(matchPriceStr, 64)
	if err != nil {
		return fmt.Errorf("failed to parse price: %v", err)
	}

	surplusAmountStr, ok := adv["surplusAmount"].(string)
	if !ok {
		return fmt.Errorf("surplusAmount not found or is not a string in adv")
	}
	surplusAmount, err := strconv.ParseFloat(surplusAmountStr, 64)
	if err != nil {
		log.Printf("Failed to parse surplusAmount: %v", err)
		j.adsTracker[taskName].HasError = true
		return fmt.Errorf("failed to parse surplusAmount: %v", err)
	}

	minSingleTransAmountStr, ok := adv["minSingleTransAmount"].(string)
	if !ok {
		log.Println("minSingleTransAmount not found or is not a string in adv")
		j.adsTracker[taskName].HasError = true
		return fmt.Errorf("minSingleTransAmount not found or is not a string in adv")
	}
	minSingleTransAmount, err := strconv.ParseFloat(minSingleTransAmountStr, 64)
	if err != nil {
		log.Printf("Failed to parse minSingleTransAmount: %v", err)
		j.adsTracker[taskName].HasError = true
		return fmt.Errorf("failed to parse minSingleTransAmount: %v", err)
	}

	maxSingleTransAmountStr, ok := adv["maxSingleTransAmount"].(string)
	if !ok {
		log.Println("maxSingleTransAmount not found or is not a string in adv")
		j.adsTracker[taskName].HasError = true
		return fmt.Errorf("maxSingleTransAmount not found or is not a string in adv")
	}
	maxSingleTransAmount, err := strconv.ParseFloat(maxSingleTransAmountStr, 64)
	if err != nil {
		log.Printf("Failed to parse maxSingleTransAmount: %v", err)
		j.adsTracker[taskName].HasError = true
		return fmt.Errorf("failed to parse maxSingleTransAmount: %v", err)
	}

	totalAmount := getOrderAmount(matchPrice, maxSingleTransAmount, minSingleTransAmount, surplusAmount)

	// Simulate an error for demonstration purposes
	orderResponse, err := j.BinanceAPI.PlaceOrder(advOrderNumber, asset, buyType, fiatUnit, tradeType, matchPrice, totalAmount)
	if err != nil {
		// log.Printf("Failed to create order for %s: %v", advOrderNumber, err)
		j.adsTracker[taskName].HasError = true
		return err
	}

	orderMessage := fmt.Sprintf(
		"📄 Order Number: %s\n💰 Match Price: %.2f\n📦 Surplus Amount: %.2f\n🔢 Transaction Limits: %.2f - %.2f\n💴 Total Amount: %.2f\n",
		advOrderNumber, matchPrice, surplusAmount, minSingleTransAmount, maxSingleTransAmount, totalAmount,
	)

	// Check if "success" key exists and is a boolean
	if success, ok := orderResponse["success"].(bool); !ok || !success {
		// log.Printf("Order creation failed for %s", orderResponse)
		j.adsTracker[taskName].HasError = true
		j.adsTracker[taskName].Err = orderResponse

		errorMessage, _ := orderResponse["msg"].(string)
		if errorMessage == "" {
			errorMessage = "Unknown error occurred."
		}

		errorCode := orderResponse["code"]
		message := fmt.Sprintf(
			"🛑 Order Fail 🛑 \n\n%s\nERR CODE: %v\nERR MSG: %s",
			orderMessage, errorCode, errorMessage,
		)

		j.bot.Send(tgbotapi.NewMessage(j.chatID,  message))
		return fmt.Errorf("order creation failed for %s", advOrderNumber)
	}

	return nil
}

func getOrderAmount(matchPrice, maxSingleTransAmount, minSingleTransAmount, totalSurplusAmount float64) float64 {
	minSurplusAmount := minSingleTransAmount / matchPrice
	maxSurplusAmount := maxSingleTransAmount / matchPrice

	// Ensure the surplus amount stays within limits
	surplusAmount := max(minSurplusAmount, min(maxSurplusAmount, totalSurplusAmount))

	// Calculate the max possible amount
	maxPossibleAmount := surplusAmount * matchPrice
	return maxPossibleAmount
}

// Helper functions to find min and max for float64
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
