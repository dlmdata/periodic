package periodic

import (
	"errors"
	"sync"
	"time"
)

type Execer interface {
	Exec() error
}

type TaskFunc func() error

func (f TaskFunc) Exec() error {
	return f()
}

type Task struct {
	Error   <-chan error
	task    Execer
	stop    chan struct{}
	wait    time.Duration
	mutex   sync.RWMutex
	running bool
}

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

func (t *Task) Running() bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.running
}

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

func Background(period time.Duration, task Execer) <-chan error {
	if period <= 0 {
		return nil
	}
	return NewTask(period, task).Error
}
