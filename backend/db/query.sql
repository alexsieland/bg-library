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

-- name: GetGameByBarcode :many
SELECT *
FROM vw_library_games
WHERE barcode = $1;

-- name: CreateGame :one
INSERT INTO games ( title, sanitized_title, barcode ) VALUES ( $1, $2, $3 )
RETURNING *;

-- name: EditGame :exec
UPDATE games
SET title = $2,
    sanitized_title = $3,
    barcode = $4
WHERE id = $1;

-- name: DeleteGame :exec
UPDATE games
SET deleted_at = now(),
    barcode = NULL
WHERE deleted_at IS NULL AND id = $1;

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
INSERT INTO patrons ( full_name, barcode ) VALUES ( $1, $2 )
RETURNING *;

-- name: EditPatron :exec
UPDATE patrons
SET full_name = $2,
    barcode = $3
WHERE id = $1;

-- name: DeletePatron :exec
UPDATE patrons
SET deleted_at = now(),
    barcode = NULL,
    full_name = 'Deleted Patron'
WHERE deleted_at IS NULL AND id = $1;

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

-- name: SearchTransactionEvents :many
SELECT transaction_id, game_id, game_title, patron_id, patron_full_name, event_type, event_timestamp, play_to_win_game_id
FROM vw_library_transaction_events
WHERE sanitized_title ILIKE $1 AND patron_full_name ILIKE $2
LIMIT $3 OFFSET $4;

-- name: ListPlayToWinGames :many
SELECT *
FROM vw_play_to_win_game_overview
WHERE sanitized_title ILIKE $1
LIMIT $2 OFFSET $3;

-- name: GetPlayToWinGame :one
SELECT *
FROM vw_play_to_win_game_overview
WHERE play_to_win_id = $1;

-- name: GetPlayToWinSessions :many
SELECT
    id AS session_id,
    play_to_win_id,
    playtime_minutes,
    created_at
FROM vw_play_to_win_sessions
WHERE play_to_win_id = $1;

-- name: GetPlayToWinEntries :many
SELECT
    id AS entry_id,
    session_id,
    play_to_win_id,
    entrant_name,
    entrant_unique_id,
    created_at
FROM vw_play_to_win_entries
WHERE play_to_win_id = $1;

-- name: CreatePlayToWinGame :one
INSERT INTO play_to_win_games (game_id) VALUES ($1)
RETURNING *;

-- name: CreatePlayToWinSession :one
INSERT INTO play_to_win_sessions (play_to_win_id, playtime_minutes) VALUES ($1, $2)
RETURNING *;

-- name: CreatePlayToWinEntry :one
INSERT INTO play_to_win_entries (session_id, entrant_name, entrant_unique_id) VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdatePlayToWinEntry :exec
UPDATE play_to_win_games
SET winner_id = $2
WHERE id = $1;

-- name: DeletePlayToWinGame :exec
UPDATE play_to_win_games
SET deleted_at = now(),
    deletion_reason = $2,
    deletion_reason_comment = $3
WHERE deleted_at IS NULL AND game_id = $1;

-- name: RestorePlayToWinGame :exec
UPDATE play_to_win_games
SET deleted_at = NULL,
    deletion_reason = NULL,
    deletion_reason_comment = NULL
WHERE deleted_at IS NOT NULL AND id = $1;

-- name: DeletePlayToWinSession :exec
UPDATE play_to_win_sessions
SET deleted_at = now(),
    deletion_reason = $2,
    deletion_reason_comment = $3
WHERE deleted_at IS NULL AND id = $1;

-- name: RestorePlayToWinSession :exec
UPDATE play_to_win_sessions
SET deleted_at = NULL,
    deletion_reason = NULL,
    deletion_reason_comment = NULL
WHERE deleted_at IS NOT NULL AND id = $1;

-- name: DeletePlayToWinEntry :exec
UPDATE play_to_win_entries
SET deleted_at = now(),
    deletion_reason = $2,
    deletion_reason_comment = $3
WHERE deleted_at IS NULL AND id = $1;

-- name: RestorePlayToWinEntry :exec
UPDATE play_to_win_entries
SET deleted_at = NULL,
    deletion_reason = NULL,
    deletion_reason_comment = NULL
WHERE id = $1;
