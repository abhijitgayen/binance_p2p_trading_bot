package utils

import (
	"encoding/json"
	"fmt"
)

// GetTypedValue retrieves a value from a map and returns it as a string.
func GetTypedValue(data map[string]interface{}, key string) string {
	value, exists := data[key]
	if !exists || value == nil {
		return "N/A"
	}

	// Use type assertion for efficiency
	switch v := value.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32:
		return fmt.Sprintf("%.6f", v) // Limit decimal places
	case float64:
		return fmt.Sprintf("%.6f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case []interface{}, map[string]interface{}:
		jsonStr, _ := json.Marshal(v)
		return string(jsonStr) // Convert JSON efficiently
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetInt retrieves an integer safely.
func GetInt(data map[string]interface{}, key string) int64 {
	switch v := data[key].(type) {
	case float64: // JSON stores numbers as float64
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	}
	return 0
}

// GetFloat retrieves a float64 safely.
func GetFloat(data map[string]interface{}, key string) float64 {
	if v, ok := data[key].(float64); ok {
		return v
	}
	return 0.0
}

// GetString retrieves a string safely.
func GetString(data map[string]interface{}, key string) string {
	if v, ok := data[key].(string); ok {
		return v
	}
	return ""
}

// GetBool retrieves a boolean safely.
func GetBool(data map[string]interface{}, key string) bool {
	if v, ok := data[key].(bool); ok {
		return v
	}
	return false
}

// ToJSON converts data to JSON string safely.
func ToJSON(data interface{}) string {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(jsonStr)
}
