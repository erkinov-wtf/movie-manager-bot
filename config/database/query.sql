/* Users Table */

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: CountUsers :one
SELECT count(*) FROM users
WHERE id = $1;

-- name: CreateUser :exec
INSERT INTO users (first_name, last_name, username, language, tmdb_api_key)
VALUES ($1, $2, $3, $4, $4);

-- name: UpdateUserTMDBKey :exec
UPDATE users
SET tmdb_api_key = $2
WHERE id = $1;


/* TV Shows Table */

-- name: GetUserTVShows :many
SELECT * FROM tv_shows
WHERE user_id = $1;

-- name: GetWatchedSeasons :one
SELECT seasons FROM tv_shows
WHERE api_id = $1 AND user_id = $2;

-- name: CreateTVShow :exec
INSERT INTO tv_shows (user_id, api_id, name, seasons, episodes, runtime, status)
VALUES ($1, $2, $3, $4, $5,$6, $7);

-- name: UpdateTVShow :exec
UPDATE tv_shows
SET (seasons = $2, episodes = $3, runtime = $4)
WHERE api_id = $1;


/* Movies Table */

-- name: GetUserMovies :many
SELECT * FROM movies
WHERE user_id = $1;

-- name: CountUserMovies :one
SELECT count(*) FROM movies
WHERE api_id = $1 AND user_id = $2;

-- name: CreateMovie :exec
INSERT INTO movies (user_id, api_id, title, runtime, status)
VALUES ($1, $2, $3, $4, $5);


/* Watchlists Table */

-- name: CreateWatchlist :exec
INSERT INTO watchlists (user_id, show_api_id, type, title, image)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserWatchlists :many
SELECT * FROM watchlists
WHERE user_id = $1;

-- name: GetUserWatchlistsWithType :many
SELECT * FROM watchlists
WHERE user_id = $1 AND type = $2;

-- name: DeleteWatchlist :exec
DELETE FROM watchlists
WHERE show_api_id = $1 AND user_id = $2;