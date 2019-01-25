package siger

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var done = false
var doneMu sync.Mutex

// RegisterHandler registers the handler for the SIGTERM signal.
func RegisterHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		doneMu.Lock()
		done = true
		doneMu.Unlock()
	}()
}

// Done returns true when a SIGTERM signal has been received.
func Done() bool {
	doneMu.Lock()
	defer doneMu.Unlock()

	return done
}
