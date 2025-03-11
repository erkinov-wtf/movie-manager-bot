// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createMovie = `-- name: CreateMovie :exec
INSERT INTO movies (user_id, api_id, title, runtime)
VALUES ($1, $2, $3, $4)
`

type CreateMovieParams struct {
	UserID  int64
	ApiID   int64
	Title   string
	Runtime *int32
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
	UserID   int64
	ApiID    int64
	Name     string
	Seasons  int32
	Episodes int32
	Runtime  int32
	Status   string
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
INSERT INTO users (first_name, last_name, username, language, tmdb_api_key)
VALUES ($1, $2, $3, $4, $5)
`

type CreateUserParams struct {
	FirstName  *string
	LastName   *string
	Username   *string
	Language   string
	TmdbApiKey *string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) error {
	_, err := q.db.Exec(ctx, createUser,
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
	UserID    int64
	ShowApiID int64
	Type      string
	Title     string
	Image     *string
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

const deleteWatchlist = `-- name: DeleteWatchlist :exec
UPDATE watchlists
SET deleted_at = NOW()
WHERE show_api_id = $1 AND user_id = $2 AND deleted_at IS NULL
`

type DeleteWatchlistParams struct {
	ShowApiID int64
	UserID    int64
}

func (q *Queries) DeleteWatchlist(ctx context.Context, arg DeleteWatchlistParams) error {
	_, err := q.db.Exec(ctx, deleteWatchlist, arg.ShowApiID, arg.UserID)
	return err
}

const getUser = `-- name: GetUser :one

SELECT id, first_name, last_name, username, language, tmdb_api_key, created_at, updated_at
FROM users
WHERE id = $1 LIMIT 1
`

// Users Table
func (q *Queries) GetUser(ctx context.Context, id int64) (User, error) {
	row := q.db.QueryRow(ctx, getUser, id)
	var i User
	err := row.Scan(
		&i.ID,
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
WHERE user_id = $1 AND deleted_at IS NULL
`

type GetUserMoviesRow struct {
	ID        pgtype.UUID
	ApiID     int64
	Title     string
	Runtime   *int32
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
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
SELECT tmdb_api_key FROM users
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetUserTMDBKey(ctx context.Context, id int64) (*string, error) {
	row := q.db.QueryRow(ctx, getUserTMDBKey, id)
	var tmdb_api_key *string
	err := row.Scan(&tmdb_api_key)
	return tmdb_api_key, err
}

const getUserTVShows = `-- name: GetUserTVShows :many

SELECT id, api_id, name, seasons, episodes, runtime, status, created_at, updated_at
FROM tv_shows
WHERE user_id = $1 AND deleted_at IS NULL
`

type GetUserTVShowsRow struct {
	ID        pgtype.UUID
	ApiID     int64
	Name      string
	Seasons   int32
	Episodes  int32
	Runtime   int32
	Status    string
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
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
SELECT id, user_id, show_api_id, type, title, image, created_at, updated_at
FROM watchlists
WHERE show_api_id = $1 AND user_id = $2 AND deleted_at IS NULL
    LIMIT 1
`

type GetUserWatchlistParams struct {
	ShowApiID int64
	UserID    int64
}

type GetUserWatchlistRow struct {
	ID        pgtype.UUID
	UserID    int64
	ShowApiID int64
	Type      string
	Title     string
	Image     *string
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
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
WHERE user_id = $1 AND deleted_at IS NULL
`

type GetUserWatchlistsRow struct {
	ID        pgtype.UUID
	ShowApiID int64
	Type      string
	Title     string
	Image     *string
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
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
WHERE user_id = $1 AND type = $2 AND deleted_at IS NULL
`

type GetUserWatchlistsWithTypeParams struct {
	UserID int64
	Type   string
}

type GetUserWatchlistsWithTypeRow struct {
	ID        pgtype.UUID
	ShowApiID int64
	Type      string
	Title     string
	Image     *string
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
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
SELECT id, first_name, last_name, username, language, created_at, updated_at
FROM users
`

type GetUsersRow struct {
	ID        int64
	FirstName *string
	LastName  *string
	Username  *string
	Language  string
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
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
SELECT seasons FROM tv_shows
WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL
`

type GetWatchedSeasonsParams struct {
	ApiID  int64
	UserID int64
}

func (q *Queries) GetWatchedSeasons(ctx context.Context, arg GetWatchedSeasonsParams) (int32, error) {
	row := q.db.QueryRow(ctx, getWatchedSeasons, arg.ApiID, arg.UserID)
	var seasons int32
	err := row.Scan(&seasons)
	return seasons, err
}

const movieExists = `-- name: MovieExists :one
SELECT EXISTS(SELECT 1 FROM movies WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL)
`

type MovieExistsParams struct {
	ApiID  int64
	UserID int64
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
WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL
`

type SoftDeleteMovieParams struct {
	ApiID  int64
	UserID int64
}

func (q *Queries) SoftDeleteMovie(ctx context.Context, arg SoftDeleteMovieParams) error {
	_, err := q.db.Exec(ctx, softDeleteMovie, arg.ApiID, arg.UserID)
	return err
}

const softDeleteTVShow = `-- name: SoftDeleteTVShow :exec
UPDATE tv_shows
SET deleted_at = NOW()
WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL
`

type SoftDeleteTVShowParams struct {
	ApiID  int64
	UserID int64
}

func (q *Queries) SoftDeleteTVShow(ctx context.Context, arg SoftDeleteTVShowParams) error {
	_, err := q.db.Exec(ctx, softDeleteTVShow, arg.ApiID, arg.UserID)
	return err
}

const tVShowExists = `-- name: TVShowExists :one
SELECT EXISTS(SELECT 1 FROM tv_shows WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL)
`

type TVShowExistsParams struct {
	ApiID  int64
	UserID int64
}

func (q *Queries) TVShowExists(ctx context.Context, arg TVShowExistsParams) (bool, error) {
	row := q.db.QueryRow(ctx, tVShowExists, arg.ApiID, arg.UserID)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const updateMovie = `-- name: UpdateMovie :exec
UPDATE movies
SET runtime = $3, title = $4
WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL
`

type UpdateMovieParams struct {
	ApiID   int64
	UserID  int64
	Runtime *int32
	Title   string
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
SET seasons = $3, episodes = $4, runtime = $5
WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL
`

type UpdateTVShowParams struct {
	ApiID    int64
	UserID   int64
	Seasons  int32
	Episodes int32
	Runtime  int32
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
WHERE id = $1
`

type UpdateUserTMDBKeyParams struct {
	ID         int64
	TmdbApiKey *string
}

func (q *Queries) UpdateUserTMDBKey(ctx context.Context, arg UpdateUserTMDBKeyParams) error {
	_, err := q.db.Exec(ctx, updateUserTMDBKey, arg.ID, arg.TmdbApiKey)
	return err
}

const userExists = `-- name: UserExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
`

func (q *Queries) UserExists(ctx context.Context, id int64) (bool, error) {
	row := q.db.QueryRow(ctx, userExists, id)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const watchlistExists = `-- name: WatchlistExists :one
SELECT EXISTS(SELECT 1 FROM watchlists WHERE show_api_id = $1 AND user_id = $2 AND type = $3 AND deleted_at IS NULL)
`

type WatchlistExistsParams struct {
	ShowApiID int64
	UserID    int64
	Type      string
}

func (q *Queries) WatchlistExists(ctx context.Context, arg WatchlistExistsParams) (bool, error) {
	row := q.db.QueryRow(ctx, watchlistExists, arg.ShowApiID, arg.UserID, arg.Type)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}
