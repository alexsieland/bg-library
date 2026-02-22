CREATE TABLE games (
    id UUID DEFAULT gen_random_uuid(),
    title VARCHAR(100) NOT NULL,
    sanitized_title VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id)
);

CREATE TABLE patrons (
    id UUID DEFAULT gen_random_uuid(),
    full_name VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (id)
);

CREATE TABLE transactions (
    id UUID DEFAULT gen_random_uuid(),
    game_id UUID REFERENCES games(id),
    patron_id UUID REFERENCES patrons(id),
    checkout_timestamp TIMESTAMP DEFAULT NOW(),
    checkin_timestamp TIMESTAMP,
    PRIMARY KEY (id)
);

CREATE INDEX idx_game_titles ON games(sanitized_title);

CREATE INDEX idx_patron_full_name ON patrons(full_name);

CREATE INDEX idx_checkout_timestamp ON transactions(checkout_timestamp);

CREATE INDEX idx_checked_out_games
ON transactions(checkin_timestamp)
WHERE checkin_timestamp IS NULL;

CREATE INDEX idx_active_games
ON games(deleted)
WHERE deleted IS NOT NULL;

CREATE INDEX idx_active_patrons
ON patrons(deleted)
WHERE deleted IS NOT NULL;

CREATE VIEW vw_library_games AS
SELECT id, title, sanitized_title, created_at
FROM games
WHERE deleted IS FALSE;

CREATE VIEW vw_library_patrons AS
SELECT id, full_name, created_at
FROM patrons
WHERE deleted IS FALSE;

CREATE VIEW vw_game_status AS
SELECT DISTINCT ON (g.id)
        g.id AS game_id,
        g.title AS game_title,
        g.sanitized_title,
        t.patron_id,
        p.full_name AS patron_full_name,
        t.id AS transaction_id,
        t.checkout_timestamp,
        t.checkin_timestamp
FROM vw_library_games AS g
LEFT JOIN transactions AS t ON t.game_id = g.id
LEFT JOIN vw_library_patrons AS p ON t.patron_id = p.id
ORDER BY g.id, t.checkout_timestamp DESC;
