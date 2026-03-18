//go:build linux

package platform

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type LinuxServiceController struct {
	exec    CommandExecutor
	timeout time.Duration
}

func NewDefaultServiceController(exec CommandExecutor) ServiceController {
	if exec == nil {
		exec = NewDefaultExecutor()
	}
	return &LinuxServiceController{exec: exec, timeout: 30 * time.Second}
}

func (s *LinuxServiceController) Start(service string) error { return s.runAction("start", service) }
func (s *LinuxServiceController) Stop(service string) error  { return s.runAction("stop", service) }
func (s *LinuxServiceController) Restart(service string) error {
	return s.runAction("restart", service)
}

func (s *LinuxServiceController) runAction(action, service string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	out, err := s.exec.Run(ctx, "systemctl", "--user", action, service)
	if err != nil {
		return fmt.Errorf("systemctl %s failed: %w: %s", action, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (s *LinuxServiceController) Status(service string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	out, err := s.exec.Run(ctx, "systemctl", "--user", "show", service, "--no-page", "--property=ActiveState,SubState,MainPID")
	if err != nil {
		return nil, fmt.Errorf("systemctl show failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	res := map[string]string{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			res[parts[0]] = parts[1]
		}
	}
	return res, nil
}
