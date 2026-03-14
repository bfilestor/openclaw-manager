package backup

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type BackupPlan struct {
	PlanID          string   `json:"plan_id"`
	Name            string   `json:"name"`
	Label           string   `json:"label"`
	Scope           []string `json:"scope"`
	ScheduleKind    string   `json:"schedule_kind"`
	DailyTime       string   `json:"daily_time,omitempty"`
	MonthlyDay      int      `json:"monthly_day,omitempty"`
	IntervalMinutes int      `json:"interval_minutes,omitempty"`
	RetentionCount  int      `json:"retention_count"`
	Enabled         bool     `json:"enabled"`
	LastRunAt       string   `json:"last_run_at,omitempty"`
	NextRunAt       string   `json:"next_run_at,omitempty"`
	CreatedBy       string   `json:"created_by,omitempty"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

type BackupPreference struct {
	Label string   `json:"label"`
	Scope []string `json:"scope"`
}

type PlanService struct {
	DB     *sql.DB
	Backup *Service
	mu     sync.Mutex
	start  bool
	stopCh chan struct{}
}

func (s *PlanService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.start {
		return
	}
	s.start = true
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
	plans, err := s.ListPlans()
	if err != nil {
		return err
	}
	for _, p := range plans {
		if !p.Enabled || strings.TrimSpace(p.NextRunAt) == "" {
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

func (s *PlanService) CreatePlan(p *BackupPlan, createdBy string) (*BackupPlan, error) {
	if p == nil {
		return nil, fmt.Errorf("plan is nil")
	}
	if strings.TrimSpace(p.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if len(p.Scope) == 0 {
		return nil, fmt.Errorf("scope is required")
	}
	if p.RetentionCount <= 0 {
		p.RetentionCount = 30
	}
	nextRunAt, err := computeNextRunAt(p, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	id := uuid.NewString()
	scopeJSON, _ := json.Marshal(p.Scope)
	_, err = s.DB.Exec(`INSERT INTO backup_plans(plan_id,name,label,scope_json,schedule_kind,daily_time,monthly_day,interval_minutes,retention_count,enabled,last_run_at,next_run_at,created_by,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		id,
		p.Name,
		p.Label,
		string(scopeJSON),
		normalizeScheduleKind(p.ScheduleKind),
		nullIfEmpty(p.DailyTime),
		nullIfZero(p.MonthlyDay),
		nullIfZero(p.IntervalMinutes),
		p.RetentionCount,
		1,
		nil,
		nextRunAt,
		nullIfEmpty(createdBy),
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	return s.GetPlan(id)
}

func (s *PlanService) ListPlans() ([]*BackupPlan, error) {
	rows, err := s.DB.Query(`SELECT plan_id,name,label,scope_json,schedule_kind,daily_time,monthly_day,interval_minutes,retention_count,enabled,last_run_at,next_run_at,created_by,created_at,updated_at FROM backup_plans ORDER BY created_at DESC`)
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
	row := s.DB.QueryRow(`SELECT plan_id,name,label,scope_json,schedule_kind,daily_time,monthly_day,interval_minutes,retention_count,enabled,last_run_at,next_run_at,created_by,created_at,updated_at FROM backup_plans WHERE plan_id=?`, planID)
	return scanPlan(row)
}

func (s *PlanService) UpdatePlan(planID string, p *BackupPlan) (*BackupPlan, error) {
	if p == nil {
		return nil, fmt.Errorf("plan is nil")
	}
	if p.RetentionCount <= 0 {
		p.RetentionCount = 30
	}
	nextRunAt, err := computeNextRunAt(p, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	scopeJSON, _ := json.Marshal(p.Scope)
	now := time.Now().UTC()
	res, err := s.DB.Exec(`UPDATE backup_plans SET name=?,label=?,scope_json=?,schedule_kind=?,daily_time=?,monthly_day=?,interval_minutes=?,retention_count=?,next_run_at=?,updated_at=? WHERE plan_id=?`,
		p.Name, p.Label, string(scopeJSON), normalizeScheduleKind(p.ScheduleKind), nullIfEmpty(p.DailyTime), nullIfZero(p.MonthlyDay), nullIfZero(p.IntervalMinutes), p.RetentionCount, nextRunAt, now.Format(time.RFC3339), planID)
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
		p, err := s.GetPlan(planID)
		if err != nil {
			return err
		}
		next, err := computeNextRunAt(p, now)
		if err != nil {
			return err
		}
		nextRunAt = next
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
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Keep historical backups, but break FK references before removing the plan.
	if _, err := tx.Exec(`UPDATE backups SET plan_id=NULL WHERE plan_id=?`, planID); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM backup_plans WHERE plan_id=?`, planID)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return sql.ErrNoRows
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *PlanService) ExecuteNow(planID, createdBy string) (string, error) {
	p, err := s.GetPlan(planID)
	if err != nil {
		return "", err
	}
	backupID, err := s.Backup.Create(p.Scope, p.Label, createdBy, planID)
	now := time.Now().UTC()
	if err == nil {
		next, _ := computeNextRunAt(p, now)
		_, _ = s.DB.Exec(`UPDATE backup_plans SET last_run_at=?,next_run_at=?,updated_at=? WHERE plan_id=?`, now.Format(time.RFC3339), next, now.Format(time.RFC3339), planID)
		_ = s.pruneBackups(planID, p.RetentionCount)
	}
	return backupID, err
}

func (s *PlanService) pruneBackups(planID string, retention int) error {
	if retention <= 0 {
		retention = 30
	}
	type item struct{ id string }
	rows, err := s.DB.Query(`SELECT backup_id FROM backups WHERE plan_id=? ORDER BY created_at DESC`, planID)
	if err != nil {
		return err
	}
	defer rows.Close()
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	if len(ids) <= retention {
		return nil
	}
	for _, id := range ids[retention:] {
		_, _ = s.DB.Exec(`DELETE FROM backups WHERE backup_id=?`, id)
		_ = removeIfExists(s.Backup.BackupHome + "/" + id + ".tar.gz")
		_ = removeIfExists(s.Backup.BackupHome + "/" + id + ".manifest.json")
	}
	return nil
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
	var lastRunAt, nextRunAt, createdBy, dailyTime sql.NullString
	var monthlyDay, intervalMinutes sql.NullInt64
	if err := scanner.Scan(&p.PlanID, &p.Name, &p.Label, &scopeJSON, &p.ScheduleKind, &dailyTime, &monthlyDay, &intervalMinutes, &p.RetentionCount, &enabledInt, &lastRunAt, &nextRunAt, &createdBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(scopeJSON), &p.Scope)
	p.Enabled = enabledInt == 1
	if dailyTime.Valid {
		p.DailyTime = dailyTime.String
	}
	if monthlyDay.Valid {
		p.MonthlyDay = int(monthlyDay.Int64)
	}
	if intervalMinutes.Valid {
		p.IntervalMinutes = int(intervalMinutes.Int64)
	}
	if lastRunAt.Valid {
		p.LastRunAt = lastRunAt.String
	}
	if nextRunAt.Valid {
		p.NextRunAt = nextRunAt.String
	}
	if createdBy.Valid {
		p.CreatedBy = createdBy.String
	}
	if p.RetentionCount <= 0 {
		p.RetentionCount = 30
	}
	return &p, nil
}

func computeNextRunAt(p *BackupPlan, from time.Time) (string, error) {
	kind := normalizeScheduleKind(p.ScheduleKind)
	now := from.UTC().Truncate(time.Second)
	switch kind {
	case "daily":
		h, m, s, err := parseHMS(p.DailyTime)
		if err != nil {
			return "", fmt.Errorf("daily_time invalid: %w", err)
		}
		next := time.Date(now.Year(), now.Month(), now.Day(), h, m, s, 0, time.UTC)
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}
		return next.Format(time.RFC3339), nil
	case "monthly":
		h, m, s, err := parseHMS(p.DailyTime)
		if err != nil {
			return "", fmt.Errorf("daily_time invalid: %w", err)
		}
		if p.MonthlyDay < 1 || p.MonthlyDay > 31 {
			return "", fmt.Errorf("monthly_day must be 1..31")
		}
		y, mon := now.Year(), now.Month()
		next := monthlyAt(y, mon, p.MonthlyDay, h, m, s)
		if !next.After(now) {
			mon++
			if mon > 12 {
				mon = 1
				y++
			}
			next = monthlyAt(y, mon, p.MonthlyDay, h, m, s)
		}
		return next.Format(time.RFC3339), nil
	default:
		if p.IntervalMinutes <= 0 {
			return "", fmt.Errorf("interval_minutes must be > 0")
		}
		return now.Add(time.Duration(p.IntervalMinutes) * time.Minute).Format(time.RFC3339), nil
	}
}

func monthlyAt(year int, month time.Month, day, hour, minute, second int) time.Time {
	maxDay := daysInMonth(year, month)
	if day > maxDay {
		day = maxDay
	}
	return time.Date(year, month, day, hour, minute, second, 0, time.UTC)
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func parseHMS(v string) (int, int, int, error) {
	parts := strings.Split(strings.TrimSpace(v), ":")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("must be HH:MM:SS")
	}
	vals := make([]int, 3)
	for i := 0; i < 3; i++ {
		n, err := atoi(parts[i])
		if err != nil {
			return 0, 0, 0, err
		}
		vals[i] = n
	}
	if vals[0] < 0 || vals[0] > 23 || vals[1] < 0 || vals[1] > 59 || vals[2] < 0 || vals[2] > 59 {
		return 0, 0, 0, fmt.Errorf("must be valid time")
	}
	return vals[0], vals[1], vals[2], nil
}

func atoi(v string) (int, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, fmt.Errorf("empty")
	}
	n := 0
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid number")
		}
		n = n*10 + int(ch-'0')
	}
	return n, nil
}

func normalizeScheduleKind(kind string) string {
	k := strings.ToLower(strings.TrimSpace(kind))
	switch k {
	case "daily", "monthly":
		return k
	default:
		return "interval"
	}
}

func nullIfZero(v int) any {
	if v == 0 {
		return nil
	}
	return v
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

