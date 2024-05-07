UPDATE templates SET content = $1, update_date = datetime('now') WHERE name = $2;
