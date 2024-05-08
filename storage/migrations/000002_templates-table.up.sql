CREATE TABLE IF NOT EXISTS templates (
    name TEXT NOT NULL PRIMARY KEY,
    create_date DATE DEFAULT (datetime('now')),
    update_date DATE DEFAULT (datetime('now')),
    content TEXT NOT NULL
);