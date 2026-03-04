package task

import (
	"fmt"
	"net/http"
	"strings"
)

type SSEHandler struct {
	Repo *Repository
}

func (h *SSEHandler) TaskEvents(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			token = strings.TrimSpace(auth[7:])
		}
	}
	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
		return
	}
	taskID := taskIDFromEventsPath(r.URL.Path)
	if taskID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	t, err := h.Repo.FindByID(taskID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	seq := 1
	for _, line := range splitLines(t.StdoutTail) {
		_, _ = fmt.Fprintf(w, "data: {\"seq\":%d,\"stream\":\"stdout\",\"line\":%q}\n\n", seq, line)
		seq++
	}
	for _, line := range splitLines(t.StderrTail) {
		_, _ = fmt.Fprintf(w, "data: {\"seq\":%d,\"stream\":\"stderr\",\"line\":%q}\n\n", seq, line)
		seq++
	}
	exitCode := 0
	if t.ExitCode != nil {
		exitCode = *t.ExitCode
	}
	_, _ = fmt.Fprintf(w, "data: {\"type\":\"done\",\"status\":%q,\"exit_code\":%d}\n\n", t.Status, exitCode)
}

func splitLines(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func taskIDFromEventsPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "tasks" && i+2 < len(parts) && parts[i+2] == "events" {
			return parts[i+1]
		}
	}
	return ""
}
