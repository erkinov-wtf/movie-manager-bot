// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createMovie = `-- name: CreateMovie :exec
INSERT INTO movies (user_id, api_id, title, runtime)
VALUES ($1, $2, $3, $4)
`

type CreateMovieParams struct {
	UserID  int64  `json:"user_id"`
	ApiID   int64  `json:"api_id"`
	Title   string `json:"title"`
	Runtime int32  `json:"runtime"`
}

func (q *Queries) CreateMovie(ctx context.Context, arg CreateMovieParams) error {
	_, err := q.db.Exec(ctx, createMovie,
		arg.UserID,
		arg.ApiID,
		arg.Title,
		arg.Runtime,
	)
	return err
}

const createTVShow = `-- name: CreateTVShow :exec
INSERT INTO tv_shows (user_id, api_id, name, seasons, episodes, runtime, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
`

type CreateTVShowParams struct {
	UserID   int64  `json:"user_id"`
	ApiID    int64  `json:"api_id"`
	Name     string `json:"name"`
	Seasons  int32  `json:"seasons"`
	Episodes int32  `json:"episodes"`
	Runtime  int32  `json:"runtime"`
	Status   string `json:"status"`
}

func (q *Queries) CreateTVShow(ctx context.Context, arg CreateTVShowParams) error {
	_, err := q.db.Exec(ctx, createTVShow,
		arg.UserID,
		arg.ApiID,
		arg.Name,
		arg.Seasons,
		arg.Episodes,
		arg.Runtime,
		arg.Status,
	)
	return err
}

const createUser = `-- name: CreateUser :exec
INSERT INTO users (tg_id, first_name, last_name, username, language, tmdb_api_key)
VALUES ($1, $2, $3, $4, $5, $6)
`

type CreateUserParams struct {
	TgID       int64   `json:"tg_id"`
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	Username   *string `json:"username"`
	Language   string  `json:"language"`
	TmdbApiKey *string `json:"tmdb_api_key"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) error {
	_, err := q.db.Exec(ctx, createUser,
		arg.TgID,
		arg.FirstName,
		arg.LastName,
		arg.Username,
		arg.Language,
		arg.TmdbApiKey,
	)
	return err
}

const createWatchlist = `-- name: CreateWatchlist :exec

INSERT INTO watchlists (user_id, show_api_id, type, title, image)
VALUES ($1, $2, $3, $4, $5)
`

type CreateWatchlistParams struct {
	UserID    int64   `json:"user_id"`
	ShowApiID int64   `json:"show_api_id"`
	Type      string  `json:"type"`
	Title     string  `json:"title"`
	Image     *string `json:"image"`
}

// Watchlists Table
func (q *Queries) CreateWatchlist(ctx context.Context, arg CreateWatchlistParams) error {
	_, err := q.db.Exec(ctx, createWatchlist,
		arg.UserID,
		arg.ShowApiID,
		arg.Type,
		arg.Title,
		arg.Image,
	)
	return err
}

const createWorkerTask = `-- name: CreateWorkerTask :one
INSERT INTO worker_tasks (worker_id, task_type, status, start_time, end_time, duration_ms,
                          error, show_id, user_id, shows_checked, updates_found, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id
`

type CreateWorkerTaskParams struct {
	WorkerID     string             `json:"worker_id"`
	TaskType     string             `json:"task_type"`
	Status       string             `json:"status"`
	StartTime    pgtype.Timestamptz `json:"start_time"`
	EndTime      pgtype.Timestamptz `json:"end_time"`
	DurationMs   *int64             `json:"duration_ms"`
	Error        *string            `json:"error"`
	ShowID       *int64             `json:"show_id"`
	UserID       *int64             `json:"user_id"`
	ShowsChecked *int32             `json:"shows_checked"`
	UpdatesFound *int32             `json:"updates_found"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
}

func (q *Queries) CreateWorkerTask(ctx context.Context, arg CreateWorkerTaskParams) (uuid.UUID, error) {
	row := q.db.QueryRow(ctx, createWorkerTask,
		arg.WorkerID,
		arg.TaskType,
		arg.Status,
		arg.StartTime,
		arg.EndTime,
		arg.DurationMs,
		arg.Error,
		arg.ShowID,
		arg.UserID,
		arg.ShowsChecked,
		arg.UpdatesFound,
		arg.CreatedAt,
	)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

const deleteWatchlist = `-- name: DeleteWatchlist :exec
UPDATE watchlists
SET deleted_at = NOW()
WHERE show_api_id = $1
  AND user_id = $2
  AND deleted_at IS NULL
`

type DeleteWatchlistParams struct {
	ShowApiID int64 `json:"show_api_id"`
	UserID    int64 `json:"user_id"`
}

func (q *Queries) DeleteWatchlist(ctx context.Context, arg DeleteWatchlistParams) error {
	_, err := q.db.Exec(ctx, deleteWatchlist, arg.ShowApiID, arg.UserID)
	return err
}

const getRecentTasks = `-- name: GetRecentTasks :many
SELECT id,
       worker_id,
       task_type,
       status,
       start_time,
       end_time,
       duration_ms,
       error,
       show_id,
       user_id,
       shows_checked,
       updates_found,
       created_at
FROM worker_tasks
WHERE worker_id = $1
ORDER BY created_at DESC LIMIT $2
`

type GetRecentTasksParams struct {
	WorkerID string `json:"worker_id"`
	Limit    int32  `json:"limit"`
}

func (q *Queries) GetRecentTasks(ctx context.Context, arg GetRecentTasksParams) ([]WorkerTask, error) {
	rows, err := q.db.Query(ctx, getRecentTasks, arg.WorkerID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []WorkerTask
	for rows.Next() {
		var i WorkerTask
		if err := rows.Scan(
			&i.ID,
			&i.WorkerID,
			&i.TaskType,
			&i.Status,
			&i.StartTime,
			&i.EndTime,
			&i.DurationMs,
			&i.Error,
			&i.ShowID,
			&i.UserID,
			&i.ShowsChecked,
			&i.UpdatesFound,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUser = `-- name: GetUser :one

SELECT id, tg_id, first_name, last_name, username, language, tmdb_api_key, created_at, updated_at
FROM users
WHERE tg_id = $1 LIMIT 1
`

// Users Table
func (q *Queries) GetUser(ctx context.Context, tgID int64) (User, error) {
	row := q.db.QueryRow(ctx, getUser, tgID)
	var i User
	err := row.Scan(
		&i.ID,
		&i.TgID,
		&i.FirstName,
		&i.LastName,
		&i.Username,
		&i.Language,
		&i.TmdbApiKey,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserMovies = `-- name: GetUserMovies :many

SELECT id, api_id, title, runtime, created_at, updated_at
FROM movies
WHERE user_id = $1
  AND deleted_at IS NULL
`

type GetUserMoviesRow struct {
	ID        uuid.UUID          `json:"id"`
	ApiID     int64              `json:"api_id"`
	Title     string             `json:"title"`
	Runtime   int32              `json:"runtime"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

// Movies Table
func (q *Queries) GetUserMovies(ctx context.Context, userID int64) ([]GetUserMoviesRow, error) {
	rows, err := q.db.Query(ctx, getUserMovies, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUserMoviesRow
	for rows.Next() {
		var i GetUserMoviesRow
		if err := rows.Scan(
			&i.ID,
			&i.ApiID,
			&i.Title,
			&i.Runtime,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserTMDBKey = `-- name: GetUserTMDBKey :one
SELECT tmdb_api_key
FROM users
WHERE tg_id = $1 LIMIT 1
`

func (q *Queries) GetUserTMDBKey(ctx context.Context, tgID int64) (*string, error) {
	row := q.db.QueryRow(ctx, getUserTMDBKey, tgID)
	var tmdb_api_key *string
	err := row.Scan(&tmdb_api_key)
	return tmdb_api_key, err
}

const getUserTVShow = `-- name: GetUserTVShow :one
SELECT id,
       api_id,
       name,
       seasons,
       episodes,
       runtime,
       status,
       created_at,
       updated_at
FROM tv_shows
WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL
`

type GetUserTVShowParams struct {
	ApiID  int64 `json:"api_id"`
	UserID int64 `json:"user_id"`
}

type GetUserTVShowRow struct {
	ID        uuid.UUID          `json:"id"`
	ApiID     int64              `json:"api_id"`
	Name      string             `json:"name"`
	Seasons   int32              `json:"seasons"`
	Episodes  int32              `json:"episodes"`
	Runtime   int32              `json:"runtime"`
	Status    string             `json:"status"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

func (q *Queries) GetUserTVShow(ctx context.Context, arg GetUserTVShowParams) (GetUserTVShowRow, error) {
	row := q.db.QueryRow(ctx, getUserTVShow, arg.ApiID, arg.UserID)
	var i GetUserTVShowRow
	err := row.Scan(
		&i.ID,
		&i.ApiID,
		&i.Name,
		&i.Seasons,
		&i.Episodes,
		&i.Runtime,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserTVShows = `-- name: GetUserTVShows :many

SELECT id,
       api_id,
       name,
       seasons,
       episodes,
       runtime,
       status,
       created_at,
       updated_at
FROM tv_shows
WHERE user_id = $1
  AND deleted_at IS NULL
`

type GetUserTVShowsRow struct {
	ID        uuid.UUID          `json:"id"`
	ApiID     int64              `json:"api_id"`
	Name      string             `json:"name"`
	Seasons   int32              `json:"seasons"`
	Episodes  int32              `json:"episodes"`
	Runtime   int32              `json:"runtime"`
	Status    string             `json:"status"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

// TV Shows Table
func (q *Queries) GetUserTVShows(ctx context.Context, userID int64) ([]GetUserTVShowsRow, error) {
	rows, err := q.db.Query(ctx, getUserTVShows, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUserTVShowsRow
	for rows.Next() {
		var i GetUserTVShowsRow
		if err := rows.Scan(
			&i.ID,
			&i.ApiID,
			&i.Name,
			&i.Seasons,
			&i.Episodes,
			&i.Runtime,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserWatchlist = `-- name: GetUserWatchlist :one
SELECT id,
       user_id,
       show_api_id,
       type,
       title,
       image,
       created_at,
       updated_at
FROM watchlists
WHERE show_api_id = $1
  AND user_id = $2
  AND deleted_at IS NULL LIMIT 1
`

type GetUserWatchlistParams struct {
	ShowApiID int64 `json:"show_api_id"`
	UserID    int64 `json:"user_id"`
}

type GetUserWatchlistRow struct {
	ID        uuid.UUID          `json:"id"`
	UserID    int64              `json:"user_id"`
	ShowApiID int64              `json:"show_api_id"`
	Type      string             `json:"type"`
	Title     string             `json:"title"`
	Image     *string            `json:"image"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

func (q *Queries) GetUserWatchlist(ctx context.Context, arg GetUserWatchlistParams) (GetUserWatchlistRow, error) {
	row := q.db.QueryRow(ctx, getUserWatchlist, arg.ShowApiID, arg.UserID)
	var i GetUserWatchlistRow
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.ShowApiID,
		&i.Type,
		&i.Title,
		&i.Image,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserWatchlists = `-- name: GetUserWatchlists :many
SELECT id, show_api_id, type, title, image, created_at, updated_at
FROM watchlists
WHERE user_id = $1
  AND deleted_at IS NULL
`

type GetUserWatchlistsRow struct {
	ID        uuid.UUID          `json:"id"`
	ShowApiID int64              `json:"show_api_id"`
	Type      string             `json:"type"`
	Title     string             `json:"title"`
	Image     *string            `json:"image"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

func (q *Queries) GetUserWatchlists(ctx context.Context, userID int64) ([]GetUserWatchlistsRow, error) {
	rows, err := q.db.Query(ctx, getUserWatchlists, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUserWatchlistsRow
	for rows.Next() {
		var i GetUserWatchlistsRow
		if err := rows.Scan(
			&i.ID,
			&i.ShowApiID,
			&i.Type,
			&i.Title,
			&i.Image,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserWatchlistsWithType = `-- name: GetUserWatchlistsWithType :many
SELECT id, show_api_id, type, title, image, created_at, updated_at
FROM watchlists
WHERE user_id = $1
  AND type = $2
  AND deleted_at IS NULL
`

type GetUserWatchlistsWithTypeParams struct {
	UserID int64  `json:"user_id"`
	Type   string `json:"type"`
}

type GetUserWatchlistsWithTypeRow struct {
	ID        uuid.UUID          `json:"id"`
	ShowApiID int64              `json:"show_api_id"`
	Type      string             `json:"type"`
	Title     string             `json:"title"`
	Image     *string            `json:"image"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

func (q *Queries) GetUserWatchlistsWithType(ctx context.Context, arg GetUserWatchlistsWithTypeParams) ([]GetUserWatchlistsWithTypeRow, error) {
	rows, err := q.db.Query(ctx, getUserWatchlistsWithType, arg.UserID, arg.Type)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUserWatchlistsWithTypeRow
	for rows.Next() {
		var i GetUserWatchlistsWithTypeRow
		if err := rows.Scan(
			&i.ID,
			&i.ShowApiID,
			&i.Type,
			&i.Title,
			&i.Image,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUsers = `-- name: GetUsers :many
SELECT id, tg_id, first_name, last_name, username, language, created_at, updated_at
FROM users
`

type GetUsersRow struct {
	ID        uuid.UUID          `json:"id"`
	TgID      int64              `json:"tg_id"`
	FirstName *string            `json:"first_name"`
	LastName  *string            `json:"last_name"`
	Username  *string            `json:"username"`
	Language  string             `json:"language"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

func (q *Queries) GetUsers(ctx context.Context) ([]GetUsersRow, error) {
	rows, err := q.db.Query(ctx, getUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUsersRow
	for rows.Next() {
		var i GetUsersRow
		if err := rows.Scan(
			&i.ID,
			&i.TgID,
			&i.FirstName,
			&i.LastName,
			&i.Username,
			&i.Language,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getWatchedSeasons = `-- name: GetWatchedSeasons :one
SELECT seasons
FROM tv_shows
WHERE api_id = $1
  AND user_id = $2
  AND deleted_at IS NULL
`

type GetWatchedSeasonsParams struct {
	ApiID  int64 `json:"api_id"`
	UserID int64 `json:"user_id"`
}

func (q *Queries) GetWatchedSeasons(ctx context.Context, arg GetWatchedSeasonsParams) (int32, error) {
	row := q.db.QueryRow(ctx, getWatchedSeasons, arg.ApiID, arg.UserID)
	var seasons int32
	err := row.Scan(&seasons)
	return seasons, err
}

const getWorkerState = `-- name: GetWorkerState :one

SELECT id,
       worker_id,
       worker_type,
       status,
       last_check_time,
       next_check_time,
       error,
       shows_checked,
       updates_found,
       created_at,
       updated_at
FROM worker_states
WHERE worker_id = $1
ORDER BY updated_at DESC LIMIT 1
`

// Workers Related
func (q *Queries) GetWorkerState(ctx context.Context, workerID string) (WorkerState, error) {
	row := q.db.QueryRow(ctx, getWorkerState, workerID)
	var i WorkerState
	err := row.Scan(
		&i.ID,
		&i.WorkerID,
		&i.WorkerType,
		&i.Status,
		&i.LastCheckTime,
		&i.NextCheckTime,
		&i.Error,
		&i.ShowsChecked,
		&i.UpdatesFound,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getWorkerTask = `-- name: GetWorkerTask :one
SELECT id,
       worker_id,
       task_type,
       status,
       start_time,
       end_time,
       duration_ms,
       error,
       show_id,
       user_id,
       shows_checked,
       updates_found,
       created_at
FROM worker_tasks
WHERE id = $1
`

func (q *Queries) GetWorkerTask(ctx context.Context, id uuid.UUID) (WorkerTask, error) {
	row := q.db.QueryRow(ctx, getWorkerTask, id)
	var i WorkerTask
	err := row.Scan(
		&i.ID,
		&i.WorkerID,
		&i.TaskType,
		&i.Status,
		&i.StartTime,
		&i.EndTime,
		&i.DurationMs,
		&i.Error,
		&i.ShowID,
		&i.UserID,
		&i.ShowsChecked,
		&i.UpdatesFound,
		&i.CreatedAt,
	)
	return i, err
}

const movieExists = `-- name: MovieExists :one
SELECT EXISTS(SELECT 1 FROM movies WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL)
`

type MovieExistsParams struct {
	ApiID  int64 `json:"api_id"`
	UserID int64 `json:"user_id"`
}

func (q *Queries) MovieExists(ctx context.Context, arg MovieExistsParams) (bool, error) {
	row := q.db.QueryRow(ctx, movieExists, arg.ApiID, arg.UserID)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const softDeleteMovie = `-- name: SoftDeleteMovie :exec
UPDATE movies
SET deleted_at = NOW()
WHERE api_id = $1
  AND user_id = $2
  AND deleted_at IS NULL
`

type SoftDeleteMovieParams struct {
	ApiID  int64 `json:"api_id"`
	UserID int64 `json:"user_id"`
}

func (q *Queries) SoftDeleteMovie(ctx context.Context, arg SoftDeleteMovieParams) error {
	_, err := q.db.Exec(ctx, softDeleteMovie, arg.ApiID, arg.UserID)
	return err
}

const softDeleteTVShow = `-- name: SoftDeleteTVShow :exec
UPDATE tv_shows
SET deleted_at = NOW()
WHERE api_id = $1
  AND user_id = $2
  AND deleted_at IS NULL
`

type SoftDeleteTVShowParams struct {
	ApiID  int64 `json:"api_id"`
	UserID int64 `json:"user_id"`
}

func (q *Queries) SoftDeleteTVShow(ctx context.Context, arg SoftDeleteTVShowParams) error {
	_, err := q.db.Exec(ctx, softDeleteTVShow, arg.ApiID, arg.UserID)
	return err
}

const tVShowExists = `-- name: TVShowExists :one
SELECT EXISTS(SELECT 1 FROM tv_shows WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL)
`

type TVShowExistsParams struct {
	ApiID  int64 `json:"api_id"`
	UserID int64 `json:"user_id"`
}

func (q *Queries) TVShowExists(ctx context.Context, arg TVShowExistsParams) (bool, error) {
	row := q.db.QueryRow(ctx, tVShowExists, arg.ApiID, arg.UserID)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const updateMovie = `-- name: UpdateMovie :exec
UPDATE movies
SET runtime = $3,
    title   = $4
WHERE api_id = $1
  AND user_id = $2
  AND deleted_at IS NULL
`

type UpdateMovieParams struct {
	ApiID   int64  `json:"api_id"`
	UserID  int64  `json:"user_id"`
	Runtime int32  `json:"runtime"`
	Title   string `json:"title"`
}

func (q *Queries) UpdateMovie(ctx context.Context, arg UpdateMovieParams) error {
	_, err := q.db.Exec(ctx, updateMovie,
		arg.ApiID,
		arg.UserID,
		arg.Runtime,
		arg.Title,
	)
	return err
}

const updateTVShow = `-- name: UpdateTVShow :exec
UPDATE tv_shows
SET seasons  = $3,
    episodes = $4,
    runtime  = $5
WHERE api_id = $1
  AND user_id = $2
  AND deleted_at IS NULL
`

type UpdateTVShowParams struct {
	ApiID    int64 `json:"api_id"`
	UserID   int64 `json:"user_id"`
	Seasons  int32 `json:"seasons"`
	Episodes int32 `json:"episodes"`
	Runtime  int32 `json:"runtime"`
}

func (q *Queries) UpdateTVShow(ctx context.Context, arg UpdateTVShowParams) error {
	_, err := q.db.Exec(ctx, updateTVShow,
		arg.ApiID,
		arg.UserID,
		arg.Seasons,
		arg.Episodes,
		arg.Runtime,
	)
	return err
}

const updateUserTMDBKey = `-- name: UpdateUserTMDBKey :exec
UPDATE users
SET tmdb_api_key = $2
WHERE tg_id = $1
`

type UpdateUserTMDBKeyParams struct {
	TgID       int64   `json:"tg_id"`
	TmdbApiKey *string `json:"tmdb_api_key"`
}

func (q *Queries) UpdateUserTMDBKey(ctx context.Context, arg UpdateUserTMDBKeyParams) error {
	_, err := q.db.Exec(ctx, updateUserTMDBKey, arg.TgID, arg.TmdbApiKey)
	return err
}

const updateWorkerTask = `-- name: UpdateWorkerTask :exec
UPDATE worker_tasks
SET status        = $2,
    end_time      = $3,
    duration_ms   = $4,
    error         = $5,
    shows_checked = $6,
    updates_found = $7
WHERE id = $1
`

type UpdateWorkerTaskParams struct {
	ID           uuid.UUID          `json:"id"`
	Status       string             `json:"status"`
	EndTime      pgtype.Timestamptz `json:"end_time"`
	DurationMs   *int64             `json:"duration_ms"`
	Error        *string            `json:"error"`
	ShowsChecked *int32             `json:"shows_checked"`
	UpdatesFound *int32             `json:"updates_found"`
}

func (q *Queries) UpdateWorkerTask(ctx context.Context, arg UpdateWorkerTaskParams) error {
	_, err := q.db.Exec(ctx, updateWorkerTask,
		arg.ID,
		arg.Status,
		arg.EndTime,
		arg.DurationMs,
		arg.Error,
		arg.ShowsChecked,
		arg.UpdatesFound,
	)
	return err
}

const upsertWorkerState = `-- name: UpsertWorkerState :one
INSERT INTO worker_states (worker_id, worker_type, status, last_check_time, next_check_time,
                           error, shows_checked, updates_found, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (worker_id) DO
UPDATE SET
    status = EXCLUDED.status,last_check_time = EXCLUDED.last_check_time,
    next_check_time = EXCLUDED.next_check_time,
    error = EXCLUDED.error,
    shows_checked = EXCLUDED.shows_checked,
    updates_found = EXCLUDED.updates_found,
    updated_at = EXCLUDED.updated_at
    RETURNING id
`

type UpsertWorkerStateParams struct {
	WorkerID      string             `json:"worker_id"`
	WorkerType    string             `json:"worker_type"`
	Status        string             `json:"status"`
	LastCheckTime pgtype.Timestamptz `json:"last_check_time"`
	NextCheckTime pgtype.Timestamptz `json:"next_check_time"`
	Error         *string            `json:"error"`
	ShowsChecked  int32              `json:"shows_checked"`
	UpdatesFound  int32              `json:"updates_found"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
}

func (q *Queries) UpsertWorkerState(ctx context.Context, arg UpsertWorkerStateParams) (uuid.UUID, error) {
	row := q.db.QueryRow(ctx, upsertWorkerState,
		arg.WorkerID,
		arg.WorkerType,
		arg.Status,
		arg.LastCheckTime,
		arg.NextCheckTime,
		arg.Error,
		arg.ShowsChecked,
		arg.UpdatesFound,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	var id uuid.UUID
	err := row.Scan(&id)
	return id, err
}

const userExists = `-- name: UserExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE tg_id = $1)
`

func (q *Queries) UserExists(ctx context.Context, tgID int64) (bool, error) {
	row := q.db.QueryRow(ctx, userExists, tgID)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const watchlistExists = `-- name: WatchlistExists :one
SELECT EXISTS(SELECT 1 FROM watchlists WHERE show_api_id = $1 AND user_id = $2 AND type = $3 AND deleted_at IS NULL)
`

type WatchlistExistsParams struct {
	ShowApiID int64  `json:"show_api_id"`
	UserID    int64  `json:"user_id"`
	Type      string `json:"type"`
}

func (q *Queries) WatchlistExists(ctx context.Context, arg WatchlistExistsParams) (bool, error) {
	row := q.db.QueryRow(ctx, watchlistExists, arg.ShowApiID, arg.UserID, arg.Type)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}
