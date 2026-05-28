CREATE TABLE games (
    id UUID UNIQUE DEFAULT gen_random_uuid(),
    title VARCHAR(100) NOT NULL, -- Immutable for historical reference but not used for searching or display
    display_title VARCHAR(100),
    sanitized_title VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    barcode VARCHAR(48), -- Not unique because a library might use UPCs
    PRIMARY KEY (id)
);

CREATE TABLE patrons (
    id UUID UNIQUE DEFAULT gen_random_uuid(),
    full_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    barcode VARCHAR(48) UNIQUE,
    PRIMARY KEY (id)
);

CREATE TABLE transactions (
    id UUID UNIQUE DEFAULT gen_random_uuid(),
    game_id UUID REFERENCES games(id),
    patron_id UUID REFERENCES patrons(id),
    checkout_timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    checkin_timestamp TIMESTAMP,
    PRIMARY KEY (id)
);


CREATE TYPE transaction_event_type AS ENUM ('check_out', 'check_in');
CREATE TABLE transaction_events (
    id UUID UNIQUE DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id),
    game_id UUID NOT NULL REFERENCES games(id),
    patron_id UUID NOT NULL REFERENCES patrons(id),
    event_type transaction_event_type NOT NULL,
    event_timestamp TIMESTAMP NOT NULL,
    recorded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);

CREATE TABLE play_to_win_groups (
    id UUID UNIQUE DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE TYPE play_to_win_game_deletion_type AS ENUM ('claimed', 'mistake', 'other');
CREATE TABLE play_to_win_games (
    id UUID UNIQUE DEFAULT gen_random_uuid(),
    game_id UUID NOT NULL UNIQUE REFERENCES games(id),
    ptw_group_id UUID NOT NULL REFERENCES play_to_win_groups(id),
    winner_id UUID,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    deletion_reason play_to_win_game_deletion_type,
    deletion_reason_comment VARCHAR(500),
    PRIMARY KEY (id)
);

CREATE TYPE play_to_win_session_deletion_type AS ENUM ('foul_play', 'too_many_players', 'too_few_players', 'abnormal_playtime', 'other');
CREATE TABLE play_to_win_sessions (
    id UUID UNIQUE DEFAULT gen_random_uuid(),
    ptw_group_id UUID NOT NULL REFERENCES play_to_win_groups(id),
    playtime_minutes INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    deletion_reason play_to_win_session_deletion_type,
    deletion_reason_comment VARCHAR(500),
    PRIMARY KEY (id),
    UNIQUE(id, ptw_group_id)
);

CREATE TYPE play_to_win_entry_deletion_type AS ENUM ('won','failed_to_claim', 'foul_play', 'duplicate_entrant', 'other');
CREATE TABLE play_to_win_entries (
    id UUID UNIQUE DEFAULT gen_random_uuid(),
    ptw_session_id UUID NOT NULL REFERENCES play_to_win_sessions(id),
    ptw_group_id UUID NOT NULL REFERENCES play_to_win_groups(id),
    entrant_name VARCHAR(100) NOT NULL,
    entrant_unique_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    deletion_reason play_to_win_entry_deletion_type,
    deletion_reason_comment VARCHAR(500),
    PRIMARY KEY (id),
    UNIQUE(id, ptw_group_id),
    UNIQUE(ptw_session_id, entrant_unique_id),
    UNIQUE(ptw_group_id, entrant_unique_id),
    FOREIGN KEY (ptw_session_id, ptw_group_id) REFERENCES play_to_win_sessions(id, ptw_group_id)
);

ALTER TABLE play_to_win_games
ADD CONSTRAINT fk_play_to_win_games_winner_id
FOREIGN KEY (winner_id, ptw_group_id)
REFERENCES play_to_win_entries(id, ptw_group_id);

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
CREATE INDEX idx_play_to_win_session_ptw_group_id ON play_to_win_sessions(ptw_group_id);
CREATE INDEX idx_play_to_win_entries_ptw_session_id ON play_to_win_entries(ptw_session_id);

CREATE VIEW vw_play_to_win_games AS
SELECT  ga.id,
        ga.game_id,
        ga.ptw_group_id,
        gr.name AS group_name,
        ga.created_at,
        ga.winner_id
FROM play_to_win_games AS ga
LEFT JOIN play_to_win_groups AS gr ON gr.id = ga.ptw_group_id
WHERE ga.deleted_at IS NULL;

CREATE VIEW vw_library_games AS
SELECT
    g.id,
    COALESCE(g.display_title, g.title) AS display_title,
    g.title,
    g.sanitized_title,
    g.barcode,
    ptw.id AS play_to_win_game_id,
    g.created_at
FROM games AS g
LEFT JOIN vw_play_to_win_games AS ptw ON g.id = ptw.game_id
WHERE deleted_at IS NULL;

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
