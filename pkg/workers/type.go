package workers

import (
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/tv"
	"golang.org/x/time/rate"
	"gopkg.in/telebot.v3"
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
	app       *app.App
	apiClient TVShowAPIClient
	bot       *telebot.Bot
	limiter   *rate.Limiter
	statePath string
	stateMux  sync.Mutex
}

type WorkerApiClient struct {
	limiter *rate.Limiter
}

type TVShowAPIClient interface {
	GetShowDetails(app *app.App, apiId int, userId int64) (*tv.TV, error)
}

func NewWorkerApiClient(requestsPerSecond float64) *WorkerApiClient {
	log.Printf("[Worker] Initializing API client with rate limit: %.2f req/sec", requestsPerSecond)
	return &WorkerApiClient{
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), int(requestsPerSecond)),
	}
}

func NewTVShowChecker(app *app.App, bot *telebot.Bot, apiClient TVShowAPIClient) *TVShowChecker {
	log.Printf("[Worker] Initializing TV Show Checker")

	stateDir := "worker_state"
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		log.Printf("[Worker] Failed to create state directory: %v", err)
	}

	return &TVShowChecker{
		app:       app,
		bot:       bot,
		apiClient: apiClient,
		statePath: filepath.Join(stateDir, "tv_checker_state.json"),
	}
}
