package repository

import (
	"context"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
	"strings"
)

type TVShowRepositoryInterface interface {
	GetUserTVShows(ctx context.Context, userID int64) ([]database.GetUserTVShowsRow, error)
	GetWatchedSeasons(ctx context.Context, apiID int64, userID int64) (int32, error)
	TVShowExists(ctx context.Context, apiID int64, userID int64) (bool, error)
	CreateTVShow(ctx context.Context, params database.CreateTVShowParams) error
	UpdateTVShow(ctx context.Context, params database.UpdateTVShowParams) error
	SoftDeleteTVShow(ctx context.Context, apiID int64, userID int64) error
}

type TVShowRepository struct {
	q *database.Queries
}

// NewTVShowRepository creates a new TV show repository
func NewTVShowRepository(db database.DBTX) TVShowRepositoryInterface {
	return &TVShowRepository{
		q: database.New(db),
	}
}

func (r *TVShowRepository) GetUserTVShows(ctx context.Context, userID int64) ([]database.GetUserTVShowsRow, error) {
	return r.q.GetUserTVShows(ctx, userID)
}

func (r *TVShowRepository) GetWatchedSeasons(ctx context.Context, apiID int64, userID int64) (int32, error) {
	seasons, err := r.q.GetWatchedSeasons(ctx, database.GetWatchedSeasonsParams{
		ApiID:  apiID,
		UserID: userID,
	})

	if err != nil && strings.Contains(err.Error(), "no rows in result set") {
		return 0, nil
	}

	return seasons, err
}

func (r *TVShowRepository) TVShowExists(ctx context.Context, apiID int64, userID int64) (bool, error) {
	return r.q.TVShowExists(ctx, database.TVShowExistsParams{
		ApiID:  apiID,
		UserID: userID,
	})
}

func (r *TVShowRepository) CreateTVShow(ctx context.Context, params database.CreateTVShowParams) error {
	return r.q.CreateTVShow(ctx, params)
}

func (r *TVShowRepository) UpdateTVShow(ctx context.Context, params database.UpdateTVShowParams) error {
	return r.q.UpdateTVShow(ctx, params)
}

func (r *TVShowRepository) SoftDeleteTVShow(ctx context.Context, apiID int64, userID int64) error {
	return r.q.SoftDeleteTVShow(ctx, database.SoftDeleteTVShowParams{
		ApiID:  apiID,
		UserID: userID,
	})
}
