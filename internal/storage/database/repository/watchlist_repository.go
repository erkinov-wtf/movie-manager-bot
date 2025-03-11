package repository

import (
	"context"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
)

type WatchlistRepositoryInterface interface {
	CreateWatchlist(ctx context.Context, params database.CreateWatchlistParams) error
	GetUserWatchlist(ctx context.Context, showAPIID int64, userID int64) (database.GetUserWatchlistRow, error)
	WatchlistExists(ctx context.Context, showAPIID int64, userID int64, showType string) (bool, error)
	GetUserWatchlists(ctx context.Context, userID int64) ([]database.GetUserWatchlistsRow, error)
	GetUserWatchlistsWithType(ctx context.Context, userID int64, showType string) ([]database.GetUserWatchlistsWithTypeRow, error)
	DeleteWatchlist(ctx context.Context, showAPIID int64, userID int64) error
}

type WatchlistRepository struct {
	q *database.Queries
}

// NewWatchlistRepository creates a new watchlist repository
func NewWatchlistRepository(db database.DBTX) WatchlistRepositoryInterface {
	return &WatchlistRepository{
		q: database.New(db),
	}
}

func (r *WatchlistRepository) CreateWatchlist(ctx context.Context, params database.CreateWatchlistParams) error {
	return r.q.CreateWatchlist(ctx, params)
}

func (r *WatchlistRepository) GetUserWatchlist(ctx context.Context, showAPIID int64, userID int64) (database.GetUserWatchlistRow, error) {
	return r.q.GetUserWatchlist(ctx, database.GetUserWatchlistParams{
		ShowApiID: showAPIID,
		UserID:    userID,
	})
}

func (r *WatchlistRepository) WatchlistExists(ctx context.Context, showAPIID int64, userID int64, showType string) (bool, error) {
	return r.q.WatchlistExists(ctx, database.WatchlistExistsParams{
		ShowApiID: showAPIID,
		UserID:    userID,
		Type:      showType,
	})
}

func (r *WatchlistRepository) GetUserWatchlists(ctx context.Context, userID int64) ([]database.GetUserWatchlistsRow, error) {
	return r.q.GetUserWatchlists(ctx, userID)
}

func (r *WatchlistRepository) GetUserWatchlistsWithType(ctx context.Context, userID int64, showType string) ([]database.GetUserWatchlistsWithTypeRow, error) {
	return r.q.GetUserWatchlistsWithType(ctx, database.GetUserWatchlistsWithTypeParams{
		UserID: userID,
		Type:   showType,
	})
}

func (r *WatchlistRepository) DeleteWatchlist(ctx context.Context, showAPIID int64, userID int64) error {
	return r.q.DeleteWatchlist(ctx, database.DeleteWatchlistParams{
		ShowApiID: showAPIID,
		UserID:    userID,
	})
}
