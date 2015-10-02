package main

import (
	"os"
	"os/signal"
	"syscall"
)

// Reaper waits for its children to finish to avoid zombie processes.
type Reaper struct{}

// Start launches the Reaper in the background.
// It only works if the current process has a pid equal to 1.
func (r Reaper) Start() {
	if 1 == os.Getpid() {
		go r.reap()
	}
}

func (r Reaper) reap() {
	c := make(chan os.Signal, 10)

	signal.Notify(c, syscall.SIGCHLD)

	for {
		var err error
		_ = <-c
		for syscall.ECHILD != err {
			_, err = syscall.Wait4(-1, nil, 0, nil)
		}
	}
}
