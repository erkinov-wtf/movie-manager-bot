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
	"sync"
	"time"
)

func (c *WorkerApiClient) GetShowDetails(app *app.App, apiId int, userId int64) (*tv.TV, error) {
	const op = "workers.GetShowDetails"
	app.Logger.WorkerDebug(op, "Attempting to fetch details for show",
		"show_id", apiId, "user_id", userId)

	err := c.limiter.Wait(context.Background())
	if err != nil {
		app.Logger.WorkerError(op, "Rate limit wait error",
			"show_id", apiId, "error", err.Error())
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	start := time.Now()
	tvData, err := tv.GetTV(app, apiId, userId)
	duration := time.Since(start)

	if err != nil {
		app.Logger.WorkerError(op, "API request failed",
			"show_id", apiId, "duration_ms", duration.Milliseconds(), "error", err.Error())
		return nil, fmt.Errorf("failed to get TV show details: %w", err)
	}

	app.Logger.WorkerInfo(op, "Successfully fetched show details",
		"show_id", apiId, "name", tvData.Name, "duration_ms", duration.Milliseconds())
	return tvData, nil
}

func (c *TVShowChecker) StartChecking(ctx context.Context, checkInterval int) {
	const op = "workers.StartChecking"
	c.app.Logger.WorkerInfo(op, "Starting TV show checker",
		"worker_id", c.workerId, "check_interval_hours", checkInterval)

	// Get the current worker state from the database
	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	workerState, err := c.app.Repository.Worker.GetWorkerState(dbCtx, c.workerId)
	if err != nil {
		c.app.Logger.WorkerError(op, "Failed to get worker state",
			"worker_id", c.workerId, "error", err.Error())
	}

	timeDuration := time.Duration(checkInterval) * time.Hour

	var initialDelay time.Duration
	if workerState.LastCheckTime.Valid {
		nextCheckTime := workerState.LastCheckTime.Time.Add(timeDuration)
		if time.Now().Before(nextCheckTime) {
			initialDelay = time.Until(nextCheckTime)
		}
	}

	if initialDelay > 0 {
		c.app.Logger.WorkerInfo(op, "Waiting until next check",
			"worker_id", c.workerId, "delay_minutes", initialDelay.Minutes())
		select {
		case <-ctx.Done():
			return
		case <-time.After(initialDelay):
		}
	}

	ticker := time.NewTicker(timeDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.app.Logger.WorkerInfo(op, "Context cancelled, stopping TV show checker",
				"worker_id", c.workerId)
			// Update worker status to idle
			c.updateWorkerStatus(StatusIdle, nil)
			return
		case <-ticker.C:
			c.app.Logger.WorkerInfo(op, "Starting check cycle", "worker_id", c.workerId)
			start := time.Now()

			// Update worker status to running
			c.updateWorkerStatus(StatusRunning, nil)

			// Create a new task for this check cycle
			taskID, err := c.createWorkerTask(TaskTypeCheckAllShows, nil, 0)
			if err != nil {
				c.app.Logger.WorkerError(op, "Failed to create task record",
					"worker_id", c.workerId, "error", err.Error())
			}

			shows, updates := c.checkAllShows()

			// Update task on completion
			c.completeWorkerTask(taskID, nil, shows, updates)

			// Update worker state
			c.updateWorkerCheck(start, shows, updates)

			c.app.Logger.WorkerInfo(op, "Completed check cycle",
				"worker_id", c.workerId,
				"duration_ms", time.Since(start).Milliseconds(),
				"shows_checked", shows,
				"updates_found", updates)
		}
	}
}

type ShowRequest struct {
	User *database.GetUsersRow
	Show database.GetUserTVShowsRow
}

func (c *TVShowChecker) checkAllShows() (int, int) {
	const op = "workers.checkAllShows"
	ctxDb, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users, err := c.app.Repository.Users.GetUsers(ctxDb)
	if err != nil {
		c.app.Logger.WorkerError(op, "Error fetching users", "error", err.Error())
		return 0, 0
	}
	c.app.Logger.WorkerInfo(op, "Found users to process", "user_count", len(users))

	showChan := make(chan ShowRequest, 1000)
	resultChan := make(chan bool, 1000) // Channel to collect update results
	var wg sync.WaitGroup

	workerCount := 5
	c.app.Logger.WorkerInfo(op, "Starting workers", "worker_count", workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerId int) {
			c.app.Logger.WorkerDebug(op, "Worker started", "worker_index", workerId)
			c.showWorker(showChan, &wg, resultChan)
			c.app.Logger.WorkerDebug(op, "Worker finished", "worker_index", workerId)
		}(i)
	}

	showCount := 0
	for _, user := range users {
		shows, err := c.app.Repository.TVShows.GetUserTVShows(ctxDb, user.TgID)
		if err != nil {
			c.app.Logger.WorkerError(op, "Error fetching shows for user",
				"user_id", user.TgID, "error", err.Error())
			continue
		}

		c.app.Logger.WorkerDebug(op, "Found shows for user",
			"user_id", user.TgID, "show_count", len(shows))

		for _, show := range shows {
			showCount++
			showChan <- ShowRequest{
				User: &user,
				Show: show,
			}
		}
	}

	c.app.Logger.WorkerInfo(op, "Queued shows for processing", "total_shows", showCount)
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

	c.app.Logger.WorkerInfo(op, "All workers completed",
		"shows_processed", showCount, "updates_found", updateCount)
	return showCount, updateCount
}

func (c *TVShowChecker) showWorker(showChan chan ShowRequest, wg *sync.WaitGroup, resultChan chan bool) {
	const op = "workers.showWorker"
	defer wg.Done()

	for req := range showChan {
		start := time.Now()
		c.app.Logger.WorkerDebug(op, "Processing show",
			"show_id", req.Show.ApiID, "user_id", req.User.TgID)

		// Create a task for this show check
		taskID, err := c.createWorkerTask(TaskTypeCheckShow, nil, req.Show.ApiID)
		if err != nil {
			c.app.Logger.WorkerError(op, "Failed to create task record for show",
				"show_id", req.Show.ApiID, "error", err.Error())
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

		c.app.Logger.WorkerDebug(op, "Completed processing show",
			"show_id", req.Show.ApiID,
			"duration_ms", time.Since(start).Milliseconds(),
			"updated", updated)
	}
}

func (c *TVShowChecker) processShow(user *database.GetUsersRow, show database.GetUserTVShowsRow) bool {
	const op = "workers.processShow"
	details, err := c.apiClient.GetShowDetails(c.app, int(show.ApiID), user.TgID)
	if err != nil {
		c.app.Logger.WorkerError(op, "Error fetching show details",
			"show_id", show.ApiID, "error", err.Error())
		return false
	}

	c.app.Logger.WorkerDebug(op, "Comparing seasons for show",
		"show_id", show.ApiID, "db_seasons", show.Seasons, "api_seasons", details.Seasons)

	if details.Seasons > show.Seasons {
		c.app.Logger.WorkerInfo(op, "New season detected for show",
			"show_id", show.ApiID, "name", details.Name,
			"old_seasons", show.Seasons, "new_seasons", details.Seasons)
		c.notifyUser(*user, &show, details)
		return true
	} else {
		c.app.Logger.WorkerDebug(op, "No new seasons for show",
			"show_id", show.ApiID, "name", details.Name)
		return false
	}
}

func (c *TVShowChecker) notifyUser(user database.GetUsersRow, watched *database.GetUserTVShowsRow, show *tv.TV) {
	const op = "workers.notifyUser"
	c.app.Logger.WorkerInfo(op, "Sending notification to user",
		"user_id", user.TgID, "show_id", show.Id,
		"name", show.Name, "seasons", show.Seasons)

	// Retrieve TV poster image
	imgBuffer, err := image.GetImage(c.app, show.PosterPath)
	if err != nil {
		c.app.Logger.WorkerError(op, "Error retrieving image",
			"poster_path", show.PosterPath, "error", err.Error())
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

	c.app.Logger.WorkerDebug(op, "Sending notification message", "user_id", user.TgID)
	_, err = c.bot.Send(&telebot.User{ID: user.TgID}, imageFile, replyMarkup, telebot.ModeMarkdown)
	if err != nil {
		c.app.Logger.WorkerError(op, "Failed to send TV details", "user_id", user.TgID, "error", err.Error())
		return
	}

	c.app.Logger.WorkerInfo(op, "Notification sent successfully", "user_id", user.TgID, "show_id", show.Id)
}

// Database interaction methods

// updateWorkerStatus updates the worker's status in the database
func (c *TVShowChecker) updateWorkerStatus(status string, err error) {
	const op = "workers.updateWorkerStatus"
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
		c.app.Logger.WorkerError(op, "Failed to update worker status",
			"worker_id", c.workerId, "status", status, "error", dbErr.Error())
	}
}

// updateWorkerCheck records a completed check cycle
func (c *TVShowChecker) updateWorkerCheck(checkTime time.Time, showsChecked, updatesFound int) {
	const op = "workers.updateWorkerCheck"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get current state
	currentState, err := c.app.Repository.Worker.GetWorkerState(ctx, c.workerId)
	if err != nil {
		c.app.Logger.WorkerError(op, "Failed to get worker state",
			"worker_id", c.workerId, "error", err.Error())
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
		c.app.Logger.WorkerError(op, "Failed to update worker check info",
			"worker_id", c.workerId, "error", err.Error())
	} else {
		c.app.Logger.WorkerDebug(op, "Updated worker check info",
			"worker_id", c.workerId,
			"shows_checked", showsChecked,
			"updates_found", updatesFound,
			"next_check", nextCheckTime)
	}
}

// createWorkerTask creates a new task record in the database
func (c *TVShowChecker) createWorkerTask(taskType string, userID *int64, showID int64) (uuid.UUID, error) {
	const op = "workers.createWorkerTask"
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

	c.app.Logger.WorkerDebug(op, "Creating worker task",
		"worker_id", c.workerId, "task_type", taskType, "show_id", showID)

	taskID, err := c.app.Repository.Worker.CreateWorkerTask(ctx, params)
	if err != nil {
		c.app.Logger.WorkerError(op, "Failed to create worker task",
			"worker_id", c.workerId, "task_type", taskType, "error", err.Error())
		return uuid.Nil, err
	}

	c.app.Logger.WorkerDebug(op, "Worker task created",
		"task_id", taskID, "worker_id", c.workerId, "task_type", taskType)
	return taskID, nil
}

// completeWorkerTask updates a task as completed or failed
func (c *TVShowChecker) completeWorkerTask(taskID uuid.UUID, err error, showsChecked, updatesFound int) {
	const op = "workers.completeWorkerTask"
	if taskID == uuid.Nil {
		c.app.Logger.WorkerWarning(op, "Cannot update task with nil UUID", "worker_id", c.workerId)
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
		c.app.Logger.WorkerError(op, "Failed to get task",
			"task_id", taskID, "worker_id", c.workerId, "error", getErr.Error())
		return
	}

	// Calculate duration in milliseconds
	var durationMs int64
	if task.StartTime.Valid {
		durationMs = now.Sub(task.StartTime.Time).Milliseconds()
	}

	var v1 int32 = int32(showsChecked)
	var v2 int32 = int32(updatesFound)

	params := database.UpdateWorkerTaskParams{
		ID:           taskID,
		Status:       status,
		EndTime:      pgtype.Timestamptz{Time: now, Valid: true},
		DurationMs:   &durationMs,
		Error:        errorStr,
		ShowsChecked: &v1,
		UpdatesFound: &v2,
	}

	err = c.app.Repository.Worker.UpdateWorkerTask(ctx, params)
	if err != nil {
		c.app.Logger.WorkerError(op, "Failed to update worker task",
			"task_id", taskID, "worker_id", c.workerId, "error", err.Error())
	} else {
		c.app.Logger.WorkerDebug(op, "Worker task completed",
			"task_id", taskID, "worker_id", c.workerId,
			"status", status, "duration_ms", durationMs,
			"shows_checked", showsChecked, "updates_found", updatesFound)
	}
}
