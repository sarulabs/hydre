package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hpcloud/tail"
)

// Service is the interface used to mock the Daemon structure.
type Service interface {
	Start(stopNotifier chan<- string)
	Stop()
	Kill()
}

// Daemon is a program executed in the background.
// Command is the UNIX command that starts the Daemon.
// The program will be stopped with a SIGTERM signal,
// but if StopCommand is defined, it will be used to stop the Daemon instead.
// The program running in the background is Command except if PidFile is defined.
// In this case the program is the one that has the pid specified in the PidFile.
// If LogFiles is not empty, the files will be used to stream
// the logs of the Daemon in addition to the Command output.
// You should probably use LogFiles if PidFile is defined.
type Daemon struct {
	Name        string
	Command     []string `yaml:"command"`
	StopCommand []string `yaml:"stopCommand"`
	PidFile     string   `yaml:"pidFile"`
	LogFiles    []string `yaml:"logFiles"`
	mu          sync.Mutex
	process     *os.Process
}

// Start executes the Daemon Command.
// It sets the Daemon process that is either the command process
// or the process defined by the PidFile.
// It also streams the Daemon logs on the standard output.
// When the daemon dies or when its logs are not  streamed anymore,
// it sends a message in the stopNotifier channel
// to report that the Daemon is not working properly anymore.
func (d *Daemon) Start(stopNotifier chan<- string) {
	log.Printf("starting `%s` daemon with command : %s\n", d.Name, d.Command)

	cmd := exec.Command(d.Command[0], d.Command[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	for _, logFile := range d.LogFiles {
		go d.printLogs(logFile, stopNotifier)
	}

	if d.PidFile == "" {
		d.startForeground(cmd, stopNotifier)
		return
	}

	d.startBackground(cmd, stopNotifier)
}

func (d *Daemon) startForeground(cmd *exec.Cmd, stopNotifier chan<- string) {
	if err := cmd.Start(); err != nil {
		stopNotifier <- fmt.Sprintf("could not start daemon `%s` : %s", d.Name, err)
		return
	}

	d.setProcess(cmd.Process)

	// send notification when the process dies
	go d.waitChild(stopNotifier)
}

func (d *Daemon) startBackground(cmd *exec.Cmd, stopNotifier chan<- string) {
	// remove pidfile content to be sure to read the correct value
	ioutil.WriteFile(d.PidFile, []byte{}, 0777)

	if err := cmd.Run(); err != nil {
		stopNotifier <- fmt.Sprintf("could not start daemon `%s` : %s", d.Name, err)
		return
	}

	pid := d.readPidFileRetry(500, 10*time.Millisecond)
	process, _ := os.FindProcess(pid)
	d.setProcess(process)

	// send notification when the process dies
	go d.wait(stopNotifier)
}

// waitChild waits for the Daemon process to finish and sends a message
// in the stopNotifier channel when that happens.
// The Daemon process need to be a child of this process.
func (d *Daemon) waitChild(stopNotifier chan<- string) {
	if p := d.getProcess(); p != nil {
		p.Wait()
	}

	stopNotifier <- fmt.Sprintf("daemon `%s` has stopped", d.Name)
}

// wait waits for the Daemon process to finish and sends a message
// in the stopNotifier channel when that happens.
// It checks every second that the process is still running.
func (d *Daemon) wait(stopNotifier chan<- string) {
	if p := d.getProcess(); p != nil {
		for isAlive(p.Pid) {
			time.Sleep(time.Second)
		}
	}

	stopNotifier <- fmt.Sprintf("daemon `%s` has stopped", d.Name)
}

// readPidFile returns the pid from the Daemon PidFile.
func (d *Daemon) readPidFile() (int, error) {
	b, err := ioutil.ReadFile(d.PidFile)
	if err != nil {
		return 0, err
	}

	pid := strings.Trim(string(b), " \n")

	return strconv.Atoi(pid)
}

// readPidFileRetry tries to read the pid from the Daemon PidFile.
// If it could not read it, it will retry maxAttempts times
// and will wait before each new attempt.
// If the pid file could not be red, it returns 0.
func (d *Daemon) readPidFileRetry(maxAttempts int, wait time.Duration) int {
	for i := 0; i < maxAttempts; i++ {
		if pid, _ := d.readPidFile(); pid != 0 {
			return pid
		}
	}

	return 0
}

// printLogs streams the content of a file on the standard output.
// When it stops working, it sends a message in the stopNotifier channel.
func (d *Daemon) printLogs(filename string, stopNotifier chan<- string) {
	t, _ := tail.TailFile(filename, tail.Config{Follow: true})

	for line := range t.Lines {
		fmt.Println(line.Text)
	}

	stopNotifier <- fmt.Sprintf("could not stream logs from `%s`", filename)
}

// Stop tries to gracefully stop the Daemon.
// If StopCommand is defined, it executes the command.
// If not it sends a SIGTERM signal to the Daemon process.
func (d *Daemon) Stop() {
	// there is a stop command to stop the process, execute it.
	if len(d.StopCommand) > 0 {
		cmd := exec.Command(d.StopCommand[0], d.StopCommand[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return
	}

	// there is no defined command to stop the process,
	// just send a SIGTERM signal.
	if p := d.getProcess(); len(d.StopCommand) == 0 && p != nil {
		p.Signal(syscall.SIGTERM)
		return
	}
}

// Kill sends a SIGKILL signal to the Daemon process.
func (d *Daemon) Kill() {
	if p := d.getProcess(); p != nil {
		p.Signal(syscall.SIGKILL)
	}
}

// getProcess is the thread safe process getter.
func (d *Daemon) getProcess() *os.Process {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.process
}

// getProcess is the thread safe process setter.
func (d *Daemon) setProcess(process *os.Process) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.process = process
}
