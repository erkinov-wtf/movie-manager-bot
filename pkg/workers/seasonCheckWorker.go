package workers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/erkinov-wtf/movie-manager-bot/internal/api/media/image"
	tv2 "github.com/erkinov-wtf/movie-manager-bot/internal/api/media/tv"
	models2 "github.com/erkinov-wtf/movie-manager-bot/internal/models"
	"gopkg.in/telebot.v3"
	"log"
	"os"
	"sync"
	"time"
)

func (c *WorkerApiClient) GetShowDetails(apiID int, userId int64) (*tv2.TV, error) {
	log.Printf("[Worker] Attempting to fetch details for show ID: %d", apiID)

	err := c.limiter.Wait(context.Background())
	if err != nil {
		log.Printf("[Worker] Rate limit wait error for show %d: %v", apiID, err)
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	start := time.Now()
	tvData, err := tv2.GetTV(apiID, userId)
	duration := time.Since(start)

	if err != nil {
		log.Printf("[Worker] API request failed for show %d after %v: %v", apiID, duration, err)
		return nil, fmt.Errorf("failed to get TV show details: %w", err)
	}

	log.Printf("[Worker] Successfully fetched show %d (%s) in %v", apiID, tvData.Name, duration)
	return tvData, nil
}

func (c *TVShowChecker) StartChecking(ctx context.Context, checkInterval time.Duration) {
	log.Printf("[Worker] Starting TV show checker with interval: %v", checkInterval)

	state, err := c.loadState()
	if err != nil {
		log.Printf("[Worker] Failed to load state: %v", err)
	}

	var initialDelay time.Duration
	if !state.LastCheckTime.IsZero() {
		nextCheckTime := state.LastCheckTime.Add(checkInterval)
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
			return
		case <-ticker.C:
			log.Printf("[Worker] Starting check cycle")
			start := time.Now()
			c.checkAllShows()

			if err = c.saveState(&CheckerState{
				LastCheckTime: start}); err != nil {
				log.Printf("[Worker] Failed to save state: %v", err)
			}

			log.Printf("[Worker] Completed check cycle in %v", time.Since(start))
		}
	}
}

type ShowRequest struct {
	User *models2.User
	Show models2.TVShows
}

func (c *TVShowChecker) checkAllShows() {
	var users []models2.User
	if err := c.db.Find(&users).Error; err != nil {
		log.Printf("[Worker] Error fetching users: %v", err)
		return
	}
	log.Printf("[Worker] Found %d users to process", len(users))

	showChan := make(chan ShowRequest, 1000)
	var wg sync.WaitGroup

	workerCount := 5
	log.Printf("[Worker] Starting %d workers", workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			log.Printf("[Worker-%d] Started", workerID)
			c.showWorker(showChan, &wg)
			log.Printf("[Worker-%d] Finished", workerID)
		}(i)
	}

	showCount := 0
	for _, user := range users {
		var shows []models2.TVShows
		if err := c.db.Where("user_id = ?", user.ID).Find(&shows).Error; err != nil {
			log.Printf("[Worker] Error fetching shows for user %d: %v", user.ID, err)
			continue
		}
		log.Printf("[Worker] Found %d shows for user %d", len(shows), user.ID)

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
	wg.Wait()
	log.Printf("[Worker] All workers completed")
}

func (c *TVShowChecker) showWorker(showChan chan ShowRequest, wg *sync.WaitGroup) {
	defer wg.Done()

	for req := range showChan {
		start := time.Now()
		log.Printf("[Worker] Processing show %d for user %d", req.Show.ApiID, req.User.ID)
		c.processShow(req.User, req.Show)
		log.Printf("[Worker] Completed processing show %d in %v", req.Show.ApiID, time.Since(start))
	}
}

func (c *TVShowChecker) processShow(user *models2.User, show models2.TVShows) {
	details, err := c.apiClient.GetShowDetails(int(show.ApiID), user.ID)
	if err != nil {
		log.Printf("[Worker] Error fetching show details for %d: %v", show.ApiID, err)
		return
	}

	log.Printf("[Worker] Current seasons for show %d: DB=%d, API=%d",
		show.ApiID, show.Seasons, details.Seasons)

	if details.Seasons > show.Seasons {
		log.Printf("[Worker] New season detected for show %d (%s): %d -> %d",
			show.ApiID, details.Name, show.Seasons, details.Seasons)
		c.notifyUser(*user, &show, details)
	} else {
		log.Printf("[Worker] No new seasons for show %d (%s)", show.ApiID, details.Name)
	}
}

func (c *TVShowChecker) notifyUser(user models2.User, watched *models2.TVShows, show *tv2.TV) {
	log.Printf("[Worker] Sending notification to user %d for show %s (ID: %d, Seasons: %d)",
		user.ID, show.Name, show.ID, show.Seasons)

	// Retrieve TV poster image
	imgBuffer, err := image.GetImage(show.PosterPath)
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
	backButton := replyMarkup.Data("📝 Update Data", fmt.Sprintf("tv|select_seasons|%v", show.ID))
	replyMarkup.Inline(
		replyMarkup.Row(backButton),
	)

	// Send the TV details with poster and buttons
	imageFile := &telebot.Photo{
		File:    telebot.File{FileReader: bytes.NewReader(imgBuffer.Bytes())},
		Caption: caption,
	}

	_, err = c.bot.Send(&telebot.User{ID: user.ID}, imageFile, replyMarkup, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("[Worker] Failed to send TV details: %v", err)
		return
	}
	return
}

func (c *TVShowChecker) loadState() (*CheckerState, error) {
	c.stateMux.Lock()
	defer c.stateMux.Unlock()

	state := &CheckerState{}

	data, err := os.ReadFile(c.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return nil, err
	}

	if err = json.Unmarshal(data, state); err != nil {
		return nil, err
	}

	return state, nil
}

func (c *TVShowChecker) saveState(state *CheckerState) error {
	c.stateMux.Lock()
	defer c.stateMux.Unlock()

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(c.statePath, data, 0644)
}
