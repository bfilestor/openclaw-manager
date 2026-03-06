package task

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("task not found")

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(t *Task) error {
	if t.Status == "" {
		t.Status = StatusPending
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(`INSERT INTO tasks(task_id,task_type,status,request_json,exit_code,stdout_tail,stderr_tail,log_path,created_by,created_at,started_at,finished_at)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		t.TaskID, t.TaskType, string(t.Status), t.RequestJSON, toNullInt(t.ExitCode), t.StdoutTail, t.StderrTail, t.LogPath, t.CreatedBy,
		t.CreatedAt.Format(time.RFC3339), toNullTime(t.StartedAt), toNullTime(t.FinishedAt))
	return err
}

func (r *Repository) FindByID(id string) (*Task, error) {
	row := r.db.QueryRow(`SELECT task_id,task_type,status,request_json,exit_code,stdout_tail,stderr_tail,log_path,created_by,created_at,started_at,finished_at FROM tasks WHERE task_id=?`, id)
	t, err := scanTask(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

func (r *Repository) UpdateStatus(taskID string, status Status) error {
	now := time.Now().UTC().Format(time.RFC3339)
	var q string
	switch status {
	case StatusRunning:
		q = `UPDATE tasks SET status=?, started_at=? WHERE task_id=?`
		_, err := r.db.Exec(q, string(status), now, taskID)
		return checkAffected(err, r.db, taskID)
	case StatusSucceeded, StatusFailed, StatusCanceled:
		q = `UPDATE tasks SET status=?, finished_at=? WHERE task_id=?`
		_, err := r.db.Exec(q, string(status), now, taskID)
		return checkAffected(err, r.db, taskID)
	default:
		q = `UPDATE tasks SET status=? WHERE task_id=?`
		_, err := r.db.Exec(q, string(status), taskID)
		return checkAffected(err, r.db, taskID)
	}
}

func (r *Repository) UpdateResult(taskID string, exitCode *int, stdoutTail, stderrTail, logPath string) error {
	res, err := r.db.Exec(`UPDATE tasks SET exit_code=?, stdout_tail=?, stderr_tail=?, log_path=? WHERE task_id=?`, toNullInt(exitCode), stdoutTail, stderrTail, logPath, taskID)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) List(f ListFilter) ([]*Task, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	where := make([]string, 0, 3)
	args := make([]any, 0, 5)
	if f.Status != "" {
		where = append(where, "status=?")
		args = append(args, string(f.Status))
	}
	if strings.TrimSpace(f.TaskType) != "" {
		where = append(where, "task_type=?")
		args = append(args, f.TaskType)
	}
	if strings.TrimSpace(f.CreatedBy) != "" {
		where = append(where, "created_by=?")
		args = append(args, f.CreatedBy)
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	var total int
	countQ := `SELECT COUNT(1) FROM tasks` + whereSQL
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ := `SELECT task_id,task_type,status,request_json,exit_code,stdout_tail,stderr_tail,log_path,created_by,created_at,started_at,finished_at FROM tasks` + whereSQL + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args2 := append(args, f.Limit, f.Offset)
	rows, err := r.db.Query(listQ, args2...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]*Task, 0)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, total, rows.Err()
}

func (r *Repository) Delete(taskID string) error {
	res, err := r.db.Exec(`DELETE FROM tasks WHERE task_id=?`, taskID)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ClearAll() (int64, error) {
	res, err := r.db.Exec(`DELETE FROM tasks`)
	if err != nil {
		return 0, err
	}
	aff, _ := res.RowsAffected()
	return aff, nil
}

func (r *Repository) ClearByCreatedBy(userID string) (int64, error) {
	res, err := r.db.Exec(`DELETE FROM tasks WHERE created_by=?`, userID)
	if err != nil {
		return 0, err
	}
	aff, _ := res.RowsAffected()
	return aff, nil
}

type scanner interface{ Scan(dest ...any) error }

func scanTask(s scanner) (*Task, error) {
	var (
		t                     Task
		status, createdAt     string
		exitCode              sql.NullInt64
		startedAt, finishedAt sql.NullString
	)
	if err := s.Scan(&t.TaskID, &t.TaskType, &status, &t.RequestJSON, &exitCode, &t.StdoutTail, &t.StderrTail, &t.LogPath, &t.CreatedBy, &createdAt, &startedAt, &finishedAt); err != nil {
		return nil, err
	}
	t.Status = Status(status)
	ct, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	t.CreatedAt = ct
	if exitCode.Valid {
		v := int(exitCode.Int64)
		t.ExitCode = &v
	}
	if startedAt.Valid {
		if st, e := time.Parse(time.RFC3339, startedAt.String); e == nil {
			t.StartedAt = &st
		}
	}
	if finishedAt.Valid {
		if ft, e := time.Parse(time.RFC3339, finishedAt.String); e == nil {
			t.FinishedAt = &ft
		}
	}
	return &t, nil
}

func toNullInt(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Valid: true, Int64: int64(*v)}
}

func toNullTime(v *time.Time) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{Valid: true, String: v.Format(time.RFC3339)}
}

func checkAffected(err error, db *sql.DB, taskID string) error {
	if err != nil {
		return err
	}
	var c int
	if e := db.QueryRow(`SELECT COUNT(1) FROM tasks WHERE task_id=?`, taskID).Scan(&c); e != nil {
		return e
	}
	if c == 0 {
		return ErrNotFound
	}
	return nil
}
