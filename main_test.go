package main

import (
	"os"
	"os/exec"
	"testing"
	"time"
)

func GetTestDir() string {
	dir, _ := os.Getwd()
	return dir + "/test/"
}

func GetStartCommand(id string) string {
	return "start-stop-daemon --start --background --quiet --pidfile " + GetTestDir() + id + ".pid --make-pidfile --exec /bin/sh " + GetTestDir() + "proc.sh proc-" + id
}

func GetStopCommand(id string) string {
	return "start-stop-daemon --stop --quiet --pidfile " + GetTestDir() + id + ".pid --exec /bin/sh " + GetTestDir() + "proc.sh proc-" + id
}

func Sleep(seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
}

func CloneDaemon(d Daemon) *Daemon {
	return &d
}

func TestAll(t *testing.T) {
	quiet := true

	master := Phoenix{
		Interval: 10,
		Port:     "9988",
		Daemons: map[string]*Daemon{
			"1": &Daemon{
				Name:    "1",
				PidFile: GetTestDir() + "1.pid",
				Start:   GetStartCommand("1"),
				Stop:    GetStopCommand("1"),
				Enabled: true,
				Quiet:   quiet,
			},
			"2": &Daemon{
				Name:    "2",
				PidFile: GetTestDir() + "2.pid",
				Start:   GetStartCommand("2"),
				Stop:    GetStopCommand("2"),
				Enabled: false,
				Quiet:   quiet,
			},
		},
		Quiet: quiet,
	}

	slave := Phoenix{
		Port:  "9988",
		Quiet: quiet,
	}

	// start daemons to create pid files
	// then stop daemons to set initial state
	// use clones to avoid Enabled to be changed on daemons
	CloneDaemon(*master.Daemons["1"]).DoStart()
	CloneDaemon(*master.Daemons["2"]).DoStart()

	Sleep(1)

	CloneDaemon(*master.Daemons["1"]).DoStop()
	CloneDaemon(*master.Daemons["2"]).DoStop()

	Sleep(1)

	// test initial status
	// prev: (1) -- (2) --
	if master.Daemons["1"].IsAlive() {
		t.Error("daemon 1 was already started")
	}
	if master.Daemons["2"].IsAlive() {
		t.Error("daemon 2 was already started")
	}

	// supervise should start daemon #1
	// prev: (1) -- (2) --
	// next: (1) OK (2) --
	master.Supervise()
	master.Listen()

	Sleep(1)

	if !master.Daemons["1"].IsAlive() {
		t.Error("daemon 1 failed to start")
	}
	if master.Daemons["2"].IsAlive() {
		t.Error("daemon 2 started and should not have")
	}

	// try to stop daemon #1 with the slave phoenix
	// prev: (1) OK (2) --
	// next: (1) -- (2) --
	slave.Send("stop", "1")

	Sleep(1)

	if master.Daemons["1"].IsAlive() {
		t.Error("daemon 1 could not stop")
	}
	if master.Daemons["2"].IsAlive() {
		t.Error("daemon 2 started and should not have")
	}

	// try to start daemon #2 with the slave phoenix
	// prev: (1) -- (2) --
	// next: (1) -- (2) OK
	slave.Send("start", "2")

	Sleep(1)

	if master.Daemons["1"].IsAlive() {
		t.Error("daemon 1 should not be alive")
	}
	if !master.Daemons["2"].IsAlive() {
		t.Error("daemon 2 could not start")
	}

	// try to kill daemon #2
	// prev: (1) -- (2) OK
	// next: (1) -- (2) --
	pid := master.Daemons["2"].Pid()
	exec.Command("kill", "-9", pid).Output()

	Sleep(1)

	if master.Daemons["1"].IsAlive() {
		t.Error("daemon 1 should not be alive")
	}
	if master.Daemons["2"].IsAlive() {
		t.Log("daemon 2 should have been killed, maybe it was resurrected really fast")
	}

	// hope that daemon #2 will restart after phoenix supervise interval
	// prev: (1) -- (2) --
	// next: (1) -- (2) OK
	Sleep(11)

	if master.Daemons["1"].IsAlive() {
		t.Error("daemon 1 should not be alive")
	}
	if !master.Daemons["2"].IsAlive() {
		t.Error("daemon 2 should have been resurrected")
	}

	// stop daemons
	master.Daemons["1"].DoStop()
	master.Daemons["2"].DoStop()
}
