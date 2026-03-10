package backup

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type BackupPlan struct {
	PlanID           string   `json:"plan_id"`
	Name             string   `json:"name"`
	Label            string   `json:"label"`
	Scope            []string `json:"scope"`
	IntervalMinutes  int      `json:"interval_minutes"`
	Enabled          bool     `json:"enabled"`
	LastRunAt        string   `json:"last_run_at,omitempty"`
	NextRunAt        string   `json:"next_run_at,omitempty"`
	CreatedBy        string   `json:"created_by,omitempty"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

type BackupPreference struct {
	Label string   `json:"label"`
	Scope []string `json:"scope"`
}

type PlanService struct {
	DB      *sql.DB
	Backup  *Service
	mu      sync.Mutex
	started bool
	stopCh  chan struct{}
}

func (s *PlanService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return
	}
	s.started = true
	s.stopCh = make(chan struct{})
	go s.loop(s.stopCh)
}

func (s *PlanService) loop(stopCh chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = s.runDuePlans()
		case <-stopCh:
			return
		}
	}
}

func (s *PlanService) runDuePlans() error {
	now := time.Now().UTC()
	rows, err := s.DB.Query(`SELECT plan_id,name,label,scope_json,interval_minutes,enabled,last_run_at,next_run_at,created_by,created_at,updated_at FROM backup_plans WHERE enabled=1`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		p, err := scanPlan(rows)
		if err != nil {
			continue
		}
		nextAt, err := time.Parse(time.RFC3339, p.NextRunAt)
		if err != nil || nextAt.After(now) {
			continue
		}
		_, _ = s.ExecuteNow(p.PlanID, p.CreatedBy)
	}
	return nil
}

func (s *PlanService) CreatePlan(name, label string, scope []string, intervalMinutes int, createdBy string) (*BackupPlan, error) {
	if intervalMinutes <= 0 {
		return nil, fmt.Errorf("interval_minutes must be > 0")
	}
	now := time.Now().UTC()
	id := uuid.NewString()
	scopeJSON, _ := json.Marshal(scope)
	nextRunAt := now.Add(time.Duration(intervalMinutes) * time.Minute).Format(time.RFC3339)
	_, err := s.DB.Exec(`INSERT INTO backup_plans(plan_id,name,label,scope_json,interval_minutes,enabled,last_run_at,next_run_at,created_by,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
		id, name, label, string(scopeJSON), intervalMinutes, 1, nil, nextRunAt, nullIfEmpty(createdBy), now.Format(time.RFC3339), now.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	return s.GetPlan(id)
}

func (s *PlanService) ListPlans() ([]*BackupPlan, error) {
	rows, err := s.DB.Query(`SELECT plan_id,name,label,scope_json,interval_minutes,enabled,last_run_at,next_run_at,created_by,created_at,updated_at FROM backup_plans ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*BackupPlan, 0)
	for rows.Next() {
		p, err := scanPlan(rows)
		if err == nil {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *PlanService) GetPlan(planID string) (*BackupPlan, error) {
	row := s.DB.QueryRow(`SELECT plan_id,name,label,scope_json,interval_minutes,enabled,last_run_at,next_run_at,created_by,created_at,updated_at FROM backup_plans WHERE plan_id=?`, planID)
	return scanPlan(row)
}

func (s *PlanService) UpdatePlan(planID, name, label string, scope []string, intervalMinutes int) (*BackupPlan, error) {
	if intervalMinutes <= 0 {
		return nil, fmt.Errorf("interval_minutes must be > 0")
	}
	scopeJSON, _ := json.Marshal(scope)
	now := time.Now().UTC()
	nextRunAt := now.Add(time.Duration(intervalMinutes) * time.Minute).Format(time.RFC3339)
	res, err := s.DB.Exec(`UPDATE backup_plans SET name=?,label=?,scope_json=?,interval_minutes=?,next_run_at=?,updated_at=? WHERE plan_id=?`,
		name, label, string(scopeJSON), intervalMinutes, nextRunAt, now.Format(time.RFC3339), planID)
	if err != nil {
		return nil, err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return nil, sql.ErrNoRows
	}
	return s.GetPlan(planID)
}

func (s *PlanService) SetPlanEnabled(planID string, enabled bool) error {
	now := time.Now().UTC()
	nextRunAt := interface{}(nil)
	if enabled {
		var intervalMinutes int
		if err := s.DB.QueryRow(`SELECT interval_minutes FROM backup_plans WHERE plan_id=?`, planID).Scan(&intervalMinutes); err != nil {
			return err
		}
		nextRunAt = now.Add(time.Duration(intervalMinutes) * time.Minute).Format(time.RFC3339)
	}
	res, err := s.DB.Exec(`UPDATE backup_plans SET enabled=?,next_run_at=?,updated_at=? WHERE plan_id=?`, boolToInt(enabled), nextRunAt, now.Format(time.RFC3339), planID)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *PlanService) DeletePlan(planID string) error {
	res, err := s.DB.Exec(`DELETE FROM backup_plans WHERE plan_id=?`, planID)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *PlanService) ExecuteNow(planID, createdBy string) (string, error) {
	p, err := s.GetPlan(planID)
	if err != nil {
		return "", err
	}
	backupID, err := s.Backup.Create(p.Scope, p.Label, createdBy)
	now := time.Now().UTC()
	if err == nil {
		next := now.Add(time.Duration(p.IntervalMinutes) * time.Minute).Format(time.RFC3339)
		_, _ = s.DB.Exec(`UPDATE backup_plans SET last_run_at=?,next_run_at=?,updated_at=? WHERE plan_id=?`, now.Format(time.RFC3339), next, now.Format(time.RFC3339), planID)
	}
	return backupID, err
}

func (s *PlanService) SavePreference(userID, label string, scope []string) error {
	if userID == "" {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339)
	scopeJSON, _ := json.Marshal(scope)
	_, err := s.DB.Exec(`INSERT INTO backup_preferences(user_id,label,scope_json,updated_at) VALUES(?,?,?,?) ON CONFLICT(user_id) DO UPDATE SET label=excluded.label,scope_json=excluded.scope_json,updated_at=excluded.updated_at`,
		userID, label, string(scopeJSON), now)
	return err
}

func (s *PlanService) GetPreference(userID string) (*BackupPreference, error) {
	if userID == "" {
		return &BackupPreference{Scope: []string{}}, nil
	}
	var label, scopeJSON string
	err := s.DB.QueryRow(`SELECT label,scope_json FROM backup_preferences WHERE user_id=?`, userID).Scan(&label, &scopeJSON)
	if err == sql.ErrNoRows {
		return &BackupPreference{Scope: []string{}}, nil
	}
	if err != nil {
		return nil, err
	}
	var scope []string
	_ = json.Unmarshal([]byte(scopeJSON), &scope)
	return &BackupPreference{Label: label, Scope: scope}, nil
}

func scanPlan(scanner interface{ Scan(dest ...any) error }) (*BackupPlan, error) {
	var p BackupPlan
	var scopeJSON string
	var enabledInt int
	var lastRunAt, nextRunAt, createdBy sql.NullString
	if err := scanner.Scan(&p.PlanID, &p.Name, &p.Label, &scopeJSON, &p.IntervalMinutes, &enabledInt, &lastRunAt, &nextRunAt, &createdBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(scopeJSON), &p.Scope)
	p.Enabled = enabledInt == 1
	if lastRunAt.Valid {
		p.LastRunAt = lastRunAt.String
	}
	if nextRunAt.Valid {
		p.NextRunAt = nextRunAt.String
	}
	if createdBy.Valid {
		p.CreatedBy = createdBy.String
	}
	return &p, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
