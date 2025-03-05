package jobs

import (
	"fmt"
	"go_binance_bot/src/apis"
	"go_binance_bot/src/config"
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
	BinanceAPI          *apis.BinanceAPI
	Queue               *priority_queue.PriorityQueue
	stopChan            chan struct{}
	adsTracker          map[string]*AdInfo
	wg                  sync.WaitGroup
	bot                 *tgbotapi.BotAPI
	chatID              int64
	lastRunTime         time.Time
	totalRuns           int
	TotalAmountToInvest float64
	NoOfOrders          int
	searchAdsInProgress int32
	mu                  sync.Mutex
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
	j.TotalAmountToInvest = getConfigFloatValue(j.BinanceAPI.Config, "total_amount_to_invest", 0)
	j.NoOfOrders = getConfigIntValue(j.BinanceAPI.Config, "no_of_orders", 1)

	j.wg.Add(2) // Add two jobs to WaitGroup

	go j.processQueueWorker()
	go j.createOrdersWorker()

	j.wg.Wait() // Wait for both workers to complete
}

func (j *Job) Stop() {
	select {
	case <-j.stopChan:
		log.Printf("Job is already closed")
	default:
		close(j.stopChan)
	}

	j.Queue.Clear()
	log.Printf("Stopping the job")
}

// Worker to process the queue
func (j *Job) processQueueWorker() {
	defer j.wg.Done()

	ticker := time.NewTicker(time.Duration(config.ProcessQueueInterval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-j.stopChan:
			return
		case <-ticker.C:
			j.Queue.ProcessTasks()
		}
	}
}

func (j *Job) createOrdersWorker() {
	go func() {
		for {
			// Check for a stop signal.
			select {
			case <-j.stopChan:
				return
			default:
			}

			// Retrieve config values.
			asset := getConfigValue(j.BinanceAPI.Config, "asset", "USDT")
			fiat := getConfigValue(j.BinanceAPI.Config, "fiat", "INR")
			page := getConfigIntValue(j.BinanceAPI.Config, "page", 1)
			rows := getConfigIntValue(j.BinanceAPI.Config, "rows", 2)
			tradeType := getConfigValue(j.BinanceAPI.Config, "trade_type", "BUY")

			// Call the API.
			adsResponse, err := j.BinanceAPI.SearchAds(asset, fiat, page, rows, tradeType)
			if err != nil {
				log.Printf("Error fetching ads: %v", err)
				// On error, wait for the configured interval.
				time.Sleep(time.Duration(config.CallJobInterval) * time.Microsecond)
			} else {
				// Process the ads concurrently without blocking the next API call.
				go j.processAdsResponse(adsResponse)
				// On success, wait only x micoseconds before next call.
				time.Sleep(time.Duration(config.CallJobInterval) * time.Microsecond)
			}

			// Update run-time stats.
			j.lastRunTime = time.Now()
			j.totalRuns++
		}
	}()
}

// processAdsResponse processes the API response by validating each ad
// and adding it to the queue if it meets the criteria.
func (j *Job) processAdsResponse(adsResponse map[string]interface{}) {
	// Extract ads from the response.
	ads, ok := adsResponse["data"].([]interface{})
	if !ok {
		j.handleErrorResponse(adsResponse)
		return
	}

	// Set a concurrency limit. Adjust this value based on your system's capacity.
	concurrencyLimit := 10
	sem := make(chan struct{}, concurrencyLimit)
	var wg sync.WaitGroup

	// Process each ad concurrently, but only up to the concurrency limit.
	for _, ad := range ads {
		sem <- struct{}{} // Acquire a slot.
		wg.Add(1)

		go func(ad interface{}) {
			defer wg.Done()
			defer func() { <-sem }() // Release the slot when done.

			if err := j.processAd(ad); err != nil {
				log.Printf("Error processing ad: %v", err)
			}
		}(ad)
	}

	wg.Wait() // Wait for all Goroutines to finish.
}

func (j *Job) processAd(ad interface{}) error {
	// Extract the ad map.
	adMap, ok := ad.(map[string]interface{})
	if !ok {
		return fmt.Errorf("ad not found or is not a map in ads")
	}

	adv, ok := adMap["adv"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("adv not found or is not a map in adMap")
	}

	// Validate and convert the price.
	price, err := j.parseFloat(adv, "price")
	if err != nil {
		return err
	}

	// Validate and convert maxSingleTransAmount.
	maxSingleTransAmount, err := j.parseFloat(adv, "maxSingleTransAmount")
	if err != nil {
		return err
	}

	// Retrieve extra filter settings.
	extraFilter, ok := j.BinanceAPI.Config["extra_filter"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("extra_filter not found or is not a map in config")
	}
	targetPrice := getConfigFloatValue(extraFilter, "price", 0)
	minimumLimit := getConfigFloatValue(extraFilter, "minimum_limit", 0)

	// Build a unique task name (assuming advNo is unique).
	taskName := fmt.Sprintf("Order %s", adv["advNo"].(string))

	// Validate based on your criteria.
	if price > targetPrice || maxSingleTransAmount <= minimumLimit {
		return nil
	}

	// Safely update shared state.
	j.mu.Lock()
	if adInfo, exists := j.adsTracker[taskName]; exists {
		if !adInfo.HasError || (adInfo.Err != nil && !shouldRetry(adInfo.Err, extraFilter)) {
			j.mu.Unlock()
			return nil
		}
	}
	j.adsTracker[taskName] = &AdInfo{Ad: adv, HasError: false}
	j.mu.Unlock()

	// Add the validated ad as a task to your priority queue.
	j.Queue.AddTask(priority_queue.Task{
		Name: taskName,
		PriorityFn: func() int {
			return calculatePriority(price)
		},
		Timestamp: time.Now(),
		Exec: func() {
			if err := j.createOrder(taskName, adv); err != nil {
				log.Printf("Error creating order for %s: %v", taskName, err)
			} else {
				log.Printf("Successfully executed task %s", taskName)
			}
		},
	})

	return nil
}

func (j *Job) parseFloat(data map[string]interface{}, key string) (float64, error) {
	valueStr, ok := data[key].(string)
	if !ok || valueStr == "" {
		return 0, fmt.Errorf("%s not found or is not a string in data", key)
	}
	return strconv.ParseFloat(valueStr, 64)
}

func (j *Job) handleErrorResponse(response map[string]interface{}) {
	errorMessage, ok := response["msg"].(string)
	if !ok || errorMessage == "" {
		errorMessage = "Unknown error occurred."
	}
	errorCode := response["code"]
	message := fmt.Sprintf(" 🛑 List Ads Fails 🛑 \n\nERR CODE: %v\nERR MSG: %s  \n\n Use command /stop to the job",
		errorCode, errorMessage)
	j.bot.Send(tgbotapi.NewMessage(j.chatID, message))
}

// Example helper to decide if we should retry based on error details and extraFilter settings.
// shouldRetry determines if a task should be retried based on error data and extra filter criteria.
// It returns true if the error code found in errData is one of the error codes specified in extraFilter["error_codes"].
func shouldRetry(errData map[string]interface{}, extraFilter map[string]interface{}) bool {
	// Retrieve error codes string from extraFilter.
	errorCodesStr, ok := extraFilter["error_codes"].(string)
	if !ok {
		// If there are no specified error codes, allow retry by default.
		return true
	}

	// Parse the comma-separated error codes into a map.
	errorCodesList := strings.Split(errorCodesStr, ",")
	allowedErrorCodes := make(map[int]bool)
	for _, codeStr := range errorCodesList {
		trimmed := strings.TrimSpace(codeStr)
		if code, err := strconv.Atoi(trimmed); err == nil {
			allowedErrorCodes[code] = true
		}
	}

	// Extract the error code from errData.
	codeInterface, ok := errData["code"]
	if !ok {
		// If there is no error code in the error data, allow retry by default.
		return true
	}

	var code int
	switch v := codeInterface.(type) {
	case float64:
		code = int(v)
	case int:
		code = v
	case string:
		if parsed, err := strconv.Atoi(v); err == nil {
			code = parsed
		} else {
			// Cannot parse the string to an integer; do not retry.
			return false
		}
	default:
		// Unknown type; do not retry.
		return false
	}

	// Allow retry only if the error code is one of the allowed error codes.
	return allowedErrorCodes[code]
}
