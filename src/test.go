package main

import (
    "fmt"
    "time"
    "go_binance_bot/src/helpers/priority_queue" // Correct import path
)

// Priority function that takes the price and creation time as input
func calculatePriority(price float64, creationTime time.Time) int {
    // Lower price gets higher base priority
    basePriority := int(100.0 - price)
    
    // More recent creation time adds to the priority
    timePriority := int(time.Since(creationTime).Minutes())
    return basePriority - timePriority
}

func main() {
    // Create a new priority queue with a sleep time of 1 second
    pq := priority_queue.NewPriorityQueue(2, 1*time.Second) // B-Tree of degree 2

    // Example ads with prices and creation times
    ads := []struct {
        price        float64
        creationTime time.Time
    }{
        {85.1, time.Now().Add(-10 * time.Minute)},
        {85, time.Now().Add(-5 * time.Minute)},
        {96.9, time.Now().Add(-2 * time.Minute)},
        {50, time.Now().Add(-20 * time.Minute)},
    }

    // Add tasks with dynamic priority functions based on price and creation time
    for i, ad := range ads {
        taskName := fmt.Sprintf("Order %c", 'A'+i)
        pq.AddTask(priority_queue.Task{
            Name: taskName,
            PriorityFn: func(price float64, creationTime time.Time) func() int {
                return func() int { return calculatePriority(price, creationTime) }
            }(ad.price, ad.creationTime),
            Exec: func(price float64, creationTime time.Time) func() {
                return func() { fmt.Printf("Executing %s with price %.2f and creation time %s\n", taskName, price, creationTime) }
            }(ad.price, ad.creationTime),
        })
    }

    // Process tasks in priority order
    pq.ProcessTasks()
}
