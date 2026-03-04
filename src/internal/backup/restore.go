package backup

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RestoreReport struct {
	BackupID       string   `json:"backup_id"`
	WillOverwrite  []string `json:"will_overwrite"`
	DryRun         bool     `json:"dry_run"`
}

func (s *Service) Restore(backupID string, dryRun bool, restartGateway bool, createdBy string) (*RestoreReport, error) {
	archive := filepath.Join(s.BackupHome, backupID+".tar.gz")
	manifestPath := filepath.Join(s.BackupHome, backupID+".manifest.json")
	if _, err := os.Stat(archive); err != nil { return nil, err }
	mb, err := os.ReadFile(manifestPath)
	if err != nil { return nil, err }
	var mf Manifest
	if err := json.Unmarshal(mb, &mf); err != nil { return nil, err }
	sha, _, err := fileSHA256AndSize(archive)
	if err != nil { return nil, err }
	if mf.SHA256 != "" && sha != mf.SHA256 {
		return nil, fmt.Errorf("backup sha256 mismatch")
	}
	files, err := listArchiveFiles(archive)
	if err != nil { return nil, err }
	report := &RestoreReport{BackupID: backupID, WillOverwrite: files, DryRun: dryRun}
	if dryRun {
		return report, nil
	}
	// pre-restore snapshot
	_, _ = s.Create([]string{"openclaw_json", "global_skills", "workspaces", "manager_revisions"}, "pre-restore-"+backupID, createdBy)
	allowed := s.allowedRestorePrefixes()
	if err := extractArchiveToRoot(archive, allowed); err != nil { return nil, err }
	if restartGateway {
		// MVP: just annotate; actual restart handled by gateway API in integration stage
	}
	return report, nil
}

func listArchiveFiles(archive string) ([]string, error) {
	f, err := os.Open(archive)
	if err != nil { return nil, err }
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil { return nil, err }
	defer gr.Close()
	tr := tar.NewReader(gr)
	out := make([]string, 0)
	for {
		h, err := tr.Next()
		if err == io.EOF { break }
		if err != nil { return nil, err }
		if h.FileInfo().IsDir() { continue }
		out = append(out, "/"+strings.TrimPrefix(h.Name, "/"))
	}
	return out, nil
}

func extractArchiveToRoot(archive string, allowedPrefixes []string) error {
	f, err := os.Open(archive)
	if err != nil { return err }
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil { return err }
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err == io.EOF { break }
		if err != nil { return err }
		if h.FileInfo().IsDir() { continue }
		target := "/" + strings.TrimPrefix(h.Name, "/")
		if !isAllowedRestoreTarget(target, allowedPrefixes) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil { return err }
		tf, err := os.Create(target)
		if err != nil { return err }
		if _, err := io.Copy(tf, tr); err != nil { tf.Close(); return err }
		tf.Close()
	}
	return nil
}

func (s *Service) getBackupMeta(backupID string) (string, error) {
	var manifestPath string
	err := s.DB.QueryRow(`SELECT manifest_path FROM backups WHERE backup_id=?`, backupID).Scan(&manifestPath)
	if err != nil { return "", err }
	return manifestPath, nil
}

func sha256Bytes(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func insertRestoreAudit(db *sql.DB, backupID, createdBy string) {
	_, _ = db.Exec(`INSERT INTO tasks(task_id,task_type,status,request_json,created_by,created_at) VALUES(?,?,?,?,?,?)`,
		"restore-"+backupID+"-"+time.Now().UTC().Format("150405"), "backup.restore", "SUCCEEDED", "{}", nullIfEmpty(createdBy), time.Now().UTC().Format(time.RFC3339))
}

func (s *Service) allowedRestorePrefixes() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Clean(s.OpenclawHome),
		filepath.Clean(s.ManagerHome),
		filepath.Clean(filepath.Join(home, ".config/systemd/user")),
	}
}

func isAllowedRestoreTarget(target string, prefixes []string) bool {
	clean := filepath.Clean(target)
	for _, p := range prefixes {
		if p == "" {
			continue
		}
		if clean == p || strings.HasPrefix(clean, p+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}
