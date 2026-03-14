package backup

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"openclaw-manager/internal/auth"
	"openclaw-manager/internal/storage"
)

func TestDeletePlanAPIWithExistingBackups(t *testing.T) {
	db := storage.NewTestDB(t)
	backupSvc := &Service{DB: db.SQL, BackupHome: t.TempDir(), OpenclawHome: t.TempDir(), ManagerHome: t.TempDir()}
	planSvc := &PlanService{DB: db.SQL, Backup: backupSvc}
	h := &APIHandler{DB: db.SQL, Service: backupSvc, PlanSvc: planSvc}

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

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/backup-plans/p1", nil)
	w := httptest.NewRecorder()
	h.DeletePlan(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got=%d body=%s", w.Code, w.Body.String())
	}

	var count int
	if err := db.SQL.QueryRow(`SELECT COUNT(1) FROM backup_plans WHERE plan_id=?`, "p1").Scan(&count); err != nil {
		t.Fatalf("count plans: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected plan deleted, got count=%d", count)
	}
}

func TestRunPlanNowCreatesAsyncTask(t *testing.T) {
	db := storage.NewTestDB(t)
	home := t.TempDir()
	backupSvc := &Service{DB: db.SQL, BackupHome: t.TempDir(), OpenclawHome: home, ManagerHome: t.TempDir()}
	planSvc := &PlanService{DB: db.SQL, Backup: backupSvc}
	h := &APIHandler{DB: db.SQL, Service: backupSvc, PlanSvc: planSvc}

	if err := os.WriteFile(filepath.Join(home, "openclaw.json"), []byte(`{"ok":true}`), 0o644); err != nil {
		t.Fatalf("write openclaw.json: %v", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.SQL.Exec(`INSERT INTO users(user_id,username,password_hash,role,status,created_at) VALUES(?,?,?,?,?,?)`, "u1", "u1", "x", "Admin", "active", now)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	_, err = db.SQL.Exec(`INSERT INTO backup_plans(plan_id,name,label,scope_json,schedule_kind,daily_time,monthly_day,interval_minutes,retention_count,enabled,last_run_at,next_run_at,created_by,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		"p1", "daily plan", "lab", `["openclaw_json"]`, "interval", nil, nil, 15, 30, 1, nil, now, "u1", now, now)
	if err != nil {
		t.Fatalf("insert plan: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/backup-plans/p1/run", nil)
	req = req.WithContext(auth.WithUserContext(req.Context(), &auth.UserContext{UserID: "u1"}))
	w := httptest.NewRecorder()
	h.RunPlanNow(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202 got=%d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		TaskID string `json:"task_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.TaskID == "" {
		t.Fatalf("expected task_id in response, body=%s", w.Body.String())
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var status string
		if err := db.SQL.QueryRow(`SELECT status FROM tasks WHERE task_id=?`, resp.TaskID).Scan(&status); err == nil && status == "SUCCEEDED" {
			var backups int
			if err := db.SQL.QueryRow(`SELECT COUNT(1) FROM backups WHERE plan_id=?`, "p1").Scan(&backups); err != nil {
				t.Fatalf("count backups: %v", err)
			}
			if backups == 0 {
				t.Fatalf("expected backup created by async plan run")
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	var status, stderr string
	_ = db.SQL.QueryRow(`SELECT status, COALESCE(stderr_tail,'') FROM tasks WHERE task_id=?`, resp.TaskID).Scan(&status, &stderr)
	t.Fatalf("task %s not finished in time, last_status=%s stderr=%s", resp.TaskID, status, stderr)
}
