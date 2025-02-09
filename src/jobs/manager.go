package jobs

import (
	"fmt"
	"go_binance_bot/src/apis"
	"go_binance_bot/src/helpers/priority_queue"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type JobManager struct {
	jobs map[int64]*Job
	mu   sync.Mutex
}

var manager *JobManager
var once sync.Once

func GetJobManager() *JobManager {
	once.Do(func() {
		manager = &JobManager{
			jobs: make(map[int64]*Job),
		}
	})
	return manager
}

func (jm *JobManager) StartJob(userID int64, api *apis.BinanceAPI, queue *priority_queue.PriorityQueue, bot *tgbotapi.BotAPI, chatID int64) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if _, exists := jm.jobs[userID]; !exists {
		job := NewJob(api, queue, bot, chatID)
		jm.jobs[userID] = job
		go job.Run() // Run the job in a separate goroutine
	}
}

func (jm *JobManager) StopJob(userID int64) error {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if job, exists := jm.jobs[userID]; exists {
		job.Stop()
		delete(jm.jobs, userID)
		return nil
	}
	return fmt.Errorf("no job exists for user %d", userID)
}

func (jm *JobManager) GetJobStatus(userID int64) string {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job, exists := jm.jobs[userID]
	if !exists {
		return "No job is currently running."
	}

	lastRunDuration := time.Since(job.lastRunTime)
	var lastRunTimeStr string
	if lastRunDuration < time.Millisecond {
		lastRunTimeStr = fmt.Sprintf("%d microseconds ago", int(lastRunDuration.Microseconds()))
	} else if lastRunDuration < time.Second {
		lastRunTimeStr = fmt.Sprintf("%d milliseconds ago", int(lastRunDuration.Milliseconds()))
	} else if lastRunDuration < time.Minute {
		lastRunTimeStr = fmt.Sprintf("%d seconds ago", int(lastRunDuration.Seconds()))
	} else if lastRunDuration < time.Hour {
		lastRunTimeStr = fmt.Sprintf("%d minutes ago", int(lastRunDuration.Minutes()))
	} else if lastRunDuration < 24*time.Hour {
		lastRunTimeStr = fmt.Sprintf("%d hours ago", int(lastRunDuration.Hours()))
	} else {
		lastRunTimeStr = fmt.Sprintf("%d days ago", int(lastRunDuration.Hours()/24))
	}

	status := fmt.Sprintf("Last Run Time: %s\nTotal Runs: %d\n", lastRunTimeStr, job.totalRuns)

	return status
}

func (jm *JobManager) IsJobRunning(userID int64) bool {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	_, exists := jm.jobs[userID]
	return exists
}
