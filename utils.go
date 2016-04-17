package main

import (
	"os"
	"syscall"
)

// isAlive uses kill 0 to check if the process is alive
func isAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return pid != 0 && err == nil
}

func fileExist(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
