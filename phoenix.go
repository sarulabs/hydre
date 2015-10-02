package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

// Phoenix is the main structure.
// It holds the same information as the yaml configuration file.
type Phoenix struct {
	Interval int
	Port     string
	Daemons  map[string]*Daemon
	Quiet    bool
}

// Cmd holds information contained in the command line arguments.
// Worker : phonix run => worker ; phoenix start/stop => commander
// Action : (run | start | stop)
// Daemon : name of the daemon that need to be started/stopped
// Conf : configuration path
type Cmd struct {
	Worker bool
	Action string
	Daemon string
	Conf   string
}

// GetCmd parses command arguments.
func (p Phoenix) GetCmd(args []string) (Cmd, error) {
	if len(args) == 3 && args[1] == "run" {
		return Cmd{
			Worker: true,
			Conf:   args[2],
		}, nil
	}

	if len(args) == 4 && (args[1] == "start" || args[1] == "stop") {
		return Cmd{
			Worker: false,
			Action: args[1],
			Daemon: args[2],
			Conf:   args[3],
		}, nil
	}

	return Cmd{}, errors.New(p.Usage())
}

// Usage returns a string explaining how to use phoenix.
func (p Phoenix) Usage() string {
	return `~~~~~~~~~~~~~~~~
  Phoenix help
~~~~~~~~~~~~~~~~

> to start the worker process:
phoenix run conf_file.yml

> to communicate with the worker process:
phoenix (start|stop) daemon_name conf_file.yml
`
}

// Serve loads the configuration and runs phoenix.
// Depending on command arguments phoenix can run as a worker or a commander.
// The worker supervises processes and listens to orders (start/stop).
// A commander can send an order to the worker.
func (p *Phoenix) Serve(args []string) {
	cmd, err := p.GetCmd(args)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	p.LoadConf(cmd.Conf)

	if cmd.Worker {
		p.Supervise()
		p.Listen()

		// Reap child processes to avoid zombie processes.
		r := Reaper{}
		r.Start()

		var wg sync.WaitGroup
		wg.Add(1)
		wg.Wait()
		return
	}

	p.Send(cmd.Action, cmd.Daemon)
}

// LoadConf loads the Phoenix struct from a configuration file.
// The configuration file is in yaml format.
func (p *Phoenix) LoadConf(conf string) {
	var phoenix Phoenix

	c, err := ioutil.ReadFile(conf)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(c, &phoenix)
	if err != nil {
		log.Fatal(err)
	}

	for name, d := range phoenix.Daemons {
		d.Enabled = true
		d.Name = name
	}

	p.Interval = phoenix.Interval
	p.Port = phoenix.Port
	p.Daemons = phoenix.Daemons
}

// Send gives an order to the worker.
// action can be start or stop.
func (p Phoenix) Send(action, daemon string) {
	conn, err := net.Dial("tcp", ":"+p.Port)
	if err != nil {
		log.Fatal(err)
	}
	conn.Write([]byte(action + " " + daemon + "\n"))
	msg, _ := bufio.NewReader(conn).ReadString('\n')
	if !p.Quiet {
		fmt.Println(strings.Trim(msg, " \n"))
	}
}

// Listen creates a goroutine waiting for orders
// that can be send via the Send function.
func (p *Phoenix) Listen() {
	go func() {
		ln, err := net.Listen("tcp", ":"+p.Port)
		if err != nil {
			log.Fatal(err)
		}
		if !p.Quiet {
			log.Println("listening to 127.0.0.1:", p.Port)
		}
		for {
			conn, err := ln.Accept()
			if err != nil && !p.Quiet {
				log.Println(err)
			}
			go p.handleConn(conn)
		}
	}()
}

// handleConn reads orders coming to a connection.
// Then it updates daemons to fulfill the order.
func (p Phoenix) handleConn(conn net.Conn) {
	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			return
		}
		msg = strings.Trim(msg, " \n")
		cmd := strings.Split(msg, " ")
		if len(cmd) == 2 {
			daemon := string(cmd[1])
			d, ok := p.Daemons[daemon]
			if !ok {
				conn.Write([]byte("could not find daemon `" + daemon + "`\n"))
				return
			}

			switch string(cmd[0]) {
			case "start":
				ret := d.DoStart()
				conn.Write(ret)
				conn.Write([]byte{'\n'})
				return
			case "stop":
				ret := d.DoStop()
				conn.Write(ret)
				conn.Write([]byte{'\n'})
				return
			default:
				conn.Write([]byte("wrong command `" + msg + "`\n"))
				return
			}
		}
	}
}

// Supervise ensures that daemons are in the correct state.
// It checks daemon state every p.Interval seconds.
func (p *Phoenix) Supervise() {
	for _, d := range p.Daemons {
		d.EnsureState()
	}
	ticker := time.NewTicker(time.Duration(p.Interval) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, d := range p.Daemons {
					d.EnsureState()
				}
			}
		}
	}()
}
