package workers

import (
	"golang.org/x/time/rate"
	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
	"log"
	"movie-manager-bot/api/media/tv"
)

type TVShowChecker struct {
	db        *gorm.DB
	bot       *telebot.Bot
	apiClient TVShowAPIClient
	limiter   *rate.Limiter
}

type WorkerApiClient struct {
	limiter *rate.Limiter
}

type TVShowAPIClient interface {
	GetShowDetails(apiID int64) (*tv.TV, error)
}

func NewWorkerApiClient(requestsPerSecond float64) *WorkerApiClient {
	log.Printf("[Worker] Initializing API client with rate limit: %.2f req/sec", requestsPerSecond)
	return &WorkerApiClient{
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), int(requestsPerSecond)),
	}
}

func NewTVShowChecker(db *gorm.DB, bot *telebot.Bot, apiClient TVShowAPIClient) *TVShowChecker {
	log.Printf("[Worker] Initializing TV Show Checker")
	return &TVShowChecker{
		db:        db,
		bot:       bot,
		apiClient: apiClient,
	}
}
