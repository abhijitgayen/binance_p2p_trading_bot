package priority_queue

import (
	"log"
	"sync"
	"time"

	"github.com/google/btree"
)

// Task represents a function-based task with a dynamic priority
type Task struct {
	Exec       func()     // Function to execute
	PriorityFn func() int // Function to determine priority dynamically
	Name       string     // Task name (for debugging)
	Timestamp  time.Time  // Insertion time (tie-breaker)
}

// Less function to define priority condition
func (t Task) Less(other btree.Item) bool {
	otherTask := other.(Task)

	// First, compare priorities
	if t.PriorityFn() > otherTask.PriorityFn() {
		return true
	} else if t.PriorityFn() < otherTask.PriorityFn() {
		return false
	}

	// If priority is the same, use timestamp as a tie-breaker
	return t.Timestamp.Before(otherTask.Timestamp)
}

// PriorityQueue struct using btree
type PriorityQueue struct {
	tree      *btree.BTree
	mu        sync.Mutex    // Ensure thread safety
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
	var tasksToRemove []btree.Item

	// Collect tasks first
	pq.tree.Ascend(func(i btree.Item) bool {
		tasksToRemove = append(tasksToRemove, i)
		return true
	})

	pq.mu.Unlock() // Unlock before executing tasks

	// Execute tasks outside the lock
	for _, task := range tasksToRemove {
		task := task.(Task)
		log.Printf("Running: %s with priority %d\n", task.Name, task.PriorityFn())
		task.Exec()
		time.Sleep(pq.sleepTime)

		// Remove task after execution
		pq.mu.Lock()
		pq.tree.Delete(task)
		pq.mu.Unlock()
	}
}

// Clear clears the priority queue and stops processing tasks
func (pq *PriorityQueue) Clear() {
	pq.mu.Lock()
	pq.tree.Clear(true)
	pq.mu.Unlock()

	log.Printf("Clearing the priority queue")
}

func (pq *PriorityQueue) ContainsTask(taskName string) bool {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	found := false
	pq.tree.Ascend(func(i btree.Item) bool {
		task := i.(Task)
		if task.Name == taskName {
			found = true
			return false // Stop iteration
		}
		return true
	})
	return found
}
