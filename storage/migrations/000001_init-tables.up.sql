CREATE TABLE IF NOT EXISTS loader (
    uuid TEXT PRIMARY KEY,
    description TEXT DEFAULT "" NOT NULL,
    aggregate_window INTEGER,
    name TEXT UNIQUE,
    create_date DATE DEFAULT (datetime('now')),
    url TEXT,
    method TEXT,
    http_engine TEXT,
    skip_verify INTEGER,
    ca TEXT,
    cert TEXT,
    key TEXT,
    benchmark_timeout INTEGER,
    body TEXT,
    gather_full_requests_stats INTEGER,
    gather_aggregate_requests_stats INTEGER
);

CREATE TABLE IF NOT EXISTS loader_requests_details (
    id INTEGER PRIMARY KEY,
    request_count INTEGER,
    abort_after INTEGER,
    connections INTEGER,
    rate_limit INTEGER,
    duration TEXT,
    keep_alive INTEGER,
    request_delay INTEGER,
    read_timeout INTEGER,
    write_timeout INTEGER,
    timeout INTEGER,

    loader_uuid TEXT,
    FOREIGN KEY (loader_uuid) REFERENCES loader (uuid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS header (
    id INTEGER PRIMARY KEY,
    header TEXT,
    loader_uuid TEXT,

    FOREIGN KEY (loader_uuid) REFERENCES loader (uuid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS parameter (
    id INTEGER PRIMARY KEY,
    parameters TEXT,
    loader_uuid TEXT,

    FOREIGN KEY (loader_uuid) REFERENCES loader (uuid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS summary (
    uuid TEXT PRIMARY KEY,
    url TEXT,
    description TEXT,
    start DATETIME,
    end DATETIME,
    total_time INTEGER,
    requests_count INTEGER,
    success_req INTEGER,
    fail_req INTEGER,
    data_transferred INTEGER,
    req_per_sec REAL,
    avg_req_time INTEGER,
    min_req_time INTEGER,
    max_req_time INTEGER,
    p50_req_time INTEGER,
    p75_req_time INTEGER,
    p90_req_time INTEGER,
    p99_req_time INTEGER,
    std_deviation REAL,
    loader_uuid TEXT,

    FOREIGN KEY(loader_uuid) REFERENCES loader (uuid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS aggregated_stats (
    id INTEGER PRIMARY KEY,
    start DATETIME,
    end DATETIME,
    duration INTEGER,
    avg_request_time INTEGER,
    max_request_time INTEGER,
    min_request_time INTEGER,
    request_count INTEGER,
    summary_uuid TEXT,
    FOREIGN KEY(summary_uuid) REFERENCES summary(uuid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS requests_stats (
    id INTEGER PRIMARY KEY,
    start DATETIME,
    end DATETIME,
    duration INTEGER,
    ret_code INTEGER,
    body_size INTEGER,
    error TEXT,
    summary_uuid TEXT,
    FOREIGN KEY (summary_uuid) REFERENCES summary(uuid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS errors (
    id INTEGER PRIMARY KEY,
    name TEXT,
    count INTEGER,
    summary TEXT,
    FOREIGN KEY(summary) REFERENCES summary(uuid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS http_codes (
    id INTEGER PRIMARY KEY,
    code INTEGER,
    count INTEGER,
    summary TEXT,
    FOREIGN KEY(summary) REFERENCES summary(uuid) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS loader_tag (
    id INTEGER PRIMARY KEY,
    key TEXT,
    value TEXT,
    loader_uuid TEXT,
    UNIQUE(key,loader_uuid),
    FOREIGN KEY(loader_uuid) REFERENCES loader (uuid) ON DELETE CASCADE
)