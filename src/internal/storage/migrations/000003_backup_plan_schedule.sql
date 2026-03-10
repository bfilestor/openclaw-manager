ALTER TABLE backup_plans ADD COLUMN schedule_kind TEXT NOT NULL DEFAULT 'interval';
ALTER TABLE backup_plans ADD COLUMN daily_time TEXT;
ALTER TABLE backup_plans ADD COLUMN monthly_day INTEGER;
ALTER TABLE backup_plans ADD COLUMN retention_count INTEGER NOT NULL DEFAULT 30;

ALTER TABLE backups ADD COLUMN plan_id TEXT REFERENCES backup_plans(plan_id);
CREATE INDEX IF NOT EXISTS idx_backups_plan_id_created_at ON backups(plan_id, created_at DESC);
