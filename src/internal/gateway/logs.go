package gateway

import (
	"bufio"
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"openclaw-manager/internal/middleware"
	"openclaw-manager/internal/storage"
)

type LogsHandler struct {
	Exec      Executor
	Validator *storage.PathValidator
	Timeout   time.Duration
}

func NewLogsHandler(exec Executor) *LogsHandler {
	v, _ := storage.NewPathValidator([]string{"/tmp/openclaw"})
	if exec == nil {
		exec = OSExecutor{}
	}
	return &LogsHandler{Exec: exec, Validator: v, Timeout: 30 * time.Second}
}

func (h *LogsHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	lines := 200
	if raw := strings.TrimSpace(r.URL.Query().Get("lines")); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"lines": "must be >= 0"}))
			return
		}
		lines = n
	}
	if lines > 1000 {
		lines = 1000
	}
	source := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("source")))
	if source == "" {
		source = "file"
	}

	var out []string
	var err error
	switch source {
	case "file":
		out, err = h.readFileLogs(lines)
	case "journald":
		out, err = h.readJournalLogs(lines)
	default:
		middleware.WriteAppError(w, middleware.NewValidation(map[string]string{"source": "must be file or journald"}))
		return
	}
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			out = []string{}
		} else {
			middleware.WriteAppError(w, err)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"logs":[` + quoteLines(out) + `],"source":"` + source + `"}`))
}

func (h *LogsHandler) readFileLogs(lines int) ([]string, error) {
	if lines == 0 {
		return []string{}, nil
	}
	file, err := latestLogFile("/tmp/openclaw")
	if err != nil {
		return nil, err
	}
	if _, err := h.Validator.Validate(file); err != nil {
		return nil, err
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	all := make([]string, 0, lines)
	s := bufio.NewScanner(f)
	for s.Scan() {
		all = append(all, s.Text())
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if len(all) <= lines {
		return all, nil
	}
	return all[len(all)-lines:], nil
}

func (h *LogsHandler) readJournalLogs(lines int) ([]string, error) {
	if lines == 0 {
		return []string{}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), h.Timeout)
	defer cancel()
	out, err := h.Exec.Run(ctx, "journalctl", "--user", "-u", "openclaw-gateway.service", "-n", strconv.Itoa(lines), "--no-pager", "--output=cat")
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, ErrCommandTimeout
	}
	if err != nil {
		return nil, err
	}
	trim := strings.TrimSpace(string(out))
	if trim == "" {
		return []string{}, nil
	}
	return strings.Split(trim, "\n"), nil
}

func latestLogFile(base string) (string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return "", err
	}
	files := make([]string, 0)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasPrefix(n, "openclaw-") && strings.HasSuffix(n, ".log") {
			files = append(files, filepath.Join(base, n))
		}
	}
	if len(files) == 0 {
		return "", os.ErrNotExist
	}
	sort.Strings(files)
	return files[len(files)-1], nil
}

func quoteLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	b := strings.Builder{}
	for i, l := range lines {
		if i > 0 {
			b.WriteByte(',')
		}
		esc := strings.ReplaceAll(l, "\\", "\\\\")
		esc = strings.ReplaceAll(esc, `"`, `\"`)
		b.WriteByte('"')
		b.WriteString(esc)
		b.WriteByte('"')
	}
	return b.String()
}
