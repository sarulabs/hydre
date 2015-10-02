package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

// Daemon holds the daemon configuration and its desired state.
type Daemon struct {
	Name    string
	PidFile string
	Start   string
	Stop    string
	Enabled bool
	Quiet   bool
}

// EnsureState tries to start or stop a daemon
// if its state does not match to the d.Enabled value.
func (d Daemon) EnsureState() {
	if d.Enabled != d.IsAlive() {
		if d.Enabled {
			d.DoStart()
		} else {
			d.DoStop()
		}
	}
}

// DoStart tries to start the daemon.
func (d *Daemon) DoStart() []byte {
	d.Enabled = true
	if !d.Quiet {
		log.Println("trying to start daemon", d.Name)
	}
	return d.Exec(d.Start)
}

// DoStop tries to stop the daemon.
func (d *Daemon) DoStop() []byte {
	d.Enabled = false
	if !d.Quiet {
		log.Println("trying to stop daemon", d.Name)
	}
	return d.Exec(d.Stop)
}

// Exec executes a command.
func (d Daemon) Exec(cmdLine string) []byte {
	cmd := strings.Split(cmdLine, " ")
	res, _ := exec.Command(cmd[0], cmd[1:]...).Output()
	return res
}

// Pid returns the pid of the daemon based on its PidFile.
// If it can not find the pid of the daemon, it returns -1.
func (d Daemon) Pid() string {
	b, err := ioutil.ReadFile(d.PidFile)
	if err != nil {
		return "-1"
	}
	pid := strings.Trim(string(b), " \n")
	if pid == "" {
		return "-1"
	}
	return pid
}

// IsAlive checks if the daemon is running by checking the /proc/ directory.
func (d Daemon) IsAlive() bool {
	_, err := os.Stat("/proc/" + d.Pid())
	return err == nil
}
