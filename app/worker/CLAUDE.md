# app/worker

## Task Structure

Every task lives in `app/worker/tasks/` and follows this skeleton:

```go
package tasks

import (
	"context"

	"github.com/yetiz-org/asynq"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/worker/internal"
)

const taskTypeFoo = "domain:foo_action"

type FooHandler struct{ internal.DefaultHandler }

func (h *FooHandler) Name() string { return taskTypeFoo }

// NewFooTask creates the asynq task. Call worker.Submit(NewFooTask(...)) to enqueue.
func NewFooTask( /* params */) *asynq.Task {
	return internal.NewTask(taskTypeFoo, nil /* or payload struct */)
}

func (h *FooHandler) Run(ctx context.Context, taskInfo *internal.TaskInfo, execCtx map[string]any) error {
	kklogger.InfoJ("worker:FooHandler.Run#exec!start", nil)

	// ... business logic ...

	taskInfo.Write([]byte("done"))
	return nil
}

```

Register in `app/worker/register.go`:

```go
Register((&tasks.FooHandler{}).Name(), &tasks.FooHandler{})
```

## Unique / Deduplication

To guarantee at-most-one execution per logical entity:

```go
worker.SubmitUnique(NewFooTask(id), 10*time.Minute)
```

`SubmitUnique` uses Redis for distributed dedup. Pick a `uniqueTTL` that covers the expected run time.

## Scheduled Tasks

Register in `_RegisterScheduledTasks`:

```go
entryIDs = append(entryIDs, RegisterSchedule(scheduler, "0 * * * *", tasks.NewFooTask()))
```

The scheduler uses a distributed lock so only one node fires scheduled tasks.

## Queues

| Queue      | Priority | Use for                         |
|------------|----------|---------------------------------|
| `critical` | 4        | User-facing real-time ops       |
| `share`    | 3        | Cross-service interactions      |
| `default`  | 2        | Standard background work        |
| `low`      | 1        | Bulk / non-urgent jobs          |
| `emails`   | 1        | Serialised mail delivery        |
| `unique`   | 1        | Single-instance scheduled tasks |

Override with `asynq.Queue("critical")` option in `Submit`.

## Asynqmon

Web UI for inspecting task queues: **http://localhost:8081** (when `make env-up` is running).

Displays queued / active / scheduled / retry / dead tasks with full payload and result.

## Error Handling & Retry

- Return `nil` → task marked **completed**.
- Return non-nil `error` → asynq schedules a retry with exponential backoff.
- Default max retries: **25** (asynq global default). Override per-task at submission:

```go
worker.Submit(NewFooTask(id), asynq.MaxRetry(3))
```

- To fail immediately without retry, wrap with `asynq.SkipRetry`:

```go
return fmt.Errorf("invalid payload: %w", asynq.SkipRetry)
```

- Tasks that exhaust all retries move to the **Dead** queue — visible in Asynqmon at http://localhost:8081.

## Logging Format

```
worker:HandlerName.Method#section!action
```
