package task

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrTaskConflict = errors.New("task conflict")
)

type HandlerFunc func(ctx context.Context, task *Task) (stdout, stderr string, exitCode int, err error)

type Engine struct {
	repo       *Repository
	maxWorkers int
	handlers   map[string]HandlerFunc
	timeouts   map[string]time.Duration

	queue   chan string
	stopCh  chan struct{}
	wg      sync.WaitGroup

	mu             sync.Mutex
	gatewayRunning bool
}

func NewEngine(repo *Repository, maxWorkers int) *Engine {
	if maxWorkers <= 0 {
		maxWorkers = 3
	}
	return &Engine{
		repo:       repo,
		maxWorkers: maxWorkers,
		handlers:   map[string]HandlerFunc{},
		timeouts:   map[string]time.Duration{},
		queue:      make(chan string, 256),
		stopCh:     make(chan struct{}),
	}
}

func (e *Engine) Register(taskType string, timeout time.Duration, h HandlerFunc) {
	e.handlers[taskType] = h
	if timeout > 0 {
		e.timeouts[taskType] = timeout
	}
}

func (e *Engine) Start() {
	for i := 0; i < e.maxWorkers; i++ {
		e.wg.Add(1)
		go e.worker()
	}
}

func (e *Engine) Stop() {
	close(e.stopCh)
	e.wg.Wait()
}

func (e *Engine) Enqueue(taskID string) {
	e.queue <- taskID
}

func (e *Engine) Cancel(taskID string) error {
	t, err := e.repo.FindByID(taskID)
	if err != nil {
		return err
	}
	if t.Status != StatusPending {
		return errors.New("only pending task can be canceled")
	}
	return e.repo.UpdateStatus(taskID, StatusCanceled)
}

func (e *Engine) worker() {
	defer e.wg.Done()
	for {
		select {
		case <-e.stopCh:
			return
		case id := <-e.queue:
			e.runOne(id)
		}
	}
}

func (e *Engine) runOne(taskID string) {
	t, err := e.repo.FindByID(taskID)
	if err != nil {
		return
	}
	if t.Status != StatusPending {
		return
	}

	isGateway := isGatewayTask(t.TaskType)
	if isGateway {
		e.mu.Lock()
		if e.gatewayRunning {
			e.mu.Unlock()
			_ = e.repo.UpdateResult(taskID, intPtr(-1), "", ErrTaskConflict.Error(), "")
			_ = e.repo.UpdateStatus(taskID, StatusFailed)
			return
		}
		e.gatewayRunning = true
		e.mu.Unlock()
		defer func() {
			e.mu.Lock()
			e.gatewayRunning = false
			e.mu.Unlock()
		}()
	}

	h, ok := e.handlers[t.TaskType]
	if !ok {
		_ = e.repo.UpdateResult(taskID, intPtr(-1), "", "handler not found", "")
		_ = e.repo.UpdateStatus(taskID, StatusFailed)
		return
	}
	_ = e.repo.UpdateStatus(taskID, StatusRunning)

	timeout := e.timeouts[t.TaskType]
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	stdout, stderr, code, runErr := h(ctx, t)
	if ctx.Err() == context.DeadlineExceeded {
		_ = e.repo.UpdateResult(taskID, intPtr(-1), stdout, stderr, "")
		_ = e.repo.UpdateStatus(taskID, StatusFailed)
		return
	}
	_ = e.repo.UpdateResult(taskID, intPtr(code), stdout, stderr, "")
	if runErr != nil || code != 0 {
		_ = e.repo.UpdateStatus(taskID, StatusFailed)
	} else {
		_ = e.repo.UpdateStatus(taskID, StatusSucceeded)
	}
}

func isGatewayTask(taskType string) bool {
	return taskType == "gateway.start" || taskType == "gateway.stop" || taskType == "gateway.restart"
}

func intPtr(v int) *int { return &v }
