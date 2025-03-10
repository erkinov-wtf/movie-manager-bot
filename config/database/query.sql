/* Users Table */

-- name: GetUser :one
SELECT id, first_name, last_name, username, language, tmdb_api_key, created_at, updated_at
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUsers :many
SELECT id, first_name, last_name, username, language, created_at, updated_at
FROM users;

-- name: UserExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE id = $1);

-- name: CreateUser :exec
INSERT INTO users (first_name, last_name, username, language, tmdb_api_key)
VALUES ($1, $2, $3, $4, $5);

-- name: UpdateUserTMDBKey :exec
UPDATE users
SET tmdb_api_key = $2
WHERE id = $1;

-- name: GetUserTMDBKey :one
SELECT tmdb_api_key FROM users
WHERE id = $1 LIMIT 1;


/* TV Shows Table */

-- name: GetUserTVShows :many
SELECT id, api_id, name, seasons, episodes, runtime, status, created_at, updated_at
FROM tv_shows
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: GetWatchedSeasons :one
SELECT seasons FROM tv_shows
WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: TVShowExists :one
SELECT EXISTS(SELECT 1 FROM tv_shows WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL);

-- name: CreateTVShow :exec
INSERT INTO tv_shows (user_id, api_id, name, seasons, episodes, runtime, status)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: UpdateTVShow :exec
UPDATE tv_shows
SET seasons = $3, episodes = $4, runtime = $5
WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: SoftDeleteTVShow :exec
UPDATE tv_shows
SET deleted_at = NOW()
WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL;


/* Movies Table */

-- name: GetUserMovies :many
SELECT id, api_id, title, runtime, created_at, updated_at
FROM movies
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: MovieExists :one
SELECT EXISTS(SELECT 1 FROM movies WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL);

-- name: CreateMovie :exec
INSERT INTO movies (user_id, api_id, title, runtime)
VALUES ($1, $2, $3, $4);

-- name: UpdateMovie :exec
UPDATE movies
SET runtime = $3, title = $4
WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: SoftDeleteMovie :exec
UPDATE movies
SET deleted_at = NOW()
WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL;


/* Watchlists Table */

-- name: CreateWatchlist :exec
INSERT INTO watchlists (user_id, show_api_id, type, title, image)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserWatchlist :one
SELECT id, user_id, show_api_id, type, title, image, created_at, updated_at
FROM watchlists
WHERE show_api_id = $1 AND user_id = $2 AND deleted_at IS NULL
    LIMIT 1;

-- name: WatchlistExists :one
SELECT EXISTS(SELECT 1 FROM watchlists WHERE show_api_id = $1 AND user_id = $2 AND type = $3 AND deleted_at IS NULL);

-- name: GetUserWatchlists :many
SELECT id, show_api_id, type, title, image, created_at, updated_at
FROM watchlists
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: GetUserWatchlistsWithType :many
SELECT id, show_api_id, type, title, image, created_at, updated_at
FROM watchlists
WHERE user_id = $1 AND type = $2 AND deleted_at IS NULL;

-- name: DeleteWatchlist :exec
UPDATE watchlists
SET deleted_at = NOW()
WHERE show_api_id = $1 AND user_id = $2 AND deleted_at IS NULL;