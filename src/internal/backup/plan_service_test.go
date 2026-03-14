package backup

import (
	"testing"
	"time"

	"openclaw-manager/internal/storage"
)

func TestDeletePlanKeepsBackupsByDetachingPlanID(t *testing.T) {
	db := storage.NewTestDB(t)
	svc := &PlanService{DB: db.SQL}

	_, err := db.SQL.Exec(`INSERT INTO users(user_id,username,password_hash,role,status,created_at) VALUES(?,?,?,?,?,?)`, "u1", "u1", "x", "Admin", "active", time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.SQL.Exec(`INSERT INTO backup_plans(plan_id,name,label,scope_json,schedule_kind,daily_time,monthly_day,interval_minutes,retention_count,enabled,last_run_at,next_run_at,created_by,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		"p1", "daily plan", "lab", `[]`, "interval", nil, nil, 15, 30, 1, nil, now, "u1", now, now)
	if err != nil {
		t.Fatalf("insert plan: %v", err)
	}

	_, err = db.SQL.Exec(`INSERT INTO backups(backup_id,label,scope_json,manifest_path,size_bytes,sha256,verified,created_by,created_at,plan_id) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		"b1", "lab", `[]`, "/tmp/b1.manifest.json", 1, "abc", 1, "u1", now, "p1")
	if err != nil {
		t.Fatalf("insert backup: %v", err)
	}

	if err := svc.DeletePlan("p1"); err != nil {
		t.Fatalf("delete plan: %v", err)
	}

	var count int
	if err := db.SQL.QueryRow(`SELECT COUNT(1) FROM backup_plans WHERE plan_id=?`, "p1").Scan(&count); err != nil {
		t.Fatalf("count plans: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected plan deleted, got count=%d", count)
	}

	var planID any
	if err := db.SQL.QueryRow(`SELECT plan_id FROM backups WHERE backup_id=?`, "b1").Scan(&planID); err != nil {
		t.Fatalf("query backup plan_id: %v", err)
	}
	if planID != nil {
		t.Fatalf("expected backup plan_id detached, got %#v", planID)
	}
}
