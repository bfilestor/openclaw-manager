//go:build linux

package platform

import (
	"context"

	"openclaw-manager/internal/gateway"
)

// LinuxExecutor reuses the existing gateway OS executor on Linux.
type LinuxExecutor struct {
	delegate gateway.OSExecutor
}

func NewDefaultExecutor() CommandExecutor {
	return LinuxExecutor{delegate: gateway.OSExecutor{}}
}

func (e LinuxExecutor) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return e.delegate.Run(ctx, name, args...)
}
