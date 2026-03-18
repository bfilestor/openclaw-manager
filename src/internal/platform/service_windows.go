//go:build windows

package platform

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type WindowsServiceController struct {
	exec    CommandExecutor
	timeout time.Duration
}

func NewDefaultServiceController(exec CommandExecutor) ServiceController {
	if exec == nil {
		exec = NewDefaultExecutor()
	}
	return &WindowsServiceController{exec: exec, timeout: 30 * time.Second}
}

func (s *WindowsServiceController) Start(service string) error { return s.runAction("start", service) }
func (s *WindowsServiceController) Stop(service string) error  { return s.runAction("stop", service) }
func (s *WindowsServiceController) Restart(service string) error {
	return s.runAction("restart", service)
}

func (s *WindowsServiceController) runAction(action, service string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	var out []byte
	var err error
	switch action {
	case "start":
		out, err = s.exec.Run(ctx, "sc", "start", service)
	case "stop":
		out, err = s.exec.Run(ctx, "sc", "stop", service)
	case "restart":
		if _, stopErr := s.exec.Run(ctx, "sc", "stop", service); stopErr != nil {
			// Ignore stop failure and continue to start; service may already be stopped.
		}
		out, err = s.exec.Run(ctx, "sc", "start", service)
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}
	if err != nil {
		return fmt.Errorf("service %s failed: %w: %s", action, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (s *WindowsServiceController) Status(service string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	out, err := s.exec.Run(ctx, "sc", "query", service)
	if err != nil {
		return nil, fmt.Errorf("sc query failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	res := map[string]string{"raw": strings.TrimSpace(string(out))}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		upper := strings.ToUpper(line)
		if strings.HasPrefix(upper, "STATE") {
			res["STATE"] = line
			switch {
			case strings.Contains(upper, "RUNNING"):
				res["ActiveState"] = "active"
				res["SubState"] = "running"
			case strings.Contains(upper, "STOPPED"):
				res["ActiveState"] = "inactive"
				res["SubState"] = "dead"
			case strings.Contains(upper, "START_PENDING"):
				res["ActiveState"] = "activating"
				res["SubState"] = "start-pending"
			case strings.Contains(upper, "STOP_PENDING"):
				res["ActiveState"] = "deactivating"
				res["SubState"] = "stop-pending"
			}
		}
	}
	return res, nil
}
