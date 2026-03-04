CREATE TABLE IF NOT EXISTS users (
  user_id TEXT PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'Viewer',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL,
  last_login_at TEXT,
  updated_at TEXT
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
  token_id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(user_id),
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TEXT NOT NULL,
  revoked INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  user_agent TEXT,
  ip_address TEXT
);

CREATE TABLE IF NOT EXISTS token_blacklist (
  jti TEXT PRIMARY KEY,
  expires_at TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks (
  task_id TEXT PRIMARY KEY,
  task_type TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'PENDING',
  request_json TEXT,
  exit_code INTEGER,
  stdout_tail TEXT,
  stderr_tail TEXT,
  log_path TEXT,
  created_by TEXT REFERENCES users(user_id),
  created_at TEXT NOT NULL,
  started_at TEXT,
  finished_at TEXT
);

CREATE TABLE IF NOT EXISTS revisions (
  revision_id TEXT PRIMARY KEY,
  target_type TEXT NOT NULL,
  target_id TEXT,
  content TEXT NOT NULL,
  sha256 TEXT NOT NULL,
  created_at TEXT NOT NULL,
  created_by TEXT REFERENCES users(user_id)
);

CREATE TABLE IF NOT EXISTS backups (
  backup_id TEXT PRIMARY KEY,
  label TEXT,
  scope_json TEXT NOT NULL,
  manifest_path TEXT NOT NULL,
  size_bytes INTEGER,
  sha256 TEXT NOT NULL,
  verified INTEGER DEFAULT 0,
  created_by TEXT REFERENCES users(user_id),
  created_at TEXT NOT NULL
);
