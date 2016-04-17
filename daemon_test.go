package main

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testDir() string {
	dir, _ := os.Getwd()
	return dir + "/test/"
}

func TestDaemonStart(t *testing.T) {
	d := &Daemon{
		Command: []string{"sleep", "60"},
	}

	require.Nil(t, d.getProcess())

	stopNotifier := make(chan string, 10)

	d.Start(stopNotifier)

	p := d.getProcess()
	require.NotNil(t, p)
	require.True(t, isAlive(p.Pid))
	require.Equal(t, 0, len(stopNotifier))

	syscall.Kill(p.Pid, 9)
	time.Sleep(1000 * time.Millisecond)

	require.True(t, !isAlive(p.Pid))
	require.Equal(t, 1, len(stopNotifier))
}

func TestDaemonStartWithPidFile(t *testing.T) {
	os.Remove(testDir() + "write.pid")

	d := &Daemon{
		Command: []string{"/bin/sh", testDir() + "write.sh", testDir() + "write.pid", "5555"},
		PidFile: testDir() + "write.pid",
	}

	pidFileValue, _ := d.readPidFile()
	require.Equal(t, 0, pidFileValue)

	stopNotifier := make(chan string)

	d.Start(stopNotifier)

	pidFileValue, _ = d.readPidFile()
	require.Equal(t, 5555, pidFileValue)
}

func TestDaemonStopCommand(t *testing.T) {
	stopCommandFile := testDir() + "stopcommand.tmp"

	d := &Daemon{
		StopCommand: []string{"touch", stopCommandFile},
	}

	os.Remove(stopCommandFile)

	w, err := os.Create(stopCommandFile)
	require.Nil(t, err)
	defer w.Close()
	defer os.Remove(stopCommandFile)

	d.Stop()

	require.True(t, fileExist(stopCommandFile))
}

func TestDaemonKill(t *testing.T) {
	d := &Daemon{
		Command: []string{"sleep", "60"},
	}

	stopNotifier := make(chan string, 10)

	// start daemon
	d.Start(stopNotifier)

	time.Sleep(100 * time.Millisecond)

	p := d.getProcess()
	require.NotNil(t, p)
	require.True(t, isAlive(p.Pid))

	// kill daemon
	d.Kill()

	time.Sleep(100 * time.Millisecond)

	require.True(t, !isAlive(p.Pid))
	require.Equal(t, 1, len(stopNotifier))
}
