package apis

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type BinanceAPI struct {
	BaseURL   string
	APIKey    string
	SecretKey string
	Config    map[string]interface{}
	Client    *http.Client
	mutex     sync.Mutex
}

func NewBinanceAPI(baseURL, apiKey, secretKey string, config map[string]interface{}) *BinanceAPI {
	return &BinanceAPI{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		SecretKey: secretKey,
		Config:    config,
		Client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:       100, // Increase concurrent connections
				MaxConnsPerHost:    100, // Prevents bottlenecks
				IdleConnTimeout:    60 * time.Second,
				DisableCompression: true,
				ForceAttemptHTTP2:  true, // Uses HTTP/2 for multiplexing
			},
			Timeout: 10 * time.Second, // Set aggressive timeout
		},
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
	url := fmt.Sprintf("%s%s?%s&signature=%s", b.BaseURL, endpoint, query, b.generateSignature(query))

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal body: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-MBX-APIKEY", b.APIKey)
	req.Header.Set("Content-Type", "application/json")

	b.mutex.Lock()
	resp, err := b.Client.Do(req)
	b.mutex.Unlock()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	return response, nil
}

// Fetch Ads
func (b *BinanceAPI) SearchAds(asset, fiat string, page, rows int, tradeType string) (map[string]interface{}, error) {
	query := fmt.Sprintf("asset=%s&fiat=%s&page=%d&rows=%d&tradeType=%s&timestamp=%d",
		asset, fiat, page, rows, tradeType, time.Now().UnixMilli())

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
	query := fmt.Sprintf("advOrderNumber=%s&asset=%s&buyType=%s&fiatUnit=%s&timestamp=%d",
		advOrderNumber, asset, buyType, fiatUnit, time.Now().UnixMilli())

	body := map[string]interface{}{
		"advOrderNumber": advOrderNumber,
		"asset":          asset,
		"buyType":        buyType,
		"fiatUnit":       fiatUnit,
		"matchPrice":     matchPrice,
		"totalAmount":    totalAmount,
		"tradeType":      tradeType,
		"origin":         "MAKE_TAKE",
	}

	return b.sendRequest("/sapi/v1/c2c/orderMatch/placeOrder", query, body)
}
