package apis

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type BinanceAPI struct {
	BaseURL   string
	APIKey    string
	SecretKey string
	Config    map[string]interface{}
}

func NewBinanceAPI(baseURL, apiKey, secretKey string, config map[string]interface{}) *BinanceAPI {
	return &BinanceAPI{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		SecretKey: secretKey,
		Config:    config,
	}
}

// Generate HMAC SHA256 Signature
func (b *BinanceAPI) generateSignature(query string) string {
	h := hmac.New(sha256.New, []byte(b.SecretKey))
	h.Write([]byte(query))
	return hex.EncodeToString(h.Sum(nil))
}

// Send API Request
func (b *BinanceAPI) sendRequest(endpoint, query string, body map[string]interface{}) (map[string]interface{}, error) {
	signature := b.generateSignature(query)
	url := fmt.Sprintf("%s%s?%s&signature=%s", b.BaseURL, endpoint, query, signature)

	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-MBX-APIKEY", b.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	var response map[string]interface{}
	json.Unmarshal(bodyBytes, &response)

	return response, nil
}

// Fetch Ads
func (b *BinanceAPI) SearchAds(asset, fiat string, page, rows int, tradeType string) (map[string]interface{}, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	query := fmt.Sprintf("asset=%s&fiat=%s&page=%d&rows=%d&tradeType=%s&timestamp=%s",
		asset, fiat, page, rows, tradeType, timestamp)

	body := map[string]interface{}{
		"asset":     asset,
		"fiat":      fiat,
		"page":      page,
		"rows":      rows,
		"tradeType": tradeType,
	}

	return b.sendRequest("/sapi/v1/c2c/ads/search", query, body)
}

// Place Order
func (b *BinanceAPI) PlaceOrder(advOrderNumber, asset, buyType, fiatUnit, tradeType string, matchPrice, totalAmount float64) (map[string]interface{}, error) {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	query := fmt.Sprintf("advOrderNumber=%s&asset=%s&buyType=%s&fiatUnit=%s&timestamp=%s",
		advOrderNumber, asset, buyType, fiatUnit, timestamp)

	body := map[string]interface{}{
		"advOrderNumber": advOrderNumber,
		"asset":          asset,
		"buyType":        buyType,
		"fiatUnit":       fiatUnit,
		"matchPrice":     matchPrice,
		"totalAmount":    totalAmount,
		"tradeType":      tradeType,
		// "buyType":        "BY_MONEY",
		"origin": "MAKE_TAKE",
	}

	return b.sendRequest("/sapi/v1/c2c/orderMatch/placeOrder", query, body)
}
