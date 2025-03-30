package workers

import (
	"bytes"
	"context"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/config/app"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/image"
	"github.com/erkinov-wtf/movie-manager-bot/internal/tmdb/tv"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"gopkg.in/telebot.v3"
	"log"
	"sync"
	"time"
)

func (c *WorkerApiClient) GetShowDetails(app *app.App, apiId int, userId int64) (*tv.TV, error) {
	log.Printf("[Worker] Attempting to fetch details for show Id: %d", apiId)

	err := c.limiter.Wait(context.Background())
	if err != nil {
		log.Printf("[Worker] Rate limit wait error for show %d: %v", apiId, err)
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	start := time.Now()
	tvData, err := tv.GetTV(app, apiId, userId)
	duration := time.Since(start)

	if err != nil {
		log.Printf("[Worker] API request failed for show %d after %v: %v", apiId, duration, err)
		return nil, fmt.Errorf("failed to get TV show details: %w", err)
	}

	log.Printf("[Worker] Successfully fetched show %d (%s) in %v", apiId, tvData.Name, duration)
	return tvData, nil
}

func (c *TVShowChecker) StartChecking(ctx context.Context, checkInterval time.Duration) {
	log.Printf("[Worker] Starting TV show checker with interval: %v", checkInterval)

	// Get the current worker state from the database
	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	workerState, err := c.app.Repository.Worker.GetWorkerState(dbCtx, c.workerId)
	if err != nil {
		log.Printf("[Worker] Failed to get worker state: %v", err)
		// Initialize if not found - this should generally be handled by NewTVShowChecker
	}

	var initialDelay time.Duration
	if workerState.LastCheckTime.Valid {
		nextCheckTime := workerState.LastCheckTime.Time.Add(checkInterval)
		if time.Now().Before(nextCheckTime) {
			initialDelay = time.Until(nextCheckTime)
		}
	}

	if initialDelay > 0 {
		log.Printf("[Worker] Waiting %v until next check", initialDelay)
		select {
		case <-ctx.Done():
			return
		case <-time.After(initialDelay):
		}
	}

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Worker] Context cancelled, stopping TV show checker")
			// Update worker status to idle
			c.updateWorkerStatus(StatusIdle, nil)
			return
		case <-ticker.C:
			log.Printf("[Worker] Starting check cycle")
			start := time.Now()

			// Update worker status to running
			c.updateWorkerStatus(StatusRunning, nil)

			// Create a new task for this check cycle
			taskID, err := c.createWorkerTask(TaskTypeCheckAllShows, nil, 0)
			if err != nil {
				log.Printf("[Worker] Failed to create task record: %v", err)
			}

			shows, updates := c.checkAllShows()

			// Update task on completion
			c.completeWorkerTask(taskID, nil, shows, updates)

			// Update worker state
			c.updateWorkerCheck(start, shows, updates)

			log.Printf("[Worker] Completed check cycle in %v", time.Since(start))
		}
	}
}

type ShowRequest struct {
	User *database.GetUsersRow
	Show database.GetUserTVShowsRow
}

func (c *TVShowChecker) checkAllShows() (int, int) {
	ctxDb, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users, err := c.app.Repository.Users.GetUsers(ctxDb)
	if err != nil {
		log.Printf("[Worker] Error fetching users: %v", err)
		return 0, 0
	}
	log.Printf("[Worker] Found %d users to process", len(users))

	showChan := make(chan ShowRequest, 1000)
	resultChan := make(chan bool, 1000) // Channel to collect update results
	var wg sync.WaitGroup

	workerCount := 5
	log.Printf("[Worker] Starting %d workers", workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerId int) {
			log.Printf("[Worker-%d] Started", workerId)
			c.showWorker(showChan, &wg, resultChan)
			log.Printf("[Worker-%d] Finished", workerId)
		}(i)
	}

	showCount := 0
	for _, user := range users {
		shows, err := c.app.Repository.TVShows.GetUserTVShows(ctxDb, user.TgID)
		if err != nil {
			log.Printf("[Worker] Error fetching shows for user %d: %v", user.TgID, err)
			continue
		}

		log.Printf("[Worker] Found %d shows for user %d", len(shows), user.TgID)

		for _, show := range shows {
			showCount++
			showChan <- ShowRequest{
				User: &user,
				Show: show,
			}
		}
	}

	log.Printf("[Worker] Queued %d total shows for processing", showCount)
	close(showChan)

	// Wait for all workers to complete
	wg.Wait()
	close(resultChan)

	// Count updates
	updateCount := 0
	for update := range resultChan {
		if update {
			updateCount++
		}
	}

	log.Printf("[Worker] All workers completed. Processed %d shows, found %d updates", showCount, updateCount)
	return showCount, updateCount
}

func (c *TVShowChecker) showWorker(showChan chan ShowRequest, wg *sync.WaitGroup, resultChan chan bool) {
	defer wg.Done()

	for req := range showChan {
		start := time.Now()
		log.Printf("[Worker] Processing show %d for user %d", req.Show.ApiID, req.User.TgID)

		// Create a task for this show check
		taskID, err := c.createWorkerTask(TaskTypeCheckShow, nil, req.Show.ApiID)
		if err != nil {
			log.Printf("[Worker] Failed to create task record for show %d: %v", req.Show.ApiID, err)
		}

		updated := c.processShow(req.User, req.Show)
		resultChan <- updated

		// Complete the task
		var processedCount int
		var updatesCount int
		if updated {
			updatesCount = 1
			processedCount = 1
		} else {
			processedCount = 1
			updatesCount = 0
		}

		c.completeWorkerTask(taskID, nil, processedCount, updatesCount)

		log.Printf("[Worker] Completed processing show %d in %v", req.Show.ApiID, time.Since(start))
	}
}

func (c *TVShowChecker) processShow(user *database.GetUsersRow, show database.GetUserTVShowsRow) bool {
	details, err := c.apiClient.GetShowDetails(c.app, int(show.ApiID), user.TgID)
	if err != nil {
		log.Printf("[Worker] Error fetching show details for %d: %v", show.ApiID, err)
		return false
	}

	log.Printf("[Worker] Current seasons for show %d: DB=%d, API=%d",
		show.ApiID, show.Seasons, details.Seasons)

	if details.Seasons > show.Seasons {
		log.Printf("[Worker] New season detected for show %d (%s): %d -> %d",
			show.ApiID, details.Name, show.Seasons, details.Seasons)
		c.notifyUser(*user, &show, details)
		return true
	} else {
		log.Printf("[Worker] No new seasons for show %d (%s)", show.ApiID, details.Name)
		return false
	}
}

func (c *TVShowChecker) notifyUser(user database.GetUsersRow, watched *database.GetUserTVShowsRow, show *tv.TV) {
	log.Printf("[Worker] Sending notification to user %d for show %s (Id: %d, Seasons: %d)",
		user.TgID, show.Name, show.Id, show.Seasons)

	// Retrieve TV poster image
	imgBuffer, err := image.GetImage(c.app, show.PosterPath)
	if err != nil {
		log.Printf("[Worker] Error retrieving image: %v", err)
		return
	}

	// Prepare TV details caption
	caption := fmt.Sprintf(
		"New Unwatched Seasons found\n\n"+
			"📺 *Name*: %v\n\n"+
			"📝 *Overview*: %v\n\n"+
			"📜 *Status*: %v\n\n"+
			"🎥 *Watched Seasons*: %v\n\n"+
			"🆕 *New Seasons*: %v\n",
		show.Name,
		show.Overview,
		show.Status,
		watched.Seasons,
		show.Seasons,
	)

	replyMarkup := &telebot.ReplyMarkup{}
	backButton := replyMarkup.Data("📝 Update Data", fmt.Sprintf("tv|select_seasons|%v", show.Id))
	replyMarkup.Inline(
		replyMarkup.Row(backButton),
	)

	// Send the TV details with poster and buttons
	imageFile := &telebot.Photo{
		File:    telebot.File{FileReader: bytes.NewReader(imgBuffer.Bytes())},
		Caption: caption,
	}

	_, err = c.bot.Send(&telebot.User{ID: user.TgID}, imageFile, replyMarkup, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("[Worker] Failed to send TV details: %v", err)
		return
	}
}

// Database interaction methods

// updateWorkerStatus updates the worker's status in the database
func (c *TVShowChecker) updateWorkerStatus(status string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var errorStr *string
	if err != nil {
		errText := err.Error()
		errorStr = &errText
	}

	now := time.Now()
	params := database.UpsertWorkerStateParams{
		WorkerID:   c.workerId,
		WorkerType: WorkerTypeTVShowChecker,
		Status:     status,
		Error:      errorStr,
		UpdatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
	}

	// Get current state to preserve other fields
	currentState, dbErr := c.app.Repository.Worker.GetWorkerState(ctx, c.workerId)
	if dbErr == nil {
		// Keep existing values
		params.LastCheckTime = currentState.LastCheckTime
		params.NextCheckTime = currentState.NextCheckTime
		params.ShowsChecked = currentState.ShowsChecked
		params.UpdatesFound = currentState.UpdatesFound
		params.CreatedAt = currentState.CreatedAt
	} else {
		// Set defaults for new record
		params.LastCheckTime = pgtype.Timestamptz{}
		params.NextCheckTime = pgtype.Timestamptz{}
		params.ShowsChecked = 0
		params.UpdatesFound = 0
		params.CreatedAt = pgtype.Timestamptz{Time: now, Valid: true}
	}

	dbErr = c.app.Repository.Worker.UpsertWorkerState(ctx, params)
	if dbErr != nil {
		log.Printf("[Worker] Failed to update worker status: %v", dbErr)
	}
}

// updateWorkerCheck records a completed check cycle
func (c *TVShowChecker) updateWorkerCheck(checkTime time.Time, showsChecked, updatesFound int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get current state
	currentState, err := c.app.Repository.Worker.GetWorkerState(ctx, c.workerId)
	if err != nil {
		log.Printf("[Worker] Failed to get worker state: %v", err)
		return
	}

	nextCheckTime := checkTime.Add(24 * time.Hour) // Default to 24 hours if no interval specified

	params := database.UpsertWorkerStateParams{
		WorkerID:      c.workerId,
		WorkerType:    WorkerTypeTVShowChecker,
		Status:        StatusIdle, // Set to idle after completion
		LastCheckTime: pgtype.Timestamptz{Time: checkTime, Valid: true},
		NextCheckTime: pgtype.Timestamptz{Time: nextCheckTime, Valid: true},
		ShowsChecked:  currentState.ShowsChecked + int32(showsChecked),
		UpdatesFound:  currentState.UpdatesFound + int32(updatesFound),
		Error:         nil, // Clear any previous error
		CreatedAt:     currentState.CreatedAt,
		UpdatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	err = c.app.Repository.Worker.UpsertWorkerState(ctx, params)
	if err != nil {
		log.Printf("[Worker] Failed to update worker check info: %v", err)
	}
}

// createWorkerTask creates a new task record in the database
func (c *TVShowChecker) createWorkerTask(taskType string, userID *int64, showID int64) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()

	// Convert nullable parameters
	var userIDVal *int64
	if userID != nil {
		userIDVal = userID
	}

	// Only set showID if it's a valid show
	var showIDVal *int64
	if showID > 0 {
		showIDVal = &showID
	}

	params := database.CreateWorkerTaskParams{
		WorkerID:  c.workerId,
		TaskType:  taskType,
		Status:    TaskStatusRunning,
		StartTime: pgtype.Timestamptz{Time: now, Valid: true},
		EndTime:   pgtype.Timestamptz{}, // Will be set on completion
		UserID:    userIDVal,
		ShowID:    showIDVal,
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	taskID, err := c.app.Repository.Worker.CreateWorkerTask(ctx, params)
	if err != nil {
		log.Printf("[Worker] Failed to create worker task: %v", err)
		return uuid.Nil, err
	}

	return taskID, nil
}

// completeWorkerTask updates a task as completed or failed
func (c *TVShowChecker) completeWorkerTask(taskID uuid.UUID, err error, showsChecked, updatesFound int) {
	if taskID == uuid.Nil {
		log.Printf("[Worker] Cannot update task with nil UUID")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()

	var status string
	var errorStr *string
	if err != nil {
		status = TaskStatusError
		errText := err.Error()
		errorStr = &errText
	} else {
		status = TaskStatusSuccess
	}

	// Get the original task to calculate duration
	task, getErr := c.app.Repository.Worker.GetWorkerTask(ctx, taskID)
	if getErr != nil {
		log.Printf("[Worker] Failed to get task for ID %s: %v", taskID, getErr)
		return
	}

	// Calculate duration in milliseconds
	var durationMs int64
	if task.StartTime.Valid {
		durationMs = now.Sub(task.StartTime.Time).Milliseconds()
	}

	params := database.UpdateWorkerTaskParams{
		ID:           taskID,
		Status:       status,
		EndTime:      pgtype.Timestamptz{Time: now, Valid: true},
		DurationMs:   &durationMs,
		Error:        errorStr,
		ShowsChecked: int32(showsChecked),
		UpdatesFound: int32(updatesFound),
	}

	err = c.app.Repository.Worker.UpdateWorkerTask(ctx, params)
	if err != nil {
		log.Printf("[Worker] Failed to update worker task: %v", err)
	}
}
