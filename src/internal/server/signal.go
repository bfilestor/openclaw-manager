package server

import (
	"os"
	"os/signal"
	"syscall"
)

func signalNotify(ch chan<- os.Signal) {
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
}
