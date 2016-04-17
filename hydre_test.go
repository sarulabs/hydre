package main

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mock Daemon is used to test Hydre Start and Stop methods.
type mockDaemon struct {
	mu           sync.Mutex
	Started      bool
	Stopped      bool
	Killed       bool
	StopNotifier chan<- string
}

func (m *mockDaemon) Start(stopNotifier chan<- string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.StopNotifier = stopNotifier
	m.Started = true
}

func (m *mockDaemon) Stop() {
	time.Sleep(500 * time.Millisecond)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Stopped = true
}

func (m *mockDaemon) Kill() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Killed = true
}

func (m *mockDaemon) IsStarted() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Started
}

func (m *mockDaemon) IsStopped() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Stopped
}

func (m *mockDaemon) IsKilled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Killed
}

func TestNewHydre(t *testing.T) {
	h, err := NewHydre("test/good-conf.yml")
	assert.Nil(t, err)
	assert.Equal(t, &Hydre{
		Timeout: 10,
		Daemons: []Service{
			&Daemon{
				Name:        "daemon1",
				Command:     []string{"start", "daemon1"},
				StopCommand: []string{"stop", "daemon1"},
				PidFile:     "path/to/pidfile",
				LogFiles:    []string{"1.log", "2.log"},
			},
			&Daemon{
				Name:    "daemon2",
				Command: []string{"start", "daemon2"},
			},
		},
	}, h)

	h, err = NewHydre("test/bad-conf.yml")
	assert.NotNil(t, err)
	assert.Nil(t, h)

	h, err = NewHydre("does-not-exist.yml")
	assert.NotNil(t, err)
	assert.Nil(t, h)
}

func TestRunAndStopGracefully(t *testing.T) {
	d1 := &mockDaemon{}
	d2 := &mockDaemon{}

	h := &Hydre{
		Timeout: 1,
		Daemons: []Service{d1, d2},
	}

	go h.Run()

	time.Sleep(100 * time.Millisecond)

	assert.True(t, d1.IsStarted())
	assert.False(t, d1.IsStopped())
	assert.False(t, d1.IsKilled())
	assert.True(t, d2.IsStarted())
	assert.False(t, d2.IsStopped())
	assert.False(t, d2.IsKilled())

	require.NotNil(t, d1.StopNotifier)
	d1.StopNotifier <- "stop daemon 1"

	time.Sleep(700 * time.Millisecond)

	assert.True(t, d1.IsStarted())
	assert.True(t, d1.IsStopped())
	assert.True(t, d1.IsKilled())
	assert.True(t, d2.IsStarted())
	assert.True(t, d2.IsStopped())
	assert.True(t, d2.IsKilled())
}

func TestStopForcefully(t *testing.T) {
	d1 := &mockDaemon{}
	d2 := &mockDaemon{}

	h := &Hydre{
		Timeout: 0,
		Daemons: []Service{d1, d2},
	}

	h.Stop()

	assert.False(t, d1.IsStopped())
	assert.True(t, d1.IsKilled())
	assert.False(t, d2.IsStopped())
	assert.True(t, d2.IsKilled())
}
