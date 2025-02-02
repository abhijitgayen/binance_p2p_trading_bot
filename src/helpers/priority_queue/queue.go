package priority_queue

import (
    "fmt"
    "github.com/google/btree"
    "sync"
    "time"
)

// Task represents a function-based task with a dynamic priority
type Task struct {
    Exec       func()  // Function to execute
    PriorityFn func() int // Function to determine priority dynamically
    Name       string  // Task name (for debugging)
}

// Less function to define priority comparison
func (t Task) Less(other btree.Item) bool {
    return t.PriorityFn() > other.(Task).PriorityFn()
}

// PriorityQueue struct using btree
type PriorityQueue struct {
    tree      *btree.BTree
    mu        sync.Mutex // Ensure thread safety
    sleepTime time.Duration // Sleep time between task executions
}

// NewPriorityQueue initializes a new priority queue with a specified sleep time
func NewPriorityQueue(degree int, sleepTime time.Duration) *PriorityQueue {
    return &PriorityQueue{
        tree:      btree.New(degree),
        sleepTime: sleepTime,
    }
}

// AddTask adds a new task to the priority queue
func (pq *PriorityQueue) AddTask(task Task) {
    pq.mu.Lock()
    defer pq.mu.Unlock()
    pq.tree.ReplaceOrInsert(task)
}

// ProcessTasks processes all tasks in order of priority
func (pq *PriorityQueue) ProcessTasks() {
    pq.mu.Lock()
    defer pq.mu.Unlock()

    pq.tree.Ascend(func(i btree.Item) bool {
        task := i.(Task)
        fmt.Printf("Running: %s with priority %d\n", task.Name, task.PriorityFn())
        task.Exec()
        time.Sleep(pq.sleepTime) // Use the defined sleep time
        return true
    })
}