package backup

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	DB         *sql.DB
	BackupHome string
	OpenclawHome string
	ManagerHome  string
}

type Manifest struct {
	BackupID  string   `json:"backup_id"`
	Label     string   `json:"label"`
	Scope     []string `json:"scope"`
	Paths     []string `json:"paths"`
	SHA256    string   `json:"sha256"`
	CreatedAt string   `json:"created_at"`
	CreatedBy string   `json:"created_by"`
}

func (s *Service) Create(scope []string, label, createdBy string) (string, error) {
	id := uuid.NewString()
	if err := os.MkdirAll(s.BackupHome, 0o755); err != nil { return "", err }
	archivePath := filepath.Join(s.BackupHome, id+".tar.gz")
	paths := s.resolveScope(scope)
	if err := makeTarGz(archivePath, paths); err != nil { return "", err }
	sha, size, err := fileSHA256AndSize(archivePath)
	if err != nil { return "", err }
	mf := Manifest{BackupID: id, Label: label, Scope: scope, Paths: paths, SHA256: sha, CreatedAt: time.Now().UTC().Format(time.RFC3339), CreatedBy: createdBy}
	mb, _ := json.MarshalIndent(mf, "", "  ")
	manifestPath := filepath.Join(s.BackupHome, id+".manifest.json")
	if err := os.WriteFile(manifestPath, mb, 0o644); err != nil { return "", err }
	_, err = s.DB.Exec(`INSERT INTO backups(backup_id,label,scope_json,manifest_path,size_bytes,sha256,verified,created_by,created_at) VALUES(?,?,?,?,?,?,?,?,?)`,
		id, label, string(mb), manifestPath, size, sha, 1, nullIfEmpty(createdBy), time.Now().UTC().Format(time.RFC3339))
	if err != nil { return "", err }
	return id, nil
}

func (s *Service) resolveScope(scope []string) []string {
	set := map[string]bool{}
	for _, x := range scope { set[strings.ToLower(strings.TrimSpace(x))] = true }
	out := make([]string, 0)
	if set["openclaw_json"] { out = append(out, filepath.Join(s.OpenclawHome, "openclaw.json")) }
	if set["global_skills"] { out = append(out, filepath.Join(s.OpenclawHome, "skills")) }
	if set["workspaces"] { out = append(out, filepath.Join(s.OpenclawHome, "workspace")) }
	if set["user_systemd_unit"] { home, _ := os.UserHomeDir(); out = append(out, filepath.Join(home, ".config/systemd/user/openclaw-gateway.service")) }
	if set["manager_revisions"] { out = append(out, filepath.Join(s.ManagerHome, "revisions")) }
	return out
}

func makeTarGz(dst string, paths []string) error {
	f, err := os.Create(dst)
	if err != nil { return err }
	defer f.Close()
	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	for _, p := range paths {
		st, err := os.Stat(p)
		if err != nil { continue }
		if st.IsDir() {
			_ = filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() { return nil }
				return addFileToTar(tw, path)
			})
		} else {
			if err := addFileToTar(tw, p); err != nil { return err }
		}
	}
	return nil
}

func addFileToTar(tw *tar.Writer, path string) error {
	st, err := os.Stat(path)
	if err != nil { return err }
	h, _ := tar.FileInfoHeader(st, "")
	h.Name = strings.TrimPrefix(path, "/")
	if err := tw.WriteHeader(h); err != nil { return err }
	f, err := os.Open(path)
	if err != nil { return err }
	defer f.Close()
	_, err = io.Copy(tw, f)
	return err
}

func fileSHA256AndSize(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil { return "", 0, err }
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil { return "", 0, err }
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func nullIfEmpty(v string) any { if v=="" { return nil }; return v }
