package main

import (
	"fmt"
	"net/http"
	"sort"
	"time"
)

// Binance API servers
var servers = []string{
	"https://api.binance.com",
	"https://api1.binance.com",
	"https://api2.binance.com",
	"https://api3.binance.com",
	"https://api4.binance.com",
	"https://api5.binance.com",
}

// Struct to hold server response time
type ServerResponse struct {
	Server string
	TimeMS int64
}

func checkResponseTime(server string) ServerResponse {
	url := server + "/sapi/v1/ping"
	start := time.Now()

	// Send GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error connecting to %s: %v\n", server, err)
		return ServerResponse{Server: server, TimeMS: 999999} // High value for failed requests
	}
	defer resp.Body.Close()

	// Calculate response time in milliseconds
	elapsed := time.Since(start).Milliseconds()
	return ServerResponse{Server: server, TimeMS: elapsed}
}

func main() {
	var results []ServerResponse

	// Check response time for all servers
	for _, server := range servers {
		fmt.Printf("Checking %s...\n", server)
		results = append(results, checkResponseTime(server))
	}

	// Sort by response time (ascending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].TimeMS < results[j].TimeMS
	})

	// Print results
	fmt.Println("\n🚀 Server Response Times:")
	for _, res := range results {
		fmt.Printf("%s: %d ms\n", res.Server, res.TimeMS)
	}

	// Print best server
	fmt.Printf("\n✅ Fastest server: %s (%d ms)\n", results[0].Server, results[0].TimeMS)
}
