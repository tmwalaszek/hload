ALTER TABLE loader_tag ADD COLUMN create_date DATE DEFAULT (datetime('now'));
ALTER TABLE loader_tag ADD COLUMN update_date DATE DEFAULT (datetime('now'))
