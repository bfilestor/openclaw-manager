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
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	DB           *sql.DB
	BackupHome   string
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

func (s *Service) Create(scope []string, label, createdBy string, planIDs ...string) (string, error) {
	id := uuid.NewString()
	if err := os.MkdirAll(s.BackupHome, 0o755); err != nil {
		return "", err
	}
	archivePath := filepath.Join(s.BackupHome, id+".tar.gz")
	paths := s.resolveScope(scope)
	if err := makeTarGz(archivePath, paths); err != nil {
		return "", err
	}
	sha, size, err := fileSHA256AndSize(archivePath)
	if err != nil {
		return "", err
	}
	mf := Manifest{BackupID: id, Label: label, Scope: scope, Paths: paths, SHA256: sha, CreatedAt: time.Now().UTC().Format(time.RFC3339), CreatedBy: createdBy}
	mb, _ := json.MarshalIndent(mf, "", "  ")
	manifestPath := filepath.Join(s.BackupHome, id+".manifest.json")
	if err := os.WriteFile(manifestPath, mb, 0o644); err != nil {
		return "", err
	}
	var planID any
	if len(planIDs) > 0 && strings.TrimSpace(planIDs[0]) != "" {
		planID = planIDs[0]
	}
	_, err = s.DB.Exec(`INSERT INTO backups(backup_id,label,scope_json,manifest_path,size_bytes,sha256,verified,created_by,created_at,plan_id) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		id, label, string(mb), manifestPath, size, sha, 1, nullIfEmpty(createdBy), time.Now().UTC().Format(time.RFC3339), planID)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) resolveScope(scope []string) []string {
	set := map[string]bool{}
	for _, x := range scope {
		set[strings.ToLower(strings.TrimSpace(x))] = true
	}
	out := make([]string, 0)
	if set["openclaw_json"] {
		out = append(out, s.resolveOpenclawCoreConfigPaths()...)
	}
	if set["global_skills"] {
		out = append(out, filepath.Join(s.OpenclawHome, "skills"))
	}
	if set["workspaces"] {
		out = append(out, s.resolveWorkspacesFromOpenClawJSON()...)
	}
	if set["plugins"] {
		out = append(out, s.resolvePluginPathsFromOpenClawJSON()...)
	}
	if set["user_systemd_unit"] {
		home, _ := os.UserHomeDir()
		out = append(out, filepath.Join(home, ".config/systemd/user/openclaw-gateway.service"))
	}
	if set["manager_db"] {
		out = append(out, s.resolveManagerDBPaths()...)
	}
	return uniqPaths(out)
}

func (s *Service) resolveManagerDBPaths() []string {
	base := ""
	if strings.TrimSpace(s.ManagerHome) != "" {
		base = filepath.Join(s.ManagerHome, "manager.db")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			base = filepath.Join(".", "manager.db")
		} else {
			base = filepath.Join(home, ".openclaw-manager", "manager.db")
		}
	}
	return []string{base, base + "-wal", base + "-shm"}
}

func (s *Service) resolveOpenclawCoreConfigPaths() []string {
	return uniqPaths([]string{
		filepath.Join(s.OpenclawHome, "openclaw.json"),
		filepath.Join(s.OpenclawHome, "models.json"),
		filepath.Join(s.OpenclawHome, "agents", "main", "agent", "models.json"),
	})
}

type openclawConfig struct {
	Agents struct {
		Defaults struct {
			Workspace string `json:"workspace"`
		} `json:"defaults"`
		List []struct {
			ID        string `json:"id"`
			Workspace string `json:"workspace"`
		} `json:"list"`
	} `json:"agents"`
	Plugins struct {
		Installs map[string]struct {
			InstallPath string `json:"installPath"`
		} `json:"installs"`
	} `json:"plugins"`
}

func (s *Service) loadOpenclawConfig() (*openclawConfig, error) {
	raw, err := os.ReadFile(filepath.Join(s.OpenclawHome, "openclaw.json"))
	if err != nil {
		return nil, err
	}
	var cfg openclawConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *Service) resolveWorkspacesFromOpenClawJSON() []string {
	defaultWorkspace := filepath.Join(s.OpenclawHome, "workspace")
	cfg, err := s.loadOpenclawConfig()
	if err != nil {
		return []string{defaultWorkspace}
	}
	if ws := strings.TrimSpace(cfg.Agents.Defaults.Workspace); ws != "" {
		defaultWorkspace = ws
	}

	workspaces := []string{defaultWorkspace}
	baseDir := filepath.Dir(defaultWorkspace)
	for _, it := range cfg.Agents.List {
		agentID := strings.TrimSpace(it.ID)
		if agentID == "" {
			continue
		}
		if ws := strings.TrimSpace(it.Workspace); ws != "" {
			workspaces = append(workspaces, ws)
			continue
		}
		if agentID == "main" {
			workspaces = append(workspaces, defaultWorkspace)
			continue
		}
		workspaces = append(workspaces, filepath.Join(baseDir, "workspace-"+agentID))
	}
	return uniqPaths(workspaces)
}

func (s *Service) resolvePluginPathsFromOpenClawJSON() []string {
	fallback := filepath.Join(s.OpenclawHome, "extensions")
	cfg, err := s.loadOpenclawConfig()
	if err != nil {
		return []string{fallback}
	}
	paths := []string{fallback}
	for _, install := range cfg.Plugins.Installs {
		if p := strings.TrimSpace(install.InstallPath); p != "" {
			paths = append(paths, p)
		}
	}
	return uniqPaths(paths)
}

func uniqPaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		clean := filepath.Clean(strings.TrimSpace(p))
		if clean == "." || clean == "" {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		out = append(out, clean)
	}
	sort.Strings(out)
	return out
}

func makeTarGz(dst string, paths []string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	gw := gzip.NewWriter(f)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	for _, p := range paths {
		st, err := os.Stat(p)
		if err != nil {
			continue
		}
		if st.IsDir() {
			_ = filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				return addFileToTar(tw, path)
			})
		} else {
			if err := addFileToTar(tw, p); err != nil {
				return err
			}
		}
	}
	return nil
}

func addFileToTar(tw *tar.Writer, path string) error {
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	h, _ := tar.FileInfoHeader(st, "")
	h.Name = strings.TrimPrefix(path, "/")
	if err := tw.WriteHeader(h); err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(tw, f)
	return err
}

func fileSHA256AndSize(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}
