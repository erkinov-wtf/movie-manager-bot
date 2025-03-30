/* Users Table */

-- name: GetUser :one
SELECT id, tg_id, first_name, last_name, username, language, tmdb_api_key, created_at, updated_at
FROM users
WHERE tg_id = $1 LIMIT 1;

-- name: GetUsers :many
SELECT id, tg_id, first_name, last_name, username, language, created_at, updated_at
FROM users;

-- name: UserExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE tg_id = $1);

-- name: CreateUser :exec
INSERT INTO users (tg_id, first_name, last_name, username, language, tmdb_api_key)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdateUserTMDBKey :exec
UPDATE users
SET tmdb_api_key = $2
WHERE tg_id = $1;

-- name: GetUserTMDBKey :one
SELECT tmdb_api_key
FROM users
WHERE tg_id = $1 LIMIT 1;


/* TV Shows Table */

-- name: GetUserTVShows :many
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
  AND deleted_at IS NULL;

-- name: GetWatchedSeasons :one
SELECT seasons
FROM tv_shows
WHERE api_id = $1
  AND user_id = $2
  AND deleted_at IS NULL;

-- name: GetUserTVShow :one
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
WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: TVShowExists :one
SELECT EXISTS(SELECT 1 FROM tv_shows WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL);

-- name: CreateTVShow :exec
INSERT INTO tv_shows (user_id, api_id, name, seasons, episodes, runtime, status)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: UpdateTVShow :exec
UPDATE tv_shows
SET seasons  = $3,
    episodes = $4,
    runtime  = $5
WHERE api_id = $1
  AND user_id = $2
  AND deleted_at IS NULL;

-- name: SoftDeleteTVShow :exec
UPDATE tv_shows
SET deleted_at = NOW()
WHERE api_id = $1
  AND user_id = $2
  AND deleted_at IS NULL;


/* Movies Table */

-- name: GetUserMovies :many
SELECT id, api_id, title, runtime, created_at, updated_at
FROM movies
WHERE user_id = $1
  AND deleted_at IS NULL;

-- name: MovieExists :one
SELECT EXISTS(SELECT 1 FROM movies WHERE api_id = $1 AND user_id = $2 AND deleted_at IS NULL);

-- name: CreateMovie :exec
INSERT INTO movies (user_id, api_id, title, runtime)
VALUES ($1, $2, $3, $4);

-- name: UpdateMovie :exec
UPDATE movies
SET runtime = $3,
    title   = $4
WHERE api_id = $1
  AND user_id = $2
  AND deleted_at IS NULL;

-- name: SoftDeleteMovie :exec
UPDATE movies
SET deleted_at = NOW()
WHERE api_id = $1
  AND user_id = $2
  AND deleted_at IS NULL;


/* Watchlists Table */

-- name: CreateWatchlist :exec
INSERT INTO watchlists (user_id, show_api_id, type, title, image)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserWatchlist :one
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
  AND deleted_at IS NULL LIMIT 1;

-- name: WatchlistExists :one
SELECT EXISTS(SELECT 1 FROM watchlists WHERE show_api_id = $1 AND user_id = $2 AND type = $3 AND deleted_at IS NULL);

-- name: GetUserWatchlists :many
SELECT id, show_api_id, type, title, image, created_at, updated_at
FROM watchlists
WHERE user_id = $1
  AND deleted_at IS NULL;

-- name: GetUserWatchlistsWithType :many
SELECT id, show_api_id, type, title, image, created_at, updated_at
FROM watchlists
WHERE user_id = $1
  AND type = $2
  AND deleted_at IS NULL;

-- name: DeleteWatchlist :exec
UPDATE watchlists
SET deleted_at = NOW()
WHERE show_api_id = $1
  AND user_id = $2
  AND deleted_at IS NULL;

/* Workers Related */

-- name: GetWorkerState :one
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
ORDER BY updated_at DESC LIMIT 1;

-- name: UpsertWorkerState :one
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
    RETURNING id;

-- name: CreateWorkerTask :one
INSERT INTO worker_tasks (worker_id, task_type, status, start_time, end_time, duration_ms,
                          error, show_id, user_id, shows_checked, updates_found, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id;

-- name: UpdateWorkerTask :exec
UPDATE worker_tasks
SET status        = $2,
    end_time      = $3,
    duration_ms   = $4,
    error         = $5,
    shows_checked = $6,
    updates_found = $7
WHERE id = $1;

-- name: GetRecentTasks :many
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
ORDER BY created_at DESC LIMIT $2;

-- name: GetWorkerTask :one
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
WHERE id = $1;