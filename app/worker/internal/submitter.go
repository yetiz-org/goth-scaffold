package internal

import (
	"errors"
	"sync"

	"github.com/yetiz-org/asynq"
)

type SubmitFunc func(task *asynq.Task, opts ...asynq.Option) error

var (
	submitMu  sync.RWMutex
	submitter SubmitFunc
)

func SetSubmitter(s SubmitFunc) {
	submitMu.Lock()
	submitter = s
	submitMu.Unlock()
}

// Submit enqueues a task from within a handler's Run method (self-dispatch).
// The submitter is injected by worker.StartClient; returns error if not configured.
func Submit(task *asynq.Task, opts ...asynq.Option) error {
	submitMu.RLock()
	s := submitter
	submitMu.RUnlock()
	if s == nil {
		return errors.New("submitter is not configured")
	}

	return s(task, opts...)
}
