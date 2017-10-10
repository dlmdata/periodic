package periodic

import (
	"errors"
	"sync"
	"time"
)

// Execer is the interface that wraps the Exec method.
type Execer interface {
	Exec() error
}

// TaskFunc is an adapter to allow the use of a regular function as an Execer.
type TaskFunc func() error

func (f TaskFunc) Exec() error {
	return f()
}

// Task executes a function at regular intervals.
type Task struct {
	Error   <-chan error
	task    Execer
	stop    chan struct{}
	wait    time.Duration
	mutex   sync.RWMutex
	running bool
}

// NewTask returns a new Task that executes the given task with the given period.
func NewTask(period time.Duration, task Execer) *Task {
	if period <= 0 {
		panic(errors.New("period must be positive"))
	}
	errors := make(chan error, 1)
	t := &Task{
		Error:   errors,
		task:    task,
		stop:    make(chan struct{}),
		wait:    period,
		running: true,
	}
	go t.start(errors)
	return t
}

// Stop blocks until the task has ended and prevents further invocations.
func (t *Task) Stop() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.running {
		t.stop <- struct{}{}
		<-t.stop
	}
}

func (t *Task) start(errors chan error) {
	for {
		select {
		case msg := <-t.stop:
			close(errors)
			t.running = false
			t.stop <- msg
			return
		default:
			select {
			case errors <- t.task.Exec():
			default:
			}
			time.Sleep(t.wait)
		}
	}
}

// Background is a convenience function that executes the given task with the given period.
func Background(period time.Duration, task Execer) <-chan error {
	if period <= 0 {
		return nil
	}
	return NewTask(period, task).Error
}
