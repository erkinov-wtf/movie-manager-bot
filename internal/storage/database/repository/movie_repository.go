package repository

import (
	"context"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
)

type MovieRepositoryInterface interface {
	GetUserMovies(ctx context.Context, userID int64) ([]database.GetUserMoviesRow, error)
	MovieExists(ctx context.Context, apiID int64, userID int64) (bool, error)
	CreateMovie(ctx context.Context, params database.CreateMovieParams) error
	UpdateMovie(ctx context.Context, params database.UpdateMovieParams) error
	SoftDeleteMovie(ctx context.Context, apiID int64, userID int64) error
}

type MovieRepository struct {
	q *database.Queries
}

func NewMovieRepository(db database.DBTX) MovieRepositoryInterface {
	return &MovieRepository{
		q: database.New(db),
	}
}

func (r *MovieRepository) GetUserMovies(ctx context.Context, userID int64) ([]database.GetUserMoviesRow, error) {
	return r.q.GetUserMovies(ctx, userID)
}

func (r *MovieRepository) MovieExists(ctx context.Context, apiID int64, userID int64) (bool, error) {
	return r.q.MovieExists(ctx, database.MovieExistsParams{
		ApiID:  apiID,
		UserID: userID,
	})
}

func (r *MovieRepository) CreateMovie(ctx context.Context, params database.CreateMovieParams) error {
	return r.q.CreateMovie(ctx, params)
}

func (r *MovieRepository) UpdateMovie(ctx context.Context, params database.UpdateMovieParams) error {
	return r.q.UpdateMovie(ctx, params)
}

func (r *MovieRepository) SoftDeleteMovie(ctx context.Context, apiID int64, userID int64) error {
	return r.q.SoftDeleteMovie(ctx, database.SoftDeleteMovieParams{
		ApiID:  apiID,
		UserID: userID,
	})
}
