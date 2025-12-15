package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/yetiz-org/asynq"
	"github.com/yetiz-org/gone/erresponse"
	"github.com/yetiz-org/gone/ghttp"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/connector/redis"
	"github.com/yetiz-org/goth-scaffold/app/worker/internal"
)

var (
	serveMux         = asynq.NewServeMux()
	srvList          []*asynq.Server
	singleSrvList    []*_SingleInstanceManager
	schedulerManager *_SingleInstanceManager
	client           *asynq.Client
	inspector        *asynq.Inspector
	scheduler        *asynq.Scheduler
)

// _ServeHandler wraps internal.Handler as asynq.Handler
type _ServeHandler struct {
	handler internal.Handler
}

func (t *_ServeHandler) _WrapError(err error) ghttp.ErrorResponse {
	var cast ghttp.ErrorResponse
	if errors.As(err, &cast) {
		return cast
	}

	return erresponse.ServerErrorCrossServiceOperationWithFormat("%v", err)
}

func (t *_ServeHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	executionContext := make(map[string]any)
	taskInfo := internal.NewTaskInfo(task.Type(), nil, task.ResultWriter())
	if err := json.Unmarshal(task.Payload(), &taskInfo.Payload); err != nil {
		kklogger.WarnJ("worker:Service.ProcessTask#unmarshal!failed", err.Error())
		taskInfo.Payload = nil
	}

	if err := t.handler.Before(ctx, taskInfo, executionContext); err != nil {
		kklogger.ErrorJ("worker:Service.ProcessTask#before!failed", err.Error())
		return t.handler.ErrorCaught(ctx, taskInfo, executionContext, t._WrapError(err))
	}

	if err := t.handler.Run(ctx, taskInfo, executionContext); err != nil {
		kklogger.ErrorJ("worker:Service.ProcessTask#run!failed", err.Error())
		return t.handler.ErrorCaught(ctx, taskInfo, executionContext, t._WrapError(err))
	}

	if err := t.handler.After(ctx, taskInfo, executionContext); err != nil {
		kklogger.ErrorJ("worker:Service.ProcessTask#after!failed", err.Error())
		return t.handler.ErrorCaught(ctx, taskInfo, executionContext, t._WrapError(err))
	}

	return nil
}

// Register registers a task handler
func Register(pattern string, handler internal.Handler) {
	if pattern == "" {
		pattern = handler.Name()
	}

	serveMux.Handle(pattern, &_ServeHandler{handler: handler})
}

// RegisterSchedule registers a single cron task
func RegisterSchedule(scheduler *asynq.Scheduler, cronSpec string, task *asynq.Task) string {
	entryID, _ := scheduler.Register(cronSpec, task)
	return entryID
}

// Submit enqueues a task
func Submit(task *asynq.Task, opts ...asynq.Option) (*TaskFuture, error) {
	if taskInfo, err := client.Enqueue(task, opts...); err != nil {
		return nil, err
	} else {
		return NewTaskFuture(Inspector(), taskInfo), nil
	}
}

// SubmitUnique enqueues a unique task
func SubmitUnique(task *asynq.Task, uniqueTTL time.Duration, opts ...asynq.Option) (*TaskFuture, error) {
	if taskInfo, err := client.Enqueue(task, append(opts, asynq.Unique(uniqueTTL))...); err != nil {
		return nil, err
	} else {
		return NewTaskFuture(Inspector(), taskInfo), nil
	}
}

// Execute runs a task directly (without enqueue)
func Execute(ctx context.Context, task *asynq.Task) error {
	return serveMux.ProcessTask(ctx, task)
}

// Inspector returns task inspector
func Inspector() *asynq.Inspector {
	return inspector
}

// Scheduler returns task scheduler
func Scheduler() *asynq.Scheduler {
	return scheduler
}

// StartClient starts Client and Inspector
func StartClient(namespace string) {
	// Check if already started
	if client != nil || inspector != nil {
		kklogger.WarnJ("worker:StartClient#already!started", "client already started")
		return
	}

	redisAddr := fmt.Sprintf("%s:%d", redis.Instance().Master().Meta().Host, redis.Instance().Master().Meta().Port)
	client = asynq.NewClientWithNamespace(asynq.RedisClientOpt{Addr: redisAddr}, namespace)
	inspector = asynq.NewInspectorWithNamespace(asynq.RedisClientOpt{Addr: redisAddr}, namespace)
	kklogger.InfoJ("worker:StartClient#started", fmt.Sprintf("namespace=%s", namespace))
}

// StopClient stops Client and Inspector
func StopClient() {
	if client != nil {
		_ = client.Close()
		client = nil
		kklogger.InfoJ("worker:StopClient#client!closed", "client closed")
	}

	if inspector != nil {
		_ = inspector.Close()
		inspector = nil
		kklogger.InfoJ("worker:StopClient#inspector!closed", "inspector closed")
	}
}

// RegisterService initializes service (without starting servers)
func RegisterService(namespace string) {
	redisAddr := fmt.Sprintf("%s:%d", redis.Instance().Master().Meta().Host, redis.Instance().Master().Meta().Port)
	scheduler = asynq.NewSchedulerWithNamespace(
		asynq.RedisClientOpt{Addr: redisAddr},
		namespace,
		&asynq.SchedulerOpts{
			Logger:   &_WorkerLogger{},
			Location: time.Local,
			LogLevel: asynq.InfoLevel,
		},
	)

	_RegisterTasks() // Register all tasks
	// Create scheduler manager (with distributed lock)
	schedulerManager = _NewSingleInstanceManager(namespace, "scheduler",
		&_SchedulerRunner{
			scheduler:    scheduler,
			registerFunc: _RegisterScheduledTasks,
		})

	// Start scheduler manager
	schedulerManager.Start()
}

func UnRegisterService() {
	// Stop scheduler manager
	if schedulerManager != nil {
		schedulerManager.Stop()
		schedulerManager = nil
	}

	// Clear scheduler reference
	if scheduler != nil {
		scheduler = nil
	}
}

// StartService starts worker service
func StartService(namespace string) {
	// Check if already started
	if len(srvList) > 0 || len(singleSrvList) > 0 {
		kklogger.WarnJ("worker:StartService#already!started", "service already started")
		return
	}

	redisAddr := fmt.Sprintf("%s:%d", redis.Instance().Master().Meta().Host, redis.Instance().Master().Meta().Port)

	// Concurrency
	srvList = append(srvList, asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 4,
				"share":    3,
				"default":  2,
				"low":      1,
			},
			Namespace: namespace,
			Logger:    &_WorkerLogger{},
		},
	))

	// Single
	srvList = append(srvList, asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 1, Queues: map[string]int{"emails": 1}, Namespace: namespace, Logger: &_WorkerLogger{}},
	))

	// Start normal srv
	for _, srv := range srvList {
		if err := srv.Start(serveMux); err != nil {
			panic(err)
		}
	}

	// Single server with distributed lock
	singleSrvList = append(singleSrvList, _NewSingleInstanceManager(namespace, "unique",
		&_ServerRunner{
			server: asynq.NewServer(
				asynq.RedisClientOpt{Addr: redisAddr},
				asynq.Config{Concurrency: 1, Queues: map[string]int{"unique": 1}, Namespace: namespace, Logger: &_WorkerLogger{}},
			),
			serveMux: serveMux,
		}))

	for _, srv := range singleSrvList {
		srv.Start()
	}
}

// StopService stops worker service
func StopService() {
	for _, srv := range singleSrvList {
		srv.Stop()
		//
		// worker.Submit(task)
		//
		//
		//
		//
		//
		//
	}

	for _, srv := range srvList {
		srv.Shutdown()
	}
}

// _WorkerLogger worker specific logger
type _WorkerLogger struct {
	kklogger.DefaultLoggerHook
}

func (l *_WorkerLogger) Fatal(args ...interface{}) {
}

// _InstanceRunner defines a runner which is managed as a single instance
type _InstanceRunner interface {
	Start() error
	Shutdown()
}

// _ServerRunner implements server runner
type _ServerRunner struct {
	server   *asynq.Server
	serveMux *asynq.ServeMux
}

func (r *_ServerRunner) Start() error {
	return r.server.Start(r.serveMux)
}

func (r *_ServerRunner) Shutdown() {
	r.server.Shutdown()
}

// _SchedulerRunner implements scheduler runner
type _SchedulerRunner struct {
	scheduler    *asynq.Scheduler
	registerFunc func(*asynq.Scheduler) []string // Register function and return entry IDs
	entryIDs     []string
	ctx          context.Context
	cancel       context.CancelFunc
	done         chan error
	running      bool // Flag scheduler running state
	mu           sync.Mutex
}

func (r *_SchedulerRunner) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Call register function to register all scheduled tasks
	if r.registerFunc != nil {
		r.entryIDs = r.registerFunc(r.scheduler)
		kklogger.InfoJ("worker:SchedulerRunner.Start#register!success", fmt.Sprintf("registered %d scheduled tasks", len(r.entryIDs)))
	}

	// Start scheduler (only when lock acquired)
	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.done = make(chan error, 1)
	r.running = true

	go func() {
		kklogger.InfoJ("worker:SchedulerRunner.Start#run!started", "scheduler.Run() started")
		r.done <- r.scheduler.Run()
		kklogger.InfoJ("worker:SchedulerRunner.Start#run!stopped", "scheduler.Run() stopped")
	}()

	return nil
}

func (r *_SchedulerRunner) Shutdown() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. Unregister all entries from scheduler
	for _, entryID := range r.entryIDs {
		if err := r.scheduler.Unregister(entryID); err != nil {
			kklogger.WarnJ("worker:SchedulerRunner.Shutdown#unregister!failed", fmt.Sprintf("entryID=%s error=%s", entryID, err.Error()))
		}
	}
	r.entryIDs = nil
	kklogger.InfoJ("worker:SchedulerRunner.Shutdown#unregister!success", "unregistered all scheduled tasks")

	// 2. Delete all scheduled tasks from Redis
	if inspector != nil {
		queues, err := inspector.Queues()
		if err != nil {
			kklogger.ErrorJ("worker:SchedulerRunner.Shutdown#list_queues!failed", err.Error())
		} else {
			totalDeleted := 0
			for _, queue := range queues {
				// Delete all scheduled tasks from the queue
				deleted, err := inspector.DeleteAllScheduledTasks(queue)
				if err != nil {
					kklogger.ErrorJ("worker:SchedulerRunner.Shutdown#delete_scheduled!failed", fmt.Sprintf("queue=%s error=%s", queue, err.Error()))
				} else {
					totalDeleted += deleted
				}
			}
			kklogger.InfoJ("worker:SchedulerRunner.Shutdown#delete_scheduled!success", fmt.Sprintf("deleted %d scheduled tasks", totalDeleted))
		}
	}

	// 3. Stop scheduler (only if running)
	if r.running {
		kklogger.InfoJ("worker:SchedulerRunner.Shutdown#shutdown!started", "stopping scheduler")
		r.scheduler.Shutdown()
		if r.done != nil {
			<-r.done
		}
		r.running = false
		kklogger.InfoJ("worker:SchedulerRunner.Shutdown#shutdown!completed", "scheduler stopped")
	} else {
		kklogger.InfoJ("worker:SchedulerRunner.Shutdown#shutdown!skipped", "scheduler was not running")
	}
}

// _SingleInstanceManager manages a distributed lock for single instance runner
type _SingleInstanceManager struct {
	lockKey  string
	hostname string
	runner   _InstanceRunner
	ctx      context.Context
	cancel   context.CancelFunc
	done     chan struct{}
	acquired bool
	mu       sync.Mutex
}

func _NewSingleInstanceManager(namespace, name string, runner _InstanceRunner) *_SingleInstanceManager {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &_SingleInstanceManager{
		lockKey:  redis.Key("WORKER", fmt.Sprintf("SIL:%s:%s", namespace, name)),
		hostname: hostname,
		runner:   runner,
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
	}
}

func (m *_SingleInstanceManager) Start() {
	go m.manage()
}

func (m *_SingleInstanceManager) Stop() {
	m.cancel()
	<-m.done
	m.release()
}

func (m *_SingleInstanceManager) manage() {
	// Try to acquire lock immediately
	m.mu.Lock()
	m.tryAcquire()
	m.mu.Unlock()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	defer close(m.done)

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.mu.Lock()
			if m.acquired {
				m.keepAlive()
			} else {
				m.tryAcquire()
			}
			m.mu.Unlock()
		}
	}
}

func (m *_SingleInstanceManager) tryAcquire() {
	result := redis.Master().SetNX(m.lockKey, m.hostname)
	if result.Error != nil {
		kklogger.ErrorJ("worker:SingleInstanceManager.tryAcquire#setnx!failed", result.Error.Error())
		return
	}

	if result.GetInt64() == 1 {
		expireResult := redis.Master().SetExpire(m.lockKey, m.hostname, 6)
		if expireResult.Error != nil {
			kklogger.ErrorJ("worker:SingleInstanceManager.tryAcquire#expire!failed", expireResult.Error.Error())
			redis.Master().Delete(m.lockKey)
			return
		}

		m.acquired = true
		if err := m.runner.Start(); err != nil {
			kklogger.ErrorJ("worker:SingleInstanceManager.tryAcquire#start!failed", err.Error())
			m.acquired = false
			redis.Master().Delete(m.lockKey)
		} else {
			kklogger.InfoJ("worker:SingleInstanceManager.tryAcquire#lock!acquired", fmt.Sprintf("hostname=%s lock=%s", m.hostname, m.lockKey))
		}
	}
}

func (m *_SingleInstanceManager) keepAlive() {
	currentValue := redis.Master().Get(m.lockKey)
	if currentValue.Error != nil || currentValue.GetString() != m.hostname {
		m.acquired = false
		m.runner.Shutdown()
		kklogger.WarnJ("worker:SingleInstanceManager.keepAlive#lock!lost", fmt.Sprintf("hostname=%s lock=%s", m.hostname, m.lockKey))
		return
	}

	result := redis.Master().SetExpire(m.lockKey, m.hostname, 6)
	if result.Error != nil {
		kklogger.ErrorJ("worker:SingleInstanceManager.keepAlive#renew!failed", result.Error.Error())
		m.acquired = false
		m.runner.Shutdown()
		kklogger.WarnJ("worker:SingleInstanceManager.keepAlive#lock!lost_on_renew", fmt.Sprintf("hostname=%s lock=%s", m.hostname, m.lockKey))
	}
}

func (m *_SingleInstanceManager) release() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.acquired {
		return
	}

	currentValue := redis.Master().Get(m.lockKey)
	if currentValue.Error == nil && currentValue.GetString() == m.hostname {
		if err := redis.Master().Delete(m.lockKey).Error; err != nil {
			kklogger.ErrorJ("worker:SingleInstanceManager.release#delete!failed", err.Error())
		}
	}

	m.acquired = false
	m.runner.Shutdown()
	kklogger.InfoJ("worker:SingleInstanceManager.release#lock!released", fmt.Sprintf("hostname=%s lock=%s", m.hostname, m.lockKey))
}
