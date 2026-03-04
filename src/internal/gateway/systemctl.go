package gateway

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidServiceName = errors.New("invalid service name")
	ErrCommandTimeout     = errors.New("command timeout")
)

type ServiceStatus struct {
	ActiveState          string `json:"active_state"`
	SubState             string `json:"sub_state"`
	MainPID              string `json:"main_pid"`
	ExecStart            string `json:"exec_start"`
	FragmentPath         string `json:"fragment_path"`
	ActiveEnterTimestamp string `json:"active_enter_timestamp"`
}

type Executor interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type OSExecutor struct{}

func (OSExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

type SystemctlService struct {
	exec    Executor
	timeout time.Duration
}

func NewSystemctlService(exec Executor) *SystemctlService {
	if exec == nil {
		exec = OSExecutor{}
	}
	return &SystemctlService{exec: exec, timeout: 30 * time.Second}
}

func (s *SystemctlService) Start(service string) error { return s.runAction("start", service) }
func (s *SystemctlService) Stop(service string) error { return s.runAction("stop", service) }
func (s *SystemctlService) Restart(service string) error { return s.runAction("restart", service) }

func (s *SystemctlService) runAction(action, service string) error {
	if !validServiceName(service) {
		return ErrInvalidServiceName
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	_, err := s.exec.Run(ctx, "systemctl", "--user", action, service)
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return ErrCommandTimeout
	}
	return err
}

func (s *SystemctlService) Status(service string) (*ServiceStatus, error) {
	if !validServiceName(service) {
		return nil, ErrInvalidServiceName
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	out, err := s.exec.Run(ctx, "systemctl", "--user", "show", service,
		"--no-page",
		"--property=ActiveState,SubState,MainPID,ExecStart,FragmentPath,ActiveEnterTimestamp",
	)
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return nil, ErrCommandTimeout
	}
	if err != nil {
		return nil, fmt.Errorf("systemctl show failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	st := &ServiceStatus{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		k, v := kv[0], kv[1]
		switch k {
		case "ActiveState":
			st.ActiveState = v
		case "SubState":
			st.SubState = v
		case "MainPID":
			st.MainPID = v
		case "ExecStart":
			st.ExecStart = v
		case "FragmentPath":
			st.FragmentPath = v
		case "ActiveEnterTimestamp":
			st.ActiveEnterTimestamp = v
		}
	}
	return st, nil
}

var serviceNameRe = regexp.MustCompile(`^[a-zA-Z0-9_.@-]+$`)

func validServiceName(name string) bool {
	if strings.TrimSpace(name) == "" || strings.Contains(name, "..") || strings.ContainsAny(name, `/\\`) {
		return false
	}
	return serviceNameRe.MatchString(name)
}
