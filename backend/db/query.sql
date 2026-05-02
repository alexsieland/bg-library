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
SET display_title = $2,
    sanitized_title = $3,
    barcode = $4
WHERE id = $1;

-- name: DeleteGame :exec
UPDATE games
SET deleted_at = now()
WHERE deleted_at IS NULL AND id = $1;

-- name: RestoreGame :exec
UPDATE games
SET deleted_at = NULL
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

-- name: CountActiveCheckoutsByPatron :one
SELECT count(*) FROM transactions WHERE patron_id = $1 AND checkin_timestamp IS NULL;

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
WHERE (sqlc.narg('checked_out')::boolean IS NULL OR checked_out = sqlc.narg('checked_out'))
  AND (sqlc.narg('sanitized_title')::text IS NULL OR sanitized_title ILIKE sqlc.narg('sanitized_title'))
  AND (sqlc.narg('game_barcode')::text IS NULL OR game_barcode ILIKE sqlc.narg('game_barcode'))
ORDER BY sanitized_title
LIMIT $1 OFFSET $2;

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

-- name: SearchTransactionEvents :many
SELECT transaction_id, game_id, game_title, patron_id, patron_full_name, event_type, event_timestamp, play_to_win_game_id
FROM vw_library_transaction_events
WHERE sanitized_title ILIKE $1 AND patron_full_name ILIKE $2
LIMIT $3 OFFSET $4;

-- name: GetPlayToWinGameOverview :one
SELECT *
FROM vw_play_to_win_game_overview
WHERE ptw_game_id = $1;

-- name: GetDeletedPlayToWinGameOverview :one
SELECT *
FROM vw_deleted_play_to_win_game_overview
WHERE ptw_game_id = $1;

-- name: ListPlayToWinGameOverviews :many
SELECT *
FROM vw_play_to_win_game_overview
LIMIT $1 OFFSET $2;

-- name: SearchPlayToWinGameOverviews :many
SELECT *
FROM vw_play_to_win_game_overview
WHERE sanitized_title ILIKE $1
LIMIT $2 OFFSET $3;

-- name: GetPlayToWinGameByLibraryGameId :one
SELECT *
FROM vw_play_to_win_games
WHERE game_id = $1;

-- name: GetPlayToWinGroup :one
SELECT *
FROM vw_play_to_win_groups
WHERE id = $1;

-- name: GetPlayToWinGroupByPlayToWinGameId :one
SELECT gr.*
FROM vw_play_to_win_groups AS gr
LEFT JOIN play_to_win_games AS ga ON gr.id = ga.ptw_group_id
WHERE ga.id = $1;

-- name: GetPlayToWinGroupByName :one
SELECT *
FROM vw_play_to_win_groups
WHERE name ILIKE $1;

-- name: CreatePlayToWinGroup :one
INSERT INTO play_to_win_groups (name) VALUES ($1)
RETURNING *;

-- name: GetPlayToWinSessionsByGroupId :many
SELECT *
FROM vw_play_to_win_sessions
WHERE ptw_group_id = $1;

-- name: GetPlayToWinEntriesByGroupId :many
SELECT *
FROM vw_play_to_win_entries
WHERE ptw_group_id = $1;

-- name: GetPlayToWinEntriesByPlayToWinGameId :many
SELECT e.*
FROM vw_play_to_win_entries e
LEFT JOIN play_to_win_games g ON e.ptw_group_id = g.ptw_group_id
WHERE g.id = $1;

-- name: CreatePlayToWinGame :one
INSERT INTO play_to_win_games (game_id, ptw_group_id) VALUES ($1, $2)
RETURNING *;

-- name: CreatePlayToWinSession :one
INSERT INTO play_to_win_sessions (ptw_group_id, playtime_minutes) VALUES ($1, $2)
RETURNING *;

-- name: CreatePlayToWinEntry :one
INSERT INTO play_to_win_entries (ptw_session_id, ptw_group_id, entrant_name, entrant_unique_id) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdatePlayToWinWinner :exec
UPDATE play_to_win_games
SET winner_id = $2
WHERE id = $1;

-- name: DeletePlayToWinGame :exec
UPDATE play_to_win_games
SET deleted_at = now(),
    deletion_reason = $2,
    deletion_reason_comment = $3
WHERE deleted_at IS NULL AND game_id = $1;


-- name: DeletePlayToWinGameByPlayToWinId :exec
UPDATE play_to_win_games
SET deleted_at = now(),
    deletion_reason = $2,
    deletion_reason_comment = $3
WHERE deleted_at IS NULL AND id = $1;

-- name: RestorePlayToWinGame :exec
UPDATE play_to_win_games
SET deleted_at = NULL,
    deletion_reason = NULL,
    deletion_reason_comment = NULL
WHERE deleted_at IS NOT NULL AND id = $1;

-- name: RestorePlayToWinGameByLibraryGameId :exec
UPDATE play_to_win_games
SET deleted_at = NULL,
    deletion_reason = NULL,
    deletion_reason_comment = NULL
WHERE deleted_at IS NOT NULL AND game_id = $1;

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

-- name: ResetPlayToWinGameWinners :exec
UPDATE play_to_win_games
SET winner_id = NULL
WHERE deletion_reason IS DISTINCT FROM 'claimed';

-- name: ListDeletedPlayToWinGameOverviews :many
SELECT *
FROM vw_deleted_play_to_win_game_overview
WHERE deletion_reason = $1
  AND (sqlc.narg('sanitized_title')::text IS NULL OR sanitized_title ILIKE sqlc.narg('sanitized_title'))
ORDER BY deleted_at DESC, sanitized_title ASC
LIMIT $2 OFFSET $3;

