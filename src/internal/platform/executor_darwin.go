//go:build darwin

package platform

import (
	"context"
	"os/exec"
)

// DarwinExecutor runs commands in native macOS environments.
type DarwinExecutor struct{}

func NewDefaultExecutor() CommandExecutor {
	return DarwinExecutor{}
}

func (DarwinExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}
