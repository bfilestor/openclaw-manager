package gateway

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
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

type GatewayDeepStatus struct {
	Service    *ServiceStatus `json:"service"`
	BindAddr   string         `json:"bind_addr"`
	Port       int            `json:"port"`
	LogPath    string         `json:"log_path"`
	NodePath   string         `json:"node_path"`
	NVMWarning bool           `json:"nvm_warning"`
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

func (s *SystemctlService) Start(service string) error   { return s.runAction("start", service) }
func (s *SystemctlService) Stop(service string) error    { return s.runAction("stop", service) }
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

func (s *SystemctlService) DeepStatus(service string) (*GatewayDeepStatus, error) {
	if !validServiceName(service) {
		return nil, ErrInvalidServiceName
	}

	result := &GatewayDeepStatus{}
	var statusErr error
	var openclawErr error
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(2)

	go func() {
		defer wg.Done()
		st, err := s.Status(service)
		mu.Lock()
		result.Service = st
		statusErr = err
		mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		parsed, err := s.fetchOpenclawDeepStatus()
		mu.Lock()
		if parsed != nil {
			result.BindAddr = parsed.BindAddr
			result.Port = parsed.Port
			result.LogPath = parsed.LogPath
			result.NodePath = parsed.NodePath
			result.NVMWarning = parsed.NVMWarning
		}
		openclawErr = err
		mu.Unlock()
	}()

	wg.Wait()

	switch {
	case statusErr == nil && openclawErr == nil:
		return result, nil
	case statusErr != nil && openclawErr != nil:
		return result, errors.Join(statusErr, openclawErr)
	case statusErr != nil:
		return result, statusErr
	default:
		return result, openclawErr
	}
}

func (s *SystemctlService) fetchOpenclawDeepStatus() (*GatewayDeepStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	out, err := s.exec.Run(ctx, "openclaw", "gateway", "status", "--deep")
	parsed := parseOpenclawDeepOutput(string(out))

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return parsed, ErrCommandTimeout
	}
	if err != nil {
		return parsed, fmt.Errorf("openclaw gateway status --deep failed: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return parsed, nil
}

func parseOpenclawDeepOutput(out string) *GatewayDeepStatus {
	result := &GatewayDeepStatus{}
	for _, raw := range strings.Split(out, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		// Newer human-readable output format.
		if strings.HasPrefix(line, "Gateway:") {
			if addr, port := parseGatewaySummary(line); addr != "" || port > 0 {
				if addr != "" {
					result.BindAddr = addr
				}
				if port > 0 {
					result.Port = port
				}
			}
			continue
		}
		if strings.HasPrefix(line, "Listening:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "Listening:"))
			if addr, port := parseBindAddress(value); addr != "" || port > 0 {
				if addr != "" {
					result.BindAddr = addr
				}
				if port > 0 {
					result.Port = port
				}
			}
			continue
		}
		if strings.HasPrefix(line, "File logs:") {
			result.LogPath = strings.TrimSpace(strings.TrimPrefix(line, "File logs:"))
			continue
		}
		if strings.HasPrefix(line, "Command:") {
			cmd := strings.TrimSpace(strings.TrimPrefix(line, "Command:"))
			if cmd != "" {
				fields := strings.Fields(cmd)
				if len(fields) > 0 {
					result.NodePath = fields[0]
					result.NVMWarning = strings.Contains(strings.ToLower(result.NodePath), ".nvm")
				}
			}
			continue
		}

		key, value, ok := parseLineKeyValue(line)
		if !ok {
			continue
		}

		switch key {
		case "bind", "bindaddr", "bindaddress":
			addr, port := parseBindAddress(value)
			if addr != "" {
				result.BindAddr = addr
			}
			if port > 0 {
				result.Port = port
			}
		case "log", "logpath", "logfile", "log_path", "log_file":
			result.LogPath = value
		case "node", "nodepath", "node_path":
			result.NodePath = value
			result.NVMWarning = strings.Contains(strings.ToLower(value), ".nvm")
		}
	}
	return result
}

func parseLineKeyValue(line string) (string, string, bool) {
	idx := strings.IndexAny(line, "=:")
	if idx <= 0 {
		return "", "", false
	}

	key := strings.ToLower(strings.TrimSpace(line[:idx]))
	key = strings.ReplaceAll(key, " ", "")
	value := strings.TrimSpace(line[idx+1:])
	value = strings.Trim(value, `"'`)
	if value == "" {
		return "", "", false
	}
	return key, value, true
}

func parseBindAddress(value string) (string, int) {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return "", 0
	}
	if idx := strings.IndexAny(clean, " \t,"); idx > 0 {
		clean = clean[:idx]
	}
	if clean == "-" || clean == "-:-" || strings.EqualFold(clean, "n/a") || strings.EqualFold(clean, "unknown") {
		return "", 0
	}

	host, portStr, err := net.SplitHostPort(clean)
	if err == nil {
		port, convErr := strconv.Atoi(portStr)
		if convErr != nil {
			return host, 0
		}
		return host, port
	}

	parts := strings.Split(clean, ":")
	if len(parts) == 2 {
		port, convErr := strconv.Atoi(parts[1])
		if convErr != nil {
			return parts[0], 0
		}
		return parts[0], port
	}

	return clean, 0
}

func parseGatewaySummary(line string) (string, int) {
	// Example: "Gateway: bind=loopback (127.0.0.1), port=18789 (service args)"
	addr := ""
	port := 0
	if m := regexp.MustCompile(`\(([^)]+)\)`).FindStringSubmatch(line); len(m) == 2 {
		addr = strings.TrimSpace(m[1])
	}
	if m := regexp.MustCompile(`\bport\s*=\s*(\d+)`).FindStringSubmatch(strings.ToLower(line)); len(m) == 2 {
		if p, err := strconv.Atoi(m[1]); err == nil {
			port = p
		}
	}
	if addr == "" {
		if m := regexp.MustCompile(`\bbind\s*=\s*([^,\s]+)`).FindStringSubmatch(strings.ToLower(line)); len(m) == 2 {
			addr = strings.TrimSpace(m[1])
		}
	}
	if a, p := parseBindAddress(fmt.Sprintf("%s:%d", addr, port)); a != "" || p > 0 {
		if a != "" {
			addr = a
		}
		if p > 0 {
			port = p
		}
	}
	if addr == "loopback" {
		addr = "127.0.0.1"
	}
	if addr == "all" || addr == "0.0.0.0" {
		addr = "0.0.0.0"
	}
	if addr == "-" || addr == "" {
		addr = ""
	}
	return addr, port
}

var serviceNameRe = regexp.MustCompile(`^[a-zA-Z0-9_.@-]+$`)

func validServiceName(name string) bool {
	if strings.TrimSpace(name) == "" || strings.Contains(name, "..") || strings.ContainsAny(name, `/\\`) {
		return false
	}
	return serviceNameRe.MatchString(name)
}
