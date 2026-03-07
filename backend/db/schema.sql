CREATE TABLE games (
    id UUID DEFAULT gen_random_uuid(),
    title VARCHAR(100) NOT NULL,
    sanitized_title VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    barcode VARCHAR(48), -- Not unique because a library might use UPCs
    PRIMARY KEY (id)
);

CREATE TABLE patrons (
    id UUID DEFAULT gen_random_uuid(),
    full_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    barcode VARCHAR(48) UNIQUE,
    PRIMARY KEY (id)
);

CREATE TABLE transactions (
    id UUID DEFAULT gen_random_uuid(),
    game_id UUID REFERENCES games(id),
    patron_id UUID REFERENCES patrons(id),
    checkout_timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    checkin_timestamp TIMESTAMP,
    PRIMARY KEY (id)
);


CREATE TYPE transaction_event_type AS ENUM ('check_out', 'check_in');
CREATE TABLE transaction_events (
    id UUID DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    game_id UUID NOT NULL REFERENCES games(id),
    patron_id UUID NOT NULL REFERENCES patrons(id),
    event_type transaction_event_type NOT NULL,
    event_timestamp TIMESTAMP NOT NULL,
    recorded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);

CREATE TYPE play_to_win_game_deletion_type AS ENUM ('claimed', 'other');
CREATE TABLE play_to_win_games (
    id UUID DEFAULT gen_random_uuid(),
    game_id UUID NOT NULL UNIQUE REFERENCES games(id),
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    deletion_reason play_to_win_game_deletion_type,
    deletion_reason_comment VARCHAR(500),
    PRIMARY KEY (id)
);

CREATE TYPE play_to_win_session_deletion_type AS ENUM ('foul_play', 'too_many_players', 'too_few_players', 'abnormal_playtime', 'other');
CREATE TABLE play_to_win_sessions (
    id UUID DEFAULT gen_random_uuid(),
    play_to_win_id UUID NOT NULL REFERENCES play_to_win_games(id),
    playtime_minutes INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    deletion_reason play_to_win_session_deletion_type,
    deletion_reason_comment VARCHAR(500),
    PRIMARY KEY (id)
);

CREATE TYPE play_to_win_entry_deletion_type AS ENUM ('winner', 'failed_to_claim', 'foul_play', 'duplicate_entrant', 'other');
CREATE TABLE play_to_win_entries (
    id UUID DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES play_to_win_sessions(id),
    entrant_name VARCHAR(100) NOT NULL,
    entrant_unique_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    deletion_reason play_to_win_entry_deletion_type,
    deletion_reason_comment VARCHAR(500),
    PRIMARY KEY (id),
    UNIQUE(session_id, entrant_unique_id)
);

CREATE INDEX idx_game_barcode ON games(barcode);

CREATE INDEX idx_game_titles ON games(sanitized_title);

CREATE INDEX idx_patron_full_name ON patrons(full_name);

CREATE INDEX idx_checkout_timestamp ON transactions(checkout_timestamp);

CREATE INDEX idx_checkin_timestamp ON transactions(checkin_timestamp);

CREATE INDEX idx_transaction_events_timestamp ON transaction_events(event_timestamp);
CREATE INDEX idx_transaction_events_game ON transaction_events(game_id);
CREATE INDEX idx_transaction_events_patron ON transaction_events(patron_id);

CREATE INDEX idx_checked_out_games
ON transactions(checkin_timestamp)
WHERE checkin_timestamp IS NULL;

CREATE INDEX idx_active_games
ON games(deleted_at)
WHERE deleted_at IS NULL;

CREATE INDEX idx_active_patrons
ON patrons(deleted_at)
WHERE deleted_at IS NULL;

CREATE INDEX idx_active_play_to_win_games
    ON play_to_win_games(deleted_at)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_play_to_win_game_game_id ON play_to_win_games(game_id);
CREATE INDEX idx_play_to_win_session_play_to_win_id ON play_to_win_sessions(play_to_win_id);
CREATE INDEX idx_play_to_win_entries_session_id ON play_to_win_entries(session_id);

CREATE VIEW vw_library_games AS
SELECT id, title, sanitized_title, barcode, created_at
FROM games
WHERE deleted_at IS NULL;

CREATE VIEW vw_library_patrons AS
SELECT id, full_name, barcode, created_at
FROM patrons
WHERE deleted_at IS NULL;

CREATE VIEW vw_play_to_win_games AS
SELECT  id, game_id, created_at
FROM play_to_win_games
WHERE deleted_at IS NULL;

CREATE VIEW vw_deleted_play_to_win_games AS
SELECT  id, game_id, deleted_at, deletion_reason, deletion_reason_comment
FROM play_to_win_games
WHERE deleted_at IS NOT NULL;

CREATE VIEW vw_play_to_win_sessions AS
SELECT  id, play_to_win_id, playtime_minutes, created_at
FROM play_to_win_sessions
WHERE deleted_at IS NULL;

CREATE VIEW vw_deleted_play_to_win_sessions AS
SELECT id, play_to_win_id, deleted_at, deletion_reason, deletion_reason_comment
FROM play_to_win_sessions
WHERE deleted_at IS NOT NULL;

CREATE VIEW vw_play_to_win_entries AS
SELECT ptw_entries.id,
       ptw_entries.session_id,
       ptw_sessions.play_to_win_id,
       ptw_entries.entrant_name,
       ptw_entries.entrant_unique_id,
       ptw_entries.created_at
FROM play_to_win_entries ptw_entries
LEFT JOIN vw_play_to_win_sessions ptw_sessions ON ptw_sessions.id = ptw_entries.session_id
WHERE ptw_entries.deleted_at IS NULL;

CREATE VIEW vw_deleted_play_to_win_entries AS
SELECT id, session_id, deleted_at, deletion_reason, deletion_reason_comment
FROM play_to_win_entries
WHERE deleted_at IS NOT NULL;

CREATE VIEW vw_library_transaction_events AS
SELECT
    te.transaction_id,
    g.id AS game_id,
    COALESCE(g.title, 'Missing Game') AS game_title,
    g.sanitized_title AS sanitized_title,
    te.patron_id,
    COALESCE(p.full_name, 'Missing Patron') AS patron_full_name,
    te.event_timestamp,
    te.event_type
FROM transaction_events AS te
         LEFT JOIN games AS g ON te.game_id = g.id
         LEFT JOIN patrons AS p ON te.patron_id = p.id
ORDER BY te.event_timestamp DESC;

CREATE VIEW vw_game_status AS
SELECT DISTINCT ON (g.id)
        g.id AS game_id,
        g.title AS game_title,
        g.sanitized_title,
        t.patron_id,
        p.full_name AS patron_full_name,
        t.id AS transaction_id,
        t.checkout_timestamp,
        t.checkin_timestamp,
        ptw.id AS play_to_win_game_id
FROM vw_library_games AS g
LEFT JOIN transactions AS t ON t.game_id = g.id
LEFT JOIN vw_library_patrons AS p ON t.patron_id = p.id
LEFT JOIN play_to_win_games AS ptw ON ptw.game_id = g.id
ORDER BY g.id, t.checkout_timestamp DESC;

CREATE OR REPLACE FUNCTION fn_record_checkout_event()
    RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO transaction_events (transaction_id, game_id, patron_id, event_type, event_timestamp)
    VALUES (NEW.id, NEW.game_id, NEW.patron_id, 'check_out', NEW.checkout_timestamp);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_checkout_event
    AFTER INSERT ON transactions
    FOR EACH ROW EXECUTE FUNCTION fn_record_checkout_event();


CREATE OR REPLACE FUNCTION fn_record_checkin_event()
    RETURNS TRIGGER AS $$
BEGIN
    IF OLD.checkin_timestamp IS NULL AND NEW.checkin_timestamp IS NOT NULL THEN
        INSERT INTO transaction_events (transaction_id, game_id, patron_id, event_type, event_timestamp)
        VALUES (NEW.id, NEW.game_id, NEW.patron_id, 'check_in', NEW.checkin_timestamp);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_checkin_event
    AFTER UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION fn_record_checkin_event();
