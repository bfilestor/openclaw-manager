CREATE TABLE IF NOT EXISTS backup_plans (
  plan_id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  label TEXT,
  scope_json TEXT NOT NULL,
  interval_minutes INTEGER NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 1,
  last_run_at TEXT,
  next_run_at TEXT,
  created_by TEXT REFERENCES users(user_id),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS backup_preferences (
  user_id TEXT PRIMARY KEY REFERENCES users(user_id),
  label TEXT,
  scope_json TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
