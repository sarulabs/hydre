package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"
)

// Hydre holds the application configuration, including the Daemons.
// It can start and stop the Daemons.
// After starting the Daemons, it will stop all of them as soon as one stops working.
type Hydre struct {
	Daemons []Service
	Timeout int
}

// hydreConf is used to load the configuration from a yaml file.
type hydreConf struct {
	Daemons map[string]*Daemon `yaml:"daemons"`
	Timeout int                `yaml:"timeout"`
}

// NewHydre loads a configuration file in an Hydre structure.
func NewHydre(file string) (*Hydre, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	conf := &hydreConf{}

	err = yaml.Unmarshal(content, conf)
	if err != nil {
		return nil, err
	}

	h := &Hydre{}
	h.Timeout = conf.Timeout
	h.Daemons = make([]Service, 0, len(conf.Daemons))

	for name, d := range conf.Daemons {
		d.Name = name
		h.Daemons = append(h.Daemons, d)
	}

	return h, nil
}

// Run starts all the Daemons and calls the Stop method
// when one of the Daemons stops working.
func (h *Hydre) Run() {
	// avoid zombie processes
	go h.reap()

	stopNotifier := make(chan string, 1000)

	for _, d := range h.Daemons {
		d.Start(stopNotifier)
	}

	log.Println(<-stopNotifier)

	h.Stop()
}

// reap waits for this process children to finish to avoid zombie processes.
func (h *Hydre) reap() {
	c := make(chan os.Signal, 1000)

	signal.Notify(c, syscall.SIGCHLD)

	for {
		var err error
		_ = <-c
		for syscall.ECHILD != err {
			_, err = syscall.Wait4(-1, nil, 0, nil)
		}
	}
}

// Stop stops all the Daemons.
// It first tries to stop the Daemon properly.
// But one Daemon could not be stopped after Timeout seconds,
// it will be stopped with a SIGKILL signal.
func (h *Hydre) Stop() {
	stopped := make(chan struct{})

	// set a timeout, letting time for the Daemons to stop properly
	go func() {
		time.Sleep(time.Duration(h.Timeout) * time.Second)
		log.Printf("could not stop all daemons properly in %d seconds\n", h.Timeout)
		stopped <- struct{}{}
	}()

	// stop all Daemons
	go func() {
		var wg sync.WaitGroup

		for _, d := range h.Daemons {
			wg.Add(1)
			go func(d Service) {
				defer wg.Done()
				d.Stop()
			}(d)
		}

		wg.Wait()

		stopped <- struct{}{}
	}()

	<-stopped

	// kill Daemons that are still alive
	for _, d := range h.Daemons {
		d.Kill()
	}
}
