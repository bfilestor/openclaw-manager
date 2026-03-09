package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
	"unicode"

	"openclaw-manager/internal/middleware"
)

type shellCommandExecutor interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type osShellCommandExecutor struct{}

func (osShellCommandExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

type ShellHandler struct {
	Exec    shellCommandExecutor
	Timeout time.Duration
}

type executeShellReq struct {
	Command string `json:"command"`
}

type executeShellResp struct {
	Command    string `json:"command"`
	Output     string `json:"output"`
	ExitCode   int    `json:"exit_code"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	DurationMS int64  `json:"duration_ms"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
}

func NewShellHandler(exec shellCommandExecutor) *ShellHandler {
	if exec == nil {
		exec = osShellCommandExecutor{}
	}
	return &ShellHandler{
		Exec:    exec,
		Timeout: 2 * time.Minute,
	}
}

func (h *ShellHandler) Execute(w http.ResponseWriter, r *http.Request) {
	var req executeShellReq
	if err := middleware.BindJSON(r, &req); err != nil {
		middleware.WriteAppError(w, err)
		return
	}

	args, parseErr := parseOpenclawCommand(req.Command)
	if parseErr != nil {
		middleware.WriteAppError(w, parseErr)
		return
	}

	timeout := h.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	startedAt := time.Now().UTC()
	out, runErr := h.Exec.Run(ctx, args[0], args[1:]...)
	finishedAt := time.Now().UTC()

	resp := executeShellResp{
		Command:    strings.TrimSpace(req.Command),
		Output:     string(out),
		ExitCode:   0,
		Success:    true,
		DurationMS: finishedAt.Sub(startedAt).Milliseconds(),
		StartedAt:  startedAt.Format(time.RFC3339),
		FinishedAt: finishedAt.Format(time.RFC3339),
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		resp.Success = false
		resp.ExitCode = -1
		resp.Error = "command timeout"
	} else if runErr != nil {
		resp.Success = false
		resp.ExitCode = exitCodeFromErr(runErr)
		resp.Error = runErr.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func parseOpenclawCommand(raw string) ([]string, *middleware.AppError) {
	cmd := strings.TrimSpace(raw)
	if cmd == "" {
		return nil, middleware.NewValidation(map[string]string{"command": "required"})
	}
	if strings.ContainsAny(cmd, "\r\n") {
		return nil, middleware.NewValidation(map[string]string{"command": "must be single line"})
	}

	args, err := splitCommandLine(cmd)
	if err != nil {
		return nil, middleware.NewValidation(map[string]string{"command": err.Error()})
	}
	if len(args) == 0 {
		return nil, middleware.NewValidation(map[string]string{"command": "required"})
	}
	if args[0] != "openclaw" {
		return nil, middleware.NewValidation(map[string]string{"command": "must start with openclaw"})
	}
	return args, nil
}

func splitCommandLine(input string) ([]string, error) {
	var args []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escaping := false
	tokenStarted := false

	flush := func() {
		if !tokenStarted && current.Len() == 0 {
			return
		}
		args = append(args, current.String())
		current.Reset()
		tokenStarted = false
	}

	for _, r := range input {
		switch {
		case escaping:
			current.WriteRune(r)
			tokenStarted = true
			escaping = false
		case r == '\\' && !inSingleQuote:
			escaping = true
			tokenStarted = true
		case r == '\'' && !inDoubleQuote:
			inSingleQuote = !inSingleQuote
			tokenStarted = true
		case r == '"' && !inSingleQuote:
			inDoubleQuote = !inDoubleQuote
			tokenStarted = true
		case unicode.IsSpace(r) && !inSingleQuote && !inDoubleQuote:
			flush()
		default:
			current.WriteRune(r)
			tokenStarted = true
		}
	}

	if escaping {
		return nil, fmt.Errorf("invalid escape at command end")
	}
	if inSingleQuote || inDoubleQuote {
		return nil, fmt.Errorf("unterminated quoted string")
	}
	flush()

	return args, nil
}

func exitCodeFromErr(err error) int {
	if err == nil {
		return 0
	}

	type exitCoder interface {
		ExitCode() int
	}
	var ec exitCoder
	if errors.As(err, &ec) {
		return ec.ExitCode()
	}
	return -1
}
