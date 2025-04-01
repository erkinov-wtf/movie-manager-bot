package workers

import (
	"context"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/tv"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/time/rate"
	"gopkg.in/telebot.v3"
	"sync"
	"time"
)

const (
	StatusIdle    = "idle"
	StatusRunning = "running"
	StatusError   = "error"

	TaskStatusRunning = "running"
	TaskStatusSuccess = "success"
	TaskStatusError   = "error"

	WorkerTypeTVShowChecker = "tv_show_checker"

	TaskTypeCheckShow     = "check_show"
	TaskTypeCheckAllShows = "check_all_shows"
)

type TVShowChecker struct {
	app       *app.App
	apiClient TVShowAPIClient
	bot       *telebot.Bot
	limiter   *rate.Limiter
	workerId  string
	stateMux  sync.Mutex
}

type WorkerApiClient struct {
	limiter *rate.Limiter
	app     *app.App // Added app for logging access
}

type TVShowAPIClient interface {
	GetShowDetails(app *app.App, apiId int, userId int64) (*tv.TV, error)
}

func NewWorkerApiClient(app *app.App, requestsPerSecond int) *WorkerApiClient {
	const op = "workers.NewWorkerApiClient"
	app.Logger.WorkerInfo(op, "Initializing API client with rate limit",
		"requests_per_second", requestsPerSecond)

	return &WorkerApiClient{
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond),
		app:     app,
	}
}

func NewTVShowChecker(app *app.App, bot *telebot.Bot, apiClient TVShowAPIClient) *TVShowChecker {
	const op = "workers.NewTVShowChecker"
	app.Logger.WorkerInfo(op, "Initializing TV Show Checker")

	// Generate a unique worker ID
	workerId := "tvshow-checker-" + time.Now().Format("20060102-150405")
	app.Logger.WorkerDebug(op, "Generated worker ID", "worker_id", workerId)

	// Initialize worker state in database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Initialize the worker state in the database
	app.Logger.WorkerDebug(op, "Initializing worker state in database", "worker_id", workerId)
	err := app.Repository.Worker.UpsertWorkerState(ctx, database.UpsertWorkerStateParams{
		WorkerID:      workerId,
		WorkerType:    WorkerTypeTVShowChecker,
		Status:        StatusIdle,
		LastCheckTime: pgtype.Timestamptz{},
		NextCheckTime: pgtype.Timestamptz{},
		Error:         nil,
		ShowsChecked:  0,
		UpdatesFound:  0,
		CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})

	if err != nil {
		app.Logger.WorkerError(op, "Failed to initialize worker state in database",
			"worker_id", workerId, "error", err.Error())
	} else {
		app.Logger.WorkerInfo(op, "Worker state initialized successfully",
			"worker_id", workerId, "type", WorkerTypeTVShowChecker)
	}

	return &TVShowChecker{
		app:       app,
		bot:       bot,
		apiClient: apiClient,
		workerId:  workerId,
	}
}
