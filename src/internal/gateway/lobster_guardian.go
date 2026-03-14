package gateway

import (
	"context"
	"errors"
	"log"
	"path/filepath"
	"time"

	"openclaw-manager/internal/auth"
	"openclaw-manager/internal/config"
	"openclaw-manager/internal/storage"
)

const (
	lobsterGuardianTickInterval = 10 * time.Second
	lobsterGuardianRetryWait    = 10 * time.Second
	lobsterGuardianMaxAttempts  = 5
)

type LobsterGuardian struct {
	Settings       *auth.SystemSettingsRepository
	Revisions      *config.RevisionRepository
	Service        *SystemctlService
	ServiceName    string
	OpenClawJSON   string
	MaxAttempts    int
	TickInterval   time.Duration
	RetryWait      time.Duration
}

func (g *LobsterGuardian) Run(ctx context.Context) {
	if g == nil || g.Settings == nil || g.Revisions == nil || g.Service == nil {
		return
	}
	if g.ServiceName == "" {
		g.ServiceName = "openclaw-gateway.service"
	}
	if g.MaxAttempts <= 0 {
		g.MaxAttempts = lobsterGuardianMaxAttempts
	}
	if g.TickInterval <= 0 {
		g.TickInterval = lobsterGuardianTickInterval
	}
	if g.RetryWait <= 0 {
		g.RetryWait = lobsterGuardianRetryWait
	}
	if g.OpenClawJSON == "" {
		log.Printf("lobster-guardian: missing openclaw.json path, guardian disabled")
		return
	}

	ticker := time.NewTicker(g.TickInterval)
	defer ticker.Stop()

	for {
		if ctx.Err() != nil {
			return
		}
		g.tick(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (g *LobsterGuardian) tick(ctx context.Context) {
	enabled, err := g.Settings.IsLobsterGuardianEnabled()
	if err != nil {
		log.Printf("lobster-guardian: read setting failed: %v", err)
		return
	}
	if !enabled {
		return
	}

	st, err := g.Service.Status(g.ServiceName)
	if err != nil {
		log.Printf("lobster-guardian: status check failed: %v", err)
		return
	}
	if st == nil || st.ActiveState != "inactive" {
		return
	}

	log.Printf("lobster-guardian: gateway inactive, start auto recovery")
	if err := g.recoverGateway(ctx); err != nil {
		log.Printf("lobster-guardian: recovery failed: %v", err)
	}
}

func (g *LobsterGuardian) recoverGateway(ctx context.Context) error {
	revs, err := g.Revisions.List("openclaw_json", "", g.MaxAttempts)
	if err != nil {
		return err
	}
	if len(revs) == 0 {
		return errors.New("no openclaw_json revisions found")
	}

	attempts := g.MaxAttempts
	if len(revs) < attempts {
		attempts = len(revs)
	}

	for i := 0; i < attempts; i++ {
		rev := revs[i]
		if writeErr := storage.AtomicWriteFile(g.OpenClawJSON, []byte(rev.Content), 0o644); writeErr != nil {
			log.Printf("lobster-guardian: attempt %d write revision %s failed: %v", i+1, rev.RevisionID, writeErr)
			continue
		}
		if restartErr := g.Service.Restart(g.ServiceName); restartErr != nil {
			log.Printf("lobster-guardian: attempt %d restart failed: %v", i+1, restartErr)
			continue
		}

		if !sleepWithContext(ctx, g.RetryWait) {
			return ctx.Err()
		}

		st, statusErr := g.Service.Status(g.ServiceName)
		if statusErr != nil {
			log.Printf("lobster-guardian: attempt %d status check failed: %v", i+1, statusErr)
			continue
		}
		if st != nil && st.ActiveState == "active" {
			log.Printf("lobster-guardian: recovered with revision %s (attempt %d)", rev.RevisionID, i+1)
			return nil
		}
		state := "unknown"
		if st != nil {
			state = st.ActiveState
		}
		log.Printf("lobster-guardian: attempt %d still failed, state=%s", i+1, state)
	}

	return errors.New("guardian failed after max attempts")
}

func sleepWithContext(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func DefaultOpenClawJSONPath(openclawHome string) string {
	if openclawHome == "" {
		return ""
	}
	return filepath.Join(openclawHome, "openclaw.json")
}
