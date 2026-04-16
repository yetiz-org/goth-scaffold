package worker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yetiz-org/asynq"
	"github.com/yetiz-org/gone/erresponse"
	"github.com/yetiz-org/gone/ghttp"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/components/slack"
	"github.com/yetiz-org/goth-scaffold/app/conf"
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
	handlerRegistry  = map[string]internal.Handler{}
	shuttingDown     atomic.Bool
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

func _isCanceledByShutdown(ctx context.Context, err error) bool {
	if !shuttingDown.Load() {
		return false
	}

	if errors.Is(err, context.Canceled) {
		return true
	}

	if ctx != nil && errors.Is(ctx.Err(), context.Canceled) {
		return true
	}

	return false
}

func _notifyTaskCanceledByShutdown(taskType string, stage string, err error) {
	hostname, _ := os.Hostname()

	notification := "[Worker Shutdown] Task canceled before completion\n"
	notification += fmt.Sprintf("- env=%s\n", conf.Config().App.Environment.Lower())
	notification += fmt.Sprintf("- host=%s\n", hostname)
	notification += fmt.Sprintf("- task=%s\n", taskType)
	notification += fmt.Sprintf("- stage=%s\n", stage)
	notification += fmt.Sprintf("- error=%s", err.Error())

	webhookURL := conf.Config().Credentials.SecretSlack.Webhook
	if sendErr := slack.NewClient(webhookURL).SendWithTimeout(notification, 2*time.Second); sendErr != nil {
		if errors.Is(sendErr, slack.ErrWebhookNotConfigured) {
			kklogger.WarnJ("worker:Service.notifyTaskCanceledByShutdown#config!missing_webhook", "Slack webhook URL not configured")
			return
		}

		kklogger.ErrorJ("worker:Service.notifyTaskCanceledByShutdown#request!failed", sendErr.Error())
		return
	}
}

func (t *_ServeHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	executionContext := make(map[string]any)

	// Use handler.Payload() to get an empty instance and decode into it
	payload := t.handler.Payload()
	if err := payload.Decode(string(task.Payload())); err != nil {
		kklogger.WarnJ("worker:Service.ProcessTask#decode!failed", err.Error())
	}

	taskInfo := internal.NewTaskInfo(task.Type(), payload, task.ResultWriter())

	if err := t.handler.Before(ctx, taskInfo, executionContext); err != nil {
		if _isCanceledByShutdown(ctx, err) {
			kklogger.WarnJ("worker:Service.ProcessTask#before!canceled_by_shutdown", fmt.Sprintf("task_type=%s error=%s", task.Type(), err.Error()))
			_notifyTaskCanceledByShutdown(task.Type(), "before", err)
		} else {
			kklogger.ErrorJ("worker:Service.ProcessTask#before!failed", err.Error())
		}

		return t.handler.ErrorCaught(ctx, taskInfo, executionContext, t._WrapError(err))
	}

	if err := t.handler.Run(ctx, taskInfo, executionContext); err != nil {
		if _isCanceledByShutdown(ctx, err) {
			kklogger.WarnJ("worker:Service.ProcessTask#run!canceled_by_shutdown", fmt.Sprintf("task_type=%s error=%s", task.Type(), err.Error()))
			_notifyTaskCanceledByShutdown(task.Type(), "run", err)
		} else {
			kklogger.ErrorJ("worker:Service.ProcessTask#run!failed", err.Error())
		}

		return t.handler.ErrorCaught(ctx, taskInfo, executionContext, t._WrapError(err))
	}

	if err := t.handler.After(ctx, taskInfo, executionContext); err != nil {
		if _isCanceledByShutdown(ctx, err) {
			kklogger.WarnJ("worker:Service.ProcessTask#after!canceled_by_shutdown", fmt.Sprintf("task_type=%s error=%s", task.Type(), err.Error()))
			_notifyTaskCanceledByShutdown(task.Type(), "after", err)
		} else {
			kklogger.ErrorJ("worker:Service.ProcessTask#after!failed", err.Error())
		}

		return t.handler.ErrorCaught(ctx, taskInfo, executionContext, t._WrapError(err))
	}

	taskInfo.FlushResult()
	return nil
}

// Register registers a task handler. The handler's Name() is used as the pattern.
func Register(handler internal.Handler) {
	pattern := handler.Name()
	handlerRegistry[pattern] = handler
	serveMux.Handle(pattern, &_ServeHandler{handler: handler})
}

// RegisterSchedule registers a cron task.
// Options are automatically sourced from the handlerRegistry (including Retention).
func RegisterSchedule(scheduler *asynq.Scheduler, cronSpec string, task *asynq.Task) string {
	var opts []asynq.Option

	if handler, ok := handlerRegistry[task.Type()]; ok {
		opts = handlerToAsynqOptions(handler)
	}

	entryID, _ := scheduler.Register(cronSpec, task, opts...)
	return entryID
}

// mergeOptions merges options; submitOpts take precedence over handlerOpts.
func mergeOptions(handlerOpts, submitOpts []asynq.Option) []asynq.Option {
	var result []asynq.Option
	result = append(result, handlerOpts...)
	result = append(result, submitOpts...)
	return result
}

// Submit enqueues a task; caller opts override handler defaults.
func Submit(task *asynq.Task, opts ...asynq.Option) (*TaskFuture, error) {
	var handlerOpts []asynq.Option

	if handler, ok := handlerRegistry[task.Type()]; ok {
		handlerOpts = handlerToAsynqOptions(handler)
	}

	finalOpts := mergeOptions(handlerOpts, opts)

	if taskInfo, err := client.Enqueue(task, finalOpts...); err != nil {
		return nil, err
	} else {
		return NewTaskFuture(Inspector(), taskInfo), nil
	}
}

// SubmitCritical enqueues a task into the critical queue.
func SubmitCritical(task *asynq.Task, opts ...asynq.Option) (*TaskFuture, error) {
	return Submit(task, append(opts, asynq.Queue("critical"))...)
}

// SubmitUnique enqueues a task with deduplication.
func SubmitUnique(task *asynq.Task, uniqueTTL time.Duration, opts ...asynq.Option) (*TaskFuture, error) {
	return Submit(task, append(opts, asynq.Unique(uniqueTTL))...)
}

// handlerToAsynqOptions converts Handler method values to asynq.Option slice.
func handlerToAsynqOptions(handler internal.Handler) []asynq.Option {
	var opts []asynq.Option

	if handler.Retention() > 0 {
		opts = append(opts, asynq.Retention(handler.Retention()))
	}

	if handler.UniqueTTL() > 0 {
		opts = append(opts, asynq.Unique(handler.UniqueTTL()))
	}

	if handler.TaskID() != "" {
		opts = append(opts, asynq.TaskID(handler.TaskID()))
	}

	if handler.Timeout() > 0 {
		opts = append(opts, asynq.Timeout(handler.Timeout()))
	}

	if handler.Queue() != "" {
		opts = append(opts, asynq.Queue(handler.Queue()))
	}

	if handler.MaxRetry() >= 0 {
		opts = append(opts, asynq.MaxRetry(handler.MaxRetry()))
	}

	if handler.Group() != "" {
		opts = append(opts, asynq.Group(handler.Group()))
	}

	return opts
}

// GetRegisteredHandlers returns all registered handlers (for API introspection).
func GetRegisteredHandlers() map[string]internal.Handler {
	return handlerRegistry
}

// Execute runs a task directly without enqueuing.
func Execute(ctx context.Context, task *asynq.Task) error {
	if handler, ok := handlerRegistry[task.Type()]; ok {
		if handler.Timeout() > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, handler.Timeout())
			defer cancel()
		}
	}

	return serveMux.ProcessTask(ctx, task)
}

// Inspector returns the task inspector.
func Inspector() *asynq.Inspector {
	return inspector
}

// Scheduler returns the task scheduler.
func Scheduler() *asynq.Scheduler {
	return scheduler
}

// StartClient starts Client and Inspector, and injects the submitter into internal package.
func StartClient(namespace string) {
	if client != nil || inspector != nil {
		kklogger.WarnJ("worker:StartClient#already!started", "client already started")
		return
	}

	redisAddr := fmt.Sprintf("%s:%d", redis.Instance().Master().Meta().Host, redis.Instance().Master().Meta().Port)
	client = asynq.NewClientWithNamespace(asynq.RedisClientOpt{Addr: redisAddr}, namespace)
	inspector = asynq.NewInspectorWithNamespace(asynq.RedisClientOpt{Addr: redisAddr}, namespace)
	internal.SetSubmitter(func(task *asynq.Task, opts ...asynq.Option) error {
		_, err := Submit(task, opts...)
		return err
	})
	kklogger.InfoJ("worker:StartClient#started", fmt.Sprintf("namespace=%s", namespace))
}

// StopClient stops Client and Inspector.
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

// RegisterService initializes service without starting servers.
// registerTasks registers all task handlers; registerScheduledTasks registers all cron tasks.
// Both functions are injected from outside to avoid worker package depending on tasks package directly.
func RegisterService(namespace string, registerTasks func(), registerScheduledTasks func(*asynq.Scheduler) []string) {
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

	// Register all business tasks (injected from outside)
	registerTasks()

	// Create scheduler manager with distributed lock
	schedulerManager = _NewSingleInstanceManager(namespace, "scheduler",
		&_SchedulerRunner{
			scheduler:    scheduler,
			registerFunc: registerScheduledTasks,
		})

	schedulerManager.Start()
}

func UnRegisterService() {
	if schedulerManager != nil {
		schedulerManager.Stop()
		schedulerManager = nil
	}

	if scheduler != nil {
		scheduler = nil
	}
}

// StartService starts the worker servers.
func StartService(namespace string) {
	if len(srvList) > 0 || len(singleSrvList) > 0 {
		kklogger.WarnJ("worker:StartService#already!started", "service already started")
		return
	}

	redisAddr := fmt.Sprintf("%s:%d", redis.Instance().Master().Meta().Host, redis.Instance().Master().Meta().Port)

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

	srvList = append(srvList, asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 1, Queues: map[string]int{"emails": 1}, Namespace: namespace, Logger: &_WorkerLogger{}},
	))

	for _, srv := range srvList {
		if err := srv.Start(serveMux); err != nil {
			panic(err)
		}
	}

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

// StopService stops all worker servers.
func StopService() {
	shuttingDown.Store(true)

	for _, srv := range singleSrvList {
		srv.Stop()
	}

	for _, srv := range srvList {
		srv.Shutdown()
	}
}

// _WorkerLogger is the worker-specific logger adapter.
type _WorkerLogger struct {
	kklogger.DefaultLoggerHook
}

func (l *_WorkerLogger) Fatal(args ...any) {}

// _InstanceRunner defines a component that can be managed as a single instance.
type _InstanceRunner interface {
	Start() error
	Shutdown()
}

// _ServerRunner wraps asynq.Server as _InstanceRunner.
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

// _SchedulerRunner wraps asynq.Scheduler as _InstanceRunner.
type _SchedulerRunner struct {
	scheduler    *asynq.Scheduler
	registerFunc func(*asynq.Scheduler) []string
	entryIDs     []string
	ctx          context.Context
	cancel       context.CancelFunc
	done         chan error
	running      bool
	mu           sync.Mutex
}

func (r *_SchedulerRunner) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.registerFunc != nil {
		r.entryIDs = r.registerFunc(r.scheduler)
		kklogger.InfoJ("worker:SchedulerRunner.Start#register!success", fmt.Sprintf("registered %d scheduled tasks", len(r.entryIDs)))
	}

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

	for _, entryID := range r.entryIDs {
		if err := r.scheduler.Unregister(entryID); err != nil {
			kklogger.WarnJ("worker:SchedulerRunner.Shutdown#unregister!failed", fmt.Sprintf("entryID=%s error=%s", entryID, err.Error()))
		}
	}

	r.entryIDs = nil
	kklogger.InfoJ("worker:SchedulerRunner.Shutdown#unregister!success", "unregistered all scheduled tasks")

	if inspector != nil {
		queues, err := inspector.Queues()
		if err != nil {
			kklogger.ErrorJ("worker:SchedulerRunner.Shutdown#list_queues!failed", err.Error())
		} else {
			totalDeleted := 0
			for _, queue := range queues {
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

// _SingleInstanceManager manages a distributed lock for a single-instance runner.
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
	m.mu.Lock()
	m.tryAcquire()
	m.mu.Unlock()

	ticker := time.NewTicker(5 * time.Second)
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
		expireResult := redis.Master().SetExpire(m.lockKey, m.hostname, 15)
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

	result := redis.Master().SetExpire(m.lockKey, m.hostname, 15)
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
