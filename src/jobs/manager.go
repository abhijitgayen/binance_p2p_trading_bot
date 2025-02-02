package jobs

import (
	"go_binance_bot/src/apis"
	"go_binance_bot/src/helpers/priority_queue"
	"sync"

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

func (jm *JobManager) StopJob(userID int64) {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	if job, exists := jm.jobs[userID]; exists {
		job.Stop()
		delete(jm.jobs, userID)
	}
}
