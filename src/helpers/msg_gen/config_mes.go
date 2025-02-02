package msg_gen

import (
	"fmt"
	"strings"
	"sort"
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
