package workers

import (
	"github.com/erkinov-wtf/movie-manager-bot/api/media/tv"
	"golang.org/x/time/rate"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CheckerState struct {
	LastCheckTime time.Time `json:"last_check_time"`
}

type TVShowChecker struct {
	db        *gorm.DB
	bot       *telebot.Bot
	apiClient TVShowAPIClient
	limiter   *rate.Limiter
	statePath string
	stateMux  sync.Mutex
}

type WorkerApiClient struct {
	limiter *rate.Limiter
}

type TVShowAPIClient interface {
	GetShowDetails(apiID int, userId int64) (*tv.TV, error)
}

func NewWorkerApiClient(requestsPerSecond float64) *WorkerApiClient {
	log.Printf("[Worker] Initializing API client with rate limit: %.2f req/sec", requestsPerSecond)
	return &WorkerApiClient{
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), int(requestsPerSecond)),
	}
}

func NewTVShowChecker(db *gorm.DB, bot *telebot.Bot, apiClient TVShowAPIClient) *TVShowChecker {
	log.Printf("[Worker] Initializing TV Show Checker")

	stateDir := "worker_state"
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		log.Printf("[Worker] Failed to create state directory: %v", err)
	}

	return &TVShowChecker{
		db:        db,
		bot:       bot,
		apiClient: apiClient,
		statePath: filepath.Join(stateDir, "tv_checker_state.json"),
	}
}
