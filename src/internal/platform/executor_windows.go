//go:build windows

package platform

import (
	"context"
	"os/exec"
)

// WindowsExecutor runs commands in native Windows environments.
type WindowsExecutor struct{}

func NewDefaultExecutor() CommandExecutor {
	return WindowsExecutor{}
}

func (WindowsExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}
