//go:build linux

package platform

import (
	"context"
	"os/exec"
)

// LinuxExecutor runs commands in native Linux environments.
type LinuxExecutor struct{}

func NewDefaultExecutor() CommandExecutor {
	return LinuxExecutor{}
}

func (LinuxExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}
