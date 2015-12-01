package main

import (
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
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
// if its state does not match the Enabled field.
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
// It returns the result of the daemon start command.
func (d *Daemon) DoStart() []byte {
	d.Enabled = true
	if !d.Quiet {
		log.Println("trying to start daemon", d.Name)
	}
	return d.Exec(d.Start)
}

// DoStop tries to stop the daemon.
// It returns the result of the daemon stop command.
func (d *Daemon) DoStop() []byte {
	d.Enabled = false
	if !d.Quiet {
		log.Println("trying to stop daemon", d.Name)
	}
	return d.Exec(d.Stop)
}

// Exec executes a command and returns its result.
func (d Daemon) Exec(cmdLine string) []byte {
	cmd := strings.Split(cmdLine, " ")
	res, _ := exec.Command(cmd[0], cmd[1:]...).Output()
	return res
}

// Pid returns the pid of the daemon based on its PidFile.
func (d Daemon) Pid() (int, error) {
	b, err := ioutil.ReadFile(d.PidFile)
	if err != nil {
		return 0, err
	}

	pid := strings.Trim(string(b), " \n")

	return strconv.Atoi(pid)
}

// IsAlive checks if the daemon is running by sending a kill signal.
func (d Daemon) IsAlive() bool {
	pid, err := d.Pid()
	if err != nil {
		return false
	}

	err = syscall.Kill(pid, 0)
	return err == nil
}
