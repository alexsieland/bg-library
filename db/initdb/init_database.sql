CREATE TABLE games (
    id UUID DEFAULT gen_random_uuid(),
    title VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (id)
);

CREATE TABLE patrons (
    id UUID DEFAULT gen_random_uuid(),
    full_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (id)
);

CREATE TABLE transactions (
    id UUID DEFAULT gen_random_uuid(),
    game_id UUID NOT NULL,
    patron_id UUID NOT NULL,
    checkout_timestamp TIMESTAMP DEFAULT NOW(),
    checkin_timestamp TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE INDEX idx_game_titles ON games(title);

CREATE INDEX idx_patron_full_name ON patrons(full_name);

CREATE INDEX idx_checked_out_games
    ON transactions(checkin_timestamp)
    WHERE checkin_timestamp IS NULL;