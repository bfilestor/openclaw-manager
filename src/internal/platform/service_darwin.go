//go:build darwin

package platform

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type DarwinServiceController struct {
	exec    CommandExecutor
	timeout time.Duration
}

func NewDefaultServiceController(exec CommandExecutor) ServiceController {
	if exec == nil {
		exec = NewDefaultExecutor()
	}
	return &DarwinServiceController{exec: exec, timeout: 30 * time.Second}
}

func (s *DarwinServiceController) Start(service string) error {
	label := normalizeLaunchdLabel(service)
	target := launchdTarget(label)
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	out, err := s.exec.Run(ctx, "launchctl", "kickstart", "-k", target)
	if err != nil {
		return fmt.Errorf("launchctl kickstart failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (s *DarwinServiceController) Stop(service string) error {
	label := normalizeLaunchdLabel(service)
	target := launchdTarget(label)
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	out, err := s.exec.Run(ctx, "launchctl", "bootout", target)
	if err != nil {
		return fmt.Errorf("launchctl bootout failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (s *DarwinServiceController) Restart(service string) error {
	_ = s.Stop(service)
	return s.Start(service)
}

func (s *DarwinServiceController) Status(service string) (map[string]string, error) {
	label := normalizeLaunchdLabel(service)
	target := launchdTarget(label)
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	out, err := s.exec.Run(ctx, "launchctl", "print", target)
	text := strings.TrimSpace(string(out))
	if err != nil {
		if strings.Contains(strings.ToLower(text), "could not find service") || strings.Contains(strings.ToLower(text), "service not found") {
			return map[string]string{"ActiveState": "inactive", "SubState": "dead", "raw": text}, nil
		}
		return nil, fmt.Errorf("launchctl print failed: %w: %s", err, text)
	}

	state := "active"
	sub := "running"
	lower := strings.ToLower(text)
	if strings.Contains(lower, "state = waiting") {
		sub = "waiting"
	}
	if strings.Contains(lower, "last exit code =") {
		// launchd service exists but may not be running continuously; keep active state for visibility.
	}
	return map[string]string{"ActiveState": state, "SubState": sub, "raw": text}, nil
}

func normalizeLaunchdLabel(service string) string {
	name := strings.TrimSpace(service)
	name = strings.TrimSuffix(name, ".service")
	return name
}

func launchdTarget(label string) string {
	uid := os.Getuid()
	return "gui/" + strconv.Itoa(uid) + "/" + label
}
