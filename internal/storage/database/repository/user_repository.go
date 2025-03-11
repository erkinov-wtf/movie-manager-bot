package repository

import (
	"context"
	"errors"
	"github.com/erkinov-wtf/movie-manager-bot/internal/storage/database"
)

type UserRepositoryInterface interface {
	GetUser(ctx context.Context, id int64) (database.User, error)
	GetUsers(ctx context.Context) ([]database.GetUsersRow, error)
	UserExists(ctx context.Context, id int64) (bool, error)
	CreateUser(ctx context.Context, params database.CreateUserParams) error
	UpdateUserTMDBKey(ctx context.Context, id int64, tmdbAPIKey string) error
	GetUserTMDBKey(ctx context.Context, id int64) (string, error)
}

type UserRepository struct {
	query *database.Queries
}

func NewUserRepository(db database.DBTX) UserRepositoryInterface {
	return &UserRepository{
		query: database.New(db),
	}
}

func (r *UserRepository) GetUser(ctx context.Context, id int64) (database.User, error) {
	return r.query.GetUser(ctx, id)
}

func (r *UserRepository) GetUsers(ctx context.Context) ([]database.GetUsersRow, error) {
	return r.query.GetUsers(ctx)
}

func (r *UserRepository) UserExists(ctx context.Context, id int64) (bool, error) {
	return r.query.UserExists(ctx, id)
}

func (r *UserRepository) CreateUser(ctx context.Context, params database.CreateUserParams) error {
	return r.query.CreateUser(ctx, params)
}

func (r *UserRepository) UpdateUserTMDBKey(ctx context.Context, id int64, tmdbAPIKey string) error {
	return r.query.UpdateUserTMDBKey(ctx, database.UpdateUserTMDBKeyParams{
		TgID:       id,
		TmdbApiKey: &tmdbAPIKey,
	})
}

func (r *UserRepository) GetUserTMDBKey(ctx context.Context, id int64) (string, error) {
	key, err := r.query.GetUserTMDBKey(ctx, id)
	if err != nil {
		return "", err
	}
	if *key != "" {
		return "", errors.New("TMDB API key not set")
	}
	return *key, nil
}
