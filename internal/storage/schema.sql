CREATE TABLE IF NOT EXISTS checks (
    id       INTEGER PRIMARY KEY,
    name     TEXT    NOT NULL UNIQUE,
    target   TEXT    NOT NULL,
    interval INTEGER NOT NULL,  
    timeout  INTEGER NOT NULL  
);

CREATE TABLE IF NOT EXISTS results (
    id         INTEGER PRIMARY KEY,
    check_id   INTEGER NOT NULL REFERENCES checks(id),
    status     TEXT    NOT NULL,  
    latency_ms INTEGER NOT NULL,
    error      TEXT    NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL  
);

CREATE INDEX IF NOT EXISTS idx_results_check_created
    ON results(check_id, created_at DESC);

CREATE TABLE IF NOT EXISTS uptime (
    check_id   INTEGER PRIMARY KEY REFERENCES checks(id),
    uptime_1h  REAL    NOT NULL DEFAULT 0,  
    uptime_24h REAL    NOT NULL DEFAULT 0,
    uptime_7d  REAL    NOT NULL DEFAULT 0
);
