CREATE OR REPLACE VIEW vw_library_patrons AS
SELECT id, full_name, barcode, created_at
FROM patrons
WHERE deleted_at IS NULL;

CREATE OR REPLACE VIEW vw_deleted_play_to_win_games AS
SELECT  id, ptw_group_id, game_id, deleted_at, deletion_reason, deletion_reason_comment, winner_id
FROM play_to_win_games
WHERE deleted_at IS NOT NULL;

CREATE OR REPLACE VIEW vw_deleted_play_to_win_game_overview AS
SELECT
    ptw_game.id AS ptw_game_id,
    ptw_game.game_id AS game_id,
    ptw_game.ptw_group_id AS ptw_group_id,
    COALESCE(g.display_title, g.title, 'Missing Game') AS game_title,
    COALESCE(g.sanitized_title, 'missing game') AS sanitized_title,
    ptw_game.deletion_reason AS deletion_reason,
    ptw_game.deletion_reason_comment AS deletion_reason_comment,
    ptw_game.deleted_at AS deleted_at,
    ptw_entry.id AS winner_id,
    ptw_entry.entrant_name AS winner_name,
    ptw_entry.entrant_unique_id AS winner_unique_id
FROM vw_deleted_play_to_win_games AS ptw_game
         LEFT JOIN games AS g ON g.id = ptw_game.game_id
         LEFT JOIN play_to_win_entries AS ptw_entry ON ptw_game.winner_id = ptw_entry.id;

CREATE OR REPLACE VIEW vw_play_to_win_groups AS
SELECT id, name, created_at
FROM play_to_win_groups
WHERE deleted_at IS NULL;

CREATE OR REPLACE VIEW vw_play_to_win_sessions AS
SELECT  id, ptw_group_id, playtime_minutes, created_at
FROM play_to_win_sessions
WHERE deleted_at IS NULL;

CREATE OR REPLACE VIEW vw_deleted_play_to_win_sessions AS
SELECT id, ptw_group_id, deleted_at, deletion_reason, deletion_reason_comment
FROM play_to_win_sessions
WHERE deleted_at IS NOT NULL;

CREATE OR REPLACE VIEW vw_play_to_win_entries AS
SELECT ptw_entries.id,
       ptw_entries.ptw_session_id,
       ptw_sessions.ptw_group_id,
       ptw_entries.entrant_name,
       ptw_entries.entrant_unique_id,
       ptw_entries.created_at
FROM play_to_win_entries ptw_entries
         LEFT JOIN vw_play_to_win_sessions ptw_sessions ON ptw_sessions.id = ptw_entries.ptw_session_id
WHERE ptw_entries.deleted_at IS NULL;

CREATE OR REPLACE VIEW vw_play_to_win_game_overview AS
SELECT
    ptw_game.id AS ptw_game_id,
    ptw_game.game_id AS game_id,
    ptw_game.ptw_group_id AS ptw_group_id,
    COALESCE(g.display_title, 'Missing Game') AS game_title,
    COALESCE(g.sanitized_title, 'missing game') AS sanitized_title,
    ptw_game.created_at AS created_at,
    ptw_entry.id AS winner_id,
    ptw_entry.entrant_name AS winner_name,
    ptw_entry.entrant_unique_id AS winner_unique_id
FROM vw_play_to_win_games AS ptw_game
         LEFT JOIN vw_library_games AS g ON g.id = ptw_game.game_id
         LEFT JOIN vw_play_to_win_entries AS ptw_entry ON ptw_game.winner_id = ptw_entry.id;

CREATE OR REPLACE VIEW vw_play_to_win_entry_events AS
SELECT ptw.id as entry_id,
       ptw.ptw_session_id,
       ps.ptw_group_id,
       ptw.entrant_name,
       ptw.entrant_unique_id,
       ptw.deletion_reason,
       ps.playtime_minutes,
       pg.deletion_reason AS game_deletion_reason,
       g.title AS game_title,
       ptw.created_at
FROM play_to_win_entries ptw
         LEFT JOIN play_to_win_sessions ps ON ps.id = ptw.ptw_session_id
         LEFT JOIN play_to_win_games pg ON pg.id = ps.ptw_group_id
         LEFT JOIN games g ON g.id = pg.game_id
WHERE (ptw.deletion_reason IN ('won','failed_to_claim') OR ptw.deletion_reason IS NULL)
  AND ps.deletion_reason IS NULL
ORDER BY ptw.created_at DESC;

CREATE OR REPLACE VIEW vw_deleted_play_to_win_entries AS
SELECT id, ptw_session_id, deleted_at, deletion_reason, deletion_reason_comment
FROM play_to_win_entries
WHERE deleted_at IS NOT NULL;

CREATE OR REPLACE VIEW vw_library_transaction_events AS
SELECT
    te.transaction_id,
    g.id AS game_id,
    COALESCE(g.title, 'Missing Game') AS game_title,
    g.sanitized_title AS sanitized_title,
    te.patron_id,
    COALESCE(p.full_name, 'Missing Patron') AS patron_full_name,
    te.event_timestamp,
    te.event_type,
    ptw.id AS play_to_win_game_id
FROM transaction_events AS te
         LEFT JOIN games AS g ON te.game_id = g.id
         LEFT JOIN patrons AS p ON te.patron_id = p.id
         LEFT JOIN vw_play_to_win_games AS ptw ON te.game_id = ptw.game_id
ORDER BY te.event_timestamp DESC;

CREATE OR REPLACE VIEW vw_game_status AS
SELECT DISTINCT ON (g.id)
    g.id AS game_id,
    g.display_title AS game_title,
    g.barcode AS game_barcode,
    g.sanitized_title,
    t.patron_id,
    p.full_name AS patron_full_name,
    p.barcode AS patron_barcode,
    t.id AS transaction_id,
    t.checkout_timestamp,
    t.checkin_timestamp,
    (checkout_timestamp IS NOT NULL AND checkin_timestamp IS NULL) AS checked_out,
    ptw.id AS ptw_game_id
FROM vw_library_games AS g
         LEFT JOIN transactions AS t ON t.game_id = g.id
         LEFT JOIN vw_library_patrons AS p ON t.patron_id = p.id
         LEFT JOIN vw_play_to_win_games AS ptw ON ptw.game_id = g.id
ORDER BY g.id, t.checkout_timestamp DESC;