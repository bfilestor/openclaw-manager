package platform

import "context"

// CommandExecutor abstracts command execution for platform-specific runtimes.
type CommandExecutor interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// ServiceController abstracts service lifecycle management.
type ServiceController interface {
	Start(service string) error
	Stop(service string) error
	Restart(service string) error
	Status(service string) (map[string]string, error)
}
