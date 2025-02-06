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

	j.wg.Add(1)
	go func() {
		defer j.wg.Done()
		for {
			select {
			case <-j.stopChan:
				return
			default:
				j.Queue.ProcessTasks()
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	for {
		select {
		case <-j.stopChan:
			return
		default:
			asset := getConfigValue(j.BinanceAPI.Config, "asset", "USDT")
			fiat := getConfigValue(j.BinanceAPI.Config, "fiat", "INR")
			page := getConfigIntValue(j.BinanceAPI.Config, "page", 1)
			rows := getConfigIntValue(j.BinanceAPI.Config, "rows", 2)
			tradeType := getConfigValue(j.BinanceAPI.Config, "trade_type", "BUY")
			j.ListAdsAndCreateOrders(asset, fiat, page, rows, tradeType)
			j.lastRunTime = time.Now()
			j.totalRuns++
			time.Sleep(time.Duration(config.CallJobInterval) * time.Second)
		}
	}
}

func (j *Job) Stop() {
	select {
	case <-j.stopChan:
		fmt.Println("Job is already closed")
	default:
		close(j.stopChan)
	}

	j.Queue.Clear()
	fmt.Println("Stopping the job")
}

func (j *Job) ListAdsAndCreateOrders(asset, fiat string, page, rows int, tradeType string) {
	if j.BinanceAPI.Config == nil {
		log.Printf("Binance API config not found")
		return
	}

	// Extract error codes to ignore from the config
	extraFilter, ok := j.BinanceAPI.Config["extra_filter"].(map[string]interface{})
	if !ok {
		log.Printf("extra_filter not found or is not a map in config")
		return
	}

	errorCodesStr, ok := extraFilter["error_codes"].(string)
	if !ok {
		log.Printf("error_codes not found or is not a string in extra_filter")
		return
	}

	errorCodesList := strings.Split(errorCodesStr, ",")

	errorIgnore := make(map[int]bool)
	for _, codeStr := range errorCodesList {
		code, err := strconv.Atoi(strings.TrimSpace(codeStr))
		if err != nil {
			log.Printf("Failed to parse error code: %v", err)
			continue
		}
		errorIgnore[code] = true
	}

	// Fetch ads from Binance API
	adsResponse, err := j.BinanceAPI.SearchAds(asset, fiat, page, rows, tradeType)
	if err != nil {
		log.Printf("Failed to fetch ads: %v", err)
		return
	}

	// Extract ads from response
	ads, ok := adsResponse["data"].([]interface{})
	if !ok {
		log.Printf("Invalid ads response format")
		message := fmt.Sprintf(" 🛑 List Ads Fails 🛑 \n\n Error: %s \n Use command  /stop to the job", "Invalid ads response format")
		j.bot.Send(tgbotapi.NewMessage(j.chatID, message))
		return
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

		maxSingleTransAmountStr, ok := adv["maxSingleTransAmount"].(string)
		if !ok {
			log.Println("minSingleTransAmount not found or is not a string in adv")
			continue
		}
		maxSingleTransAmount, err := strconv.ParseFloat(maxSingleTransAmountStr, 64)
		if err != nil {
			log.Printf("Failed to parse minSingleTransAmount: %v", err)
			continue
		}

		extraFilter, ok := j.BinanceAPI.Config["extra_filter"].(map[string]interface{})
		if !ok {
			log.Printf("extra_filter not found or is not a map in config")
			return
		}

		targetPrice := getConfigFloatValue(extraFilter, "price", 0)
		if !ok {
			log.Printf("extra_filter not found or is not a map in config")
			return
		}

		minimumLimit := getConfigFloatValue(extraFilter, "minimum_limit", 0)
		if !ok {
			log.Printf("extra_filter not found or is not a map in config")
			return
		}

		taskName := fmt.Sprintf("Order %s", adv["advNo"].(string))

		// fmt.Printf("Price: %.2f, Target Price: %.2f\n minSingleTransAmount: %.2f \n minimumLimit %.2f ", price, targetPrice, minSingleTransAmount, minimumLimit)

		if price > targetPrice {
			continue
		}

		if maxSingleTransAmount <= minimumLimit {
			continue
		}

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
