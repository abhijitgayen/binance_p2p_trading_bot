package msg_gen

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

func GenerateConfigMessage(botConfig map[string]interface{}) string {
	if len(botConfig) == 0 {
		return "No configuration found for this user."
	}

	// Extract and sort keys
	keys := make([]string, 0, len(botConfig))
	for key := range botConfig {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("📊 *Bot Configuration Details* 📊\n\n")

	for _, key := range keys {
		value := botConfig[key]
		switch v := value.(type) {
		case string:
			if key == "api_key" || key == "secret_key" {
				sb.WriteString(fmt.Sprintf("  🔑 *%s*: `%s`\n", key, maskKey(v)))
			} else {
				sb.WriteString(fmt.Sprintf("  🔹 *%s*: `%s`\n", key, v))
			}
		case map[string]interface{}:
			sb.WriteString(fmt.Sprintf("  🛠 *%s*:\n", key))
			for subKey, subValue := range v {
				sb.WriteString(fmt.Sprintf("      🔹 *%s*: `%v`\n", subKey, subValue))
			}
		default:
			sb.WriteString(fmt.Sprintf("  🔹 *%s*: `%v`\n", key, v))
		}
	}

	return sb.String()
}

// maskKey masks the API key, showing only the first and last 4 characters.
func maskKey(key string) string {
	if len(key) < 8 {
		return "Not Configured ❌"
	}
	return fmt.Sprintf("%s*****%s ✅", key[:4], key[len(key)-4:])
}

func safeGetString(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		return fmt.Sprintf("%v", value)
	}
	return "N/A"
}

func safeGetFloat64(data map[string]interface{}, key string) float64 {
	if value, ok := data[key].(float64); ok {
		return value
	}
	return 0.0
}

func GenerateOrderMessage(responsePlaceOrder map[string]interface{}) string {

	log.Printf("responsePlaceOrder: %v", responsePlaceOrder)

	if success, ok := responsePlaceOrder["success"].(bool); !ok || !success {
		return "Error: Order response is not successful or missing required data."
	}

	orderMatch, _ := responsePlaceOrder["data"].(map[string]interface{})["orderMatch"].(map[string]interface{})

	var builder strings.Builder

	builder.WriteString("*📝 Order Information:*\n\n")
	builder.WriteString(fmt.Sprintf("📋 *Order Number:* `%v`\n", safeGetString(orderMatch, "orderNumber")))
	builder.WriteString(fmt.Sprintf("📋 *Adv Order Number:* `%v`\n\n", safeGetString(orderMatch, "advOrderNumber")))

	builder.WriteString(fmt.Sprintf("🧑‍💻 *User Id:* `%v`\n", safeGetFloat64(orderMatch, "userId")))
	builder.WriteString(fmt.Sprintf("👤 *Adv User Id:* `%v`\n\n", safeGetFloat64(orderMatch, "advUserId")))

	builder.WriteString("🛍️ *Buyer Information:*\n")
	builder.WriteString(fmt.Sprintf("- *Nickname:* `%v`\n", safeGetString(orderMatch, "buyerNickname")))
	builder.WriteString(fmt.Sprintf("- *Name:* `%v`\n\n", safeGetString(orderMatch, "buyerName")))

	builder.WriteString("🛍️ *Seller Information:*\n")
	builder.WriteString(fmt.Sprintf("- *Nickname:* `%v`\n", safeGetString(orderMatch, "sellerNickname")))
	builder.WriteString(fmt.Sprintf("- *Name:* `%v`\n\n", safeGetString(orderMatch, "sellerName")))

	builder.WriteString("💰 *Transaction Details:*\n")
	builder.WriteString(fmt.Sprintf("- *Amount:* `%v %v`\n", safeGetString(orderMatch, "amount"), safeGetString(orderMatch, "asset")))
	builder.WriteString(fmt.Sprintf("- *Price:* `%v %v/%v`\n", safeGetString(orderMatch, "price"), safeGetString(orderMatch, "fiatUnit"), safeGetString(orderMatch, "asset")))
	builder.WriteString(fmt.Sprintf("- *Total Price:* `%v %v`\n\n", safeGetString(orderMatch, "totalPrice"), safeGetString(orderMatch, "fiatUnit")))

	builder.WriteString("💼 *Trade Information:*\n")
	builder.WriteString(fmt.Sprintf("- *Trade Type:* `%v`\n", safeGetString(orderMatch, "tradeType")))
	builder.WriteString(fmt.Sprintf("- *Pay Type:* `%v`\n", safeGetString(orderMatch, "payType")))

	return builder.String()
}

func GetTypedValue(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		switch v := value.(type) {
		case string:
			return v
		case int, int8, int16, int32, int64:
			return fmt.Sprintf("%d", v)
		case uint, uint8, uint16, uint32, uint64:
			return fmt.Sprintf("%d", v)
		case float32, float64:
			return fmt.Sprintf("%f", v)
		case bool:
			return fmt.Sprintf("%t", v)
		case []interface{}:
			return fmt.Sprintf("Array: %v", v)
		case map[string]interface{}:
			return fmt.Sprintf("Map: %v", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return "N/A"
}
