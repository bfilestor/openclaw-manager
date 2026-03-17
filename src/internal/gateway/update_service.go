package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type UpdateService struct {
	exec    Executor
	timeout time.Duration
}

type VersionStatus struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	Channel         string `json:"channel"`
	UpdateAvailable bool   `json:"update_available"`
}

func NewUpdateService(exec Executor) *UpdateService {
	if exec == nil {
		exec = OSExecutor{}
	}
	return &UpdateService{exec: exec, timeout: 30 * time.Second}
}

func (s *UpdateService) VersionStatus() (*VersionStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	currentOut, err := s.exec.Run(ctx, "openclaw", "--version")
	if err != nil {
		return nil, fmt.Errorf("openclaw --version failed: %w", err)
	}
	current := strings.TrimSpace(string(currentOut))
	if current == "" {
		current = "unknown"
	}

	statusOut, err := s.exec.Run(ctx, "openclaw", "update", "status", "--json")
	if err != nil {
		return nil, fmt.Errorf("openclaw update status --json failed: %w", err)
	}
	statusOut = stripNonJSONPrefix(statusOut)

	var payload struct {
		Availability struct {
			Available     bool   `json:"available"`
			LatestVersion string `json:"latestVersion"`
		} `json:"availability"`
		Update struct {
			Registry struct {
				LatestVersion string `json:"latestVersion"`
			} `json:"registry"`
		} `json:"update"`
		Channel struct {
			Value string `json:"value"`
		} `json:"channel"`
	}
	if err := json.Unmarshal(statusOut, &payload); err != nil {
		return nil, fmt.Errorf("parse update status json failed: %w", err)
	}

	latest := strings.TrimSpace(payload.Availability.LatestVersion)
	if latest == "" {
		latest = strings.TrimSpace(payload.Update.Registry.LatestVersion)
	}
	if latest == "" {
		latest = current
	}

	channel := strings.TrimSpace(payload.Channel.Value)
	if channel == "" {
		channel = "stable"
	}

	return &VersionStatus{
		CurrentVersion:  current,
		LatestVersion:   latest,
		Channel:         channel,
		UpdateAvailable: payload.Availability.Available,
	}, nil
}

func (s *UpdateService) Upgrade() (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	out, err := s.exec.Run(ctx, "openclaw", "update", "--yes", "--json")
	if err != nil {
		return nil, fmt.Errorf("openclaw update failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return parseJSONOutput(out), nil
}

func (s *UpdateService) RollbackTo(version string) (map[string]any, error) {
	version = strings.TrimSpace(version)
	if version == "" {
		return nil, fmt.Errorf("rollback version required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	out, err := s.exec.Run(ctx, "openclaw", "update", "--tag", version, "--yes", "--json")
	if err != nil {
		return nil, fmt.Errorf("openclaw rollback failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return parseJSONOutput(out), nil
}

func parseJSONOutput(out []byte) map[string]any {
	result := map[string]any{}
	clean := stripNonJSONPrefix(out)
	if uErr := json.Unmarshal(clean, &result); uErr != nil {
		result["raw"] = strings.TrimSpace(string(out))
	}
	return result
}

func stripNonJSONPrefix(raw []byte) []byte {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return raw
	}
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return []byte(trimmed)
	}
	if idx := strings.IndexAny(trimmed, "{["); idx >= 0 {
		return []byte(strings.TrimSpace(trimmed[idx:]))
	}
	return []byte(trimmed)
}
