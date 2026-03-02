-- name: ListGames :many
SELECT *
FROM vw_library_games
ORDER BY sanitized_title
LIMIT $1 OFFSET $2;

-- name: SearchGames :many
SELECT *
FROM vw_library_games
WHERE sanitized_title ILIKE $1
ORDER BY sanitized_title
LIMIT $2 OFFSET $3;

-- name: GetGame :one
SELECT *
FROM vw_library_games
WHERE id = $1;

-- name: GetGameByBarcode :one
SELECT *
FROM vw_library_games
WHERE barcode = $1;

-- name: CreateGame :one
INSERT INTO games ( title, sanitized_title ) VALUES ( $1, $2 )
RETURNING *;

-- name: EditGame :exec
UPDATE games
    SET title = $2,
        sanitized_title = $3
WHERE id = $1;

-- name: DeleteGame :exec
UPDATE games
    SET deleted = TRUE
WHERE id = $1;

-- name: ListPatrons :many
SELECT *
FROM vw_library_patrons
ORDER BY full_name
LIMIT $1 OFFSET $2;

-- name: SearchPatrons :many
SELECT *
FROM vw_library_patrons
WHERE full_name ILIKE $1
ORDER BY full_name
LIMIT $2 OFFSET $3;

-- name: GetPatron :one
SELECT *
FROM vw_library_patrons
WHERE id = $1;

-- name: GetPatronByBarcode :one
SELECT *
FROM vw_library_patrons
WHERE barcode = $1;

-- name: CreatePatron :one
INSERT INTO patrons ( full_name ) VALUES ( $1 )
RETURNING *;

-- name: EditPatron :exec
UPDATE patrons
set full_name = $2
WHERE id = $1;

-- name: DeletePatron :exec
UPDATE patrons
set deleted = TRUE
WHERE id = $1;

-- name: CheckOutGame :one
INSERT INTO transactions (game_id, patron_id) VALUES ($1, $2)
RETURNING *;

-- name: CheckInGame :exec
UPDATE transactions
SET checkin_timestamp = now()
WHERE id = $1 AND checkin_timestamp IS NULL;

-- name: ListGamesStatus :many
SELECT *
FROM vw_game_status
ORDER BY sanitized_title
LIMIT $1 OFFSET $2;

-- name: SearchGameStatus :many
SELECT *
FROM vw_game_status
WHERE sanitized_title ILIKE $1
ORDER BY sanitized_title
LIMIT $2 OFFSET $3;

-- name: GetGameStatus :one
SELECT *
FROM vw_game_status
WHERE game_id = $1;

-- name: ListCheckedOutGames :many
SELECT *
FROM vw_game_status
WHERE checkin_timestamp IS NULL AND checkout_timestamp IS NOT NULL
ORDER BY sanitized_title
LIMIT $1 OFFSET $2;

-- name: SearchCheckedOutGames :many
SELECT *
FROM vw_game_status
WHERE checkin_timestamp IS NULL AND vw_game_status.sanitized_title ILIKE $1
ORDER BY sanitized_title
LIMIT $2 OFFSET $3;
