package jobs

import (
	"fmt"
	"go_binance_bot/src/config"
	"go_binance_bot/src/helpers/msg_gen"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

	tradeType := "BUY"
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

	if (j.TotalAmountToInvest - totalAmount) < 0 {
		if j.TotalAmountToInvest > minSingleTransAmount {
			totalAmount = j.TotalAmountToInvest
		} else {
			log.Printf("Inappropriate Amount:  TotalAmountToInvest -> %v totalAmount -> %v", j.TotalAmountToInvest, totalAmount)
			orderMessage := fmt.Sprintf(
				"📄 Order Number: %s\n💰 Match Price: %.2f\n📦 Surplus Amount: %.2f\n🔢 Transaction Limits: %.2f - %.2f\n💴 Total Amount: %.2f\n",
				advOrderNumber, matchPrice, surplusAmount, minSingleTransAmount, maxSingleTransAmount, totalAmount,
			)
			message := fmt.Sprintf(" 🛑 Inappropriate Amount 🛑 \n\n %s \n Use command  /stop to the job", orderMessage)
			j.bot.Send(tgbotapi.NewMessage(j.chatID, message))
			j.adsTracker[taskName].HasError = true
			return nil
		}
	}

	orderMessage := fmt.Sprintf(
		"📄 Order Number: %s\n💰 Match Price: %.2f\n📦 Surplus Amount: %.2f\n🔢 Transaction Limits: %.2f - %.2f\n💴 Total Amount: %.2f\n",
		advOrderNumber, matchPrice, surplusAmount, minSingleTransAmount, maxSingleTransAmount, totalAmount,
	)

	// Simulate an error for demonstration purposes
	orderResponse, err := j.BinanceAPI.PlaceOrder(advOrderNumber, asset, buyType, fiatUnit, tradeType, matchPrice, totalAmount)
	if err != nil {
		log.Printf("Failed to create order for %s: %v", advOrderNumber, err)
		j.adsTracker[taskName].HasError = true
		return err
	}

	// Check if "success" key exists and is a boolean
	if success, ok := orderResponse["success"].(bool); !ok || !success {
		// log.Printf("Order creation failed for %s", orderResponse)
		j.adsTracker[taskName].HasError = true
		j.adsTracker[taskName].Err = orderResponse

		errorMessage, ok := orderResponse["msg"].(string)
		if !ok || errorMessage == "" {
			errorMessage = "Unknown error occurred."
		}

		errorCode := orderResponse["code"]
		message := fmt.Sprintf(
			"🛑 Order Fail 🛑 \n\n%s\nERR CODE: %v\nERR MSG: %s",
			orderMessage, errorCode, errorMessage,
		)

		j.bot.Send(tgbotapi.NewMessage(j.chatID, message))
		adminMessage := fmt.Sprintf("User ID: %d\n%s", j.chatID, message)
		j.bot.Send(tgbotapi.NewMessage(config.NotifyUserId, adminMessage))
		return nil
	}

	j.TotalAmountToInvest -= totalAmount
	j.NoOfOrders--

	orderDetailsMsg := msg_gen.GenerateOrderMessage(orderResponse)
	message := fmt.Sprintf("🎉 Order Success 🎉 \n\n%s \n\n%s", orderMessage, orderDetailsMsg)
	j.bot.Send(tgbotapi.NewMessage(j.chatID, message))
	adminMessage := fmt.Sprintf("User ID: %d\n%s", j.chatID, message)
	j.bot.Send(tgbotapi.NewMessage(config.NotifyUserId, adminMessage))

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

func getConfigValue(config map[string]interface{}, key, defaultValue string) string {
	if value, ok := config[key].(string); ok {
		return value
	}
	return defaultValue
}

func getConfigIntValue(config map[string]interface{}, key string, defaultValue int) int {
	if value, ok := config[key].(float64); ok {
		return int(value)
	}
	return defaultValue
}

func getConfigFloatValue(config map[string]interface{}, key string, defaultValue float64) float64 {
	if value, ok := config[key].(float64); ok {
		return float64(value)
	}
	return defaultValue
}
