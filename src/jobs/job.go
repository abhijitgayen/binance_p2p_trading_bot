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
	"sync/atomic"
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
	ticker := time.NewTicker(time.Duration(config.CallJobInterval) * time.Microsecond)
	defer ticker.Stop()

	for {
		select {
		case <-j.stopChan:
			return
		case <-ticker.C:
			// Check if a SearchAds call is already in progress.
			// If yes, skip this tick.
			if !atomic.CompareAndSwapInt32(&j.searchAdsInProgress, 0, 1) {
				// A call is already in flight, so we simply skip this tick.
				continue
			}

			asset := getConfigValue(j.BinanceAPI.Config, "asset", "USDT")
			fiat := getConfigValue(j.BinanceAPI.Config, "fiat", "INR")
			page := getConfigIntValue(j.BinanceAPI.Config, "page", 1)
			rows := getConfigIntValue(j.BinanceAPI.Config, "rows", 2)
			tradeType := getConfigValue(j.BinanceAPI.Config, "trade_type", "BUY")

			// Call SearchAds asynchronously so the worker doesn’t block.
			go func() {
				// Ensure we release the flag when done.
				defer atomic.StoreInt32(&j.searchAdsInProgress, 0)

				adsResponse, err := j.BinanceAPI.SearchAds(asset, fiat, page, rows, tradeType)
				if err != nil {
					log.Printf("Error fetching ads: %v", err)
					return
				}
				// If successful, process the ads response concurrently.
				// This function validates each ad and adds it to the queue.
				j.processAdsResponse(adsResponse)
			}()

			// Update run-time stats (these run regardless of the SearchAds result).
			j.lastRunTime = time.Now()
			j.totalRuns++
		}
	}
}

// processAdsResponse processes the API response by validating each ad
// and adding it to the queue if it meets the criteria.
func (j *Job) processAdsResponse(adsResponse map[string]interface{}) {
	// Extract ads from the response.
	ads, ok := adsResponse["data"].([]interface{})
	if !ok {
		log.Printf("Invalid ads response format")
		errorMessage, ok := adsResponse["msg"].(string)
		if !ok || errorMessage == "" {
			errorMessage = "Unknown error occurred."
		}

		errorCode := adsResponse["code"]

		message := fmt.Sprintf(" 🛑 List Ads Fails 🛑 \n\nERR CODE: %v\nERR MSG: %s  \n\n Use command /stop to the job",
			errorCode, errorMessage)
		j.bot.Send(tgbotapi.NewMessage(j.chatID, message))
		return
	}

	for _, ad := range ads {
		adMap, ok := ad.(map[string]interface{})
		if !ok {
			log.Println("ad not found or is not a map in ads")
			continue
		}

		adv, ok := adMap["adv"].(map[string]interface{})
		if !ok {
			log.Println("adv not found or is not a map in adMap")
			continue
		}

		// Validate and convert the price.
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

		// Validate and convert maxSingleTransAmount.
		maxSingleTransAmountStr, ok := adv["maxSingleTransAmount"].(string)
		if !ok {
			log.Println("maxSingleTransAmount not found or is not a string in adv")
			continue
		}
		maxSingleTransAmount, err := strconv.ParseFloat(maxSingleTransAmountStr, 64)
		if err != nil {
			log.Printf("Failed to parse maxSingleTransAmount: %v", err)
			continue
		}

		// Retrieve extra filter settings.
		extraFilter, ok := j.BinanceAPI.Config["extra_filter"].(map[string]interface{})
		if !ok {
			log.Printf("extra_filter not found or is not a map in config")
			continue
		}
		targetPrice := getConfigFloatValue(extraFilter, "price", 0)
		minimumLimit := getConfigFloatValue(extraFilter, "minimum_limit", 0)

		// Build a unique task name (assuming advNo is unique).
		taskName := fmt.Sprintf("Order %s", adv["advNo"].(string))

		// Validate based on your criteria.
		if price > targetPrice || maxSingleTransAmount <= minimumLimit {
			continue
		}

		// Protect shared state.
		if adInfo, exists := j.adsTracker[taskName]; exists {
			if !adInfo.HasError || (adInfo.Err != nil && !shouldRetry(adInfo.Err, extraFilter)) {
				continue
			}
		}
		j.adsTracker[taskName] = &AdInfo{Ad: adv, HasError: false}

		// Add the validated ad as a task to your priority queue.
		j.Queue.AddTask(priority_queue.Task{
			Name: taskName,
			PriorityFn: func() int {
				return calculatePriority(price)
			},
			Timestamp: time.Now(),
			Exec: func() {
				err := j.createOrder(taskName, adv)
				if err != nil {
					log.Printf("Error creating order for %s: %v", taskName, err)
				} else {
					log.Printf("Successfully executed task %s", taskName)
				}
			},
		})
	}
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
