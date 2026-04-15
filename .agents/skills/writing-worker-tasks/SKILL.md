---
name: writing-worker-tasks
description: Use when adding, updating, testing, or reviewing asynq worker tasks under app/worker/tasks/, including scheduled tasks, enqueue paths, payload design, dry-run support, or worker registration.
---

# Writing Worker Tasks

## Core Rule

Copy the repo's worker contract before copying business logic.
Task registration, payload metadata, dry-run support, and enqueue path selection are part of the contract.

## When to Use

- Add a task under `app/worker/tasks/`.
- Change task payload, queue, retry, timeout, or schedule.
- Add enqueue calls from handlers, services, or other tasks.
- Update worker tests under `tests/units/worker/`.

Don't use when:
- The work is a one-off script with no queue/retry value.
- The operation must stay synchronous within the HTTP request.

## Read Order (before editing)

1. `app/worker/register.go` — registered task names and handlers.
2. `app/worker/internal/handler.go` — `DefaultHandler` contract and defaults.
3. `app/worker/service.go` — `Submit` and `SubmitUnique` enqueue helpers.
4. One existing task in `app/worker/tasks/` — use it as concrete reference.
5. Existing tests in `tests/units/worker/`.

## Task Skeleton

```go
package tasks

import (
    "context"

    "github.com/yetiz-org/asynq"
    kklogger "github.com/yetiz-org/goth-kklogger"
    "github.com/example/myapp/app/worker/internal"
)

// Task: send a welcome email after user registration.
// Flow:
//   1. Load user by ID.
//   2. Send email via mailer service.

const taskTypeSendWelcomeEmail = "user:send_welcome_email"

type SendWelcomeEmailPayload struct {
    internal.BasePayload
    UserID uint64 `json:"user_id" required:"true" desc:"User to welcome"`
}

type SendWelcomeEmailHandler struct{ internal.DefaultHandler }

func (h *SendWelcomeEmailHandler) Name() string { return taskTypeSendWelcomeEmail }

// NewSendWelcomeEmailTask creates the asynq task.
// Enqueue via: worker.Submit(NewSendWelcomeEmailTask(userID))
func NewSendWelcomeEmailTask(userID uint64) *asynq.Task {
    return internal.NewTask(taskTypeSendWelcomeEmail, SendWelcomeEmailPayload{UserID: userID})
}

func (h *SendWelcomeEmailHandler) Run(ctx context.Context, taskInfo *internal.TaskInfo, execCtx map[string]any) error {
    kklogger.InfoJ("worker:SendWelcomeEmailHandler.Run#exec!start", nil)

    payload := &SendWelcomeEmailPayload{}
    if err := taskInfo.UnmarshalPayload(payload); err != nil {
        return err
    }

    // business logic ...

    taskInfo.Write([]byte("sent"))
    return nil
}
```

Register in `app/worker/register.go`:

```go
Register((&tasks.SendWelcomeEmailHandler{}).Name(), &tasks.SendWelcomeEmailHandler{})
```

## Payload Rules

- Embed `internal.BasePayload` for dry_run and standard fields.
- Use `json`, `required`, `desc` struct tags — these are reflected into admin metadata.
- One payload struct per task type — do not reuse payload structs across tasks.

## Dry-Run Support

When a task supports safe simulation, implement `internal.DryRunSupporter` on the payload:

```go
func (p *FooPayload) IsDryRun() bool { return p.DryRun }
```

Branch before side effects:

```go
if payload.IsDryRun() {
    taskInfo.Write([]byte("dry-run: would send email"))
    return nil
}
```

## Enqueue

```go
// One-time
worker.Submit(tasks.NewFooTask(id))

// Deduplicated (at-most-once per TTL window)
worker.SubmitUnique(tasks.NewFooTask(id), 10*time.Minute)
```

Use `SubmitUnique` for tasks that must not run in parallel for the same resource.

## Scheduled Tasks

In `_RegisterScheduledTasks` (app/worker/scheduler.go or similar):

```go
entryIDs = append(entryIDs, RegisterSchedule(scheduler, "0 * * * *", tasks.NewFooTask()))
```

## Logging

```go
kklogger.InfoJ("worker:HandlerName.Run#exec!start", nil)
kklogger.ErrorJ("worker:HandlerName.Run#exec!db_error", err.Error())
```

Format: `worker:HandlerName.Run#section!action` — English only.

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| Not embedding `BasePayload` | Always embed for standard dry_run and metadata |
| Reusing payload struct across tasks | One payload per task type |
| Business logic before `UnmarshalPayload` | Unmarshal first, then logic |
| Not registering in `register.go` | Task won't be picked up by worker |
| Missing task-level comment block | Add `// Task:` / `// Flow:` before the const |
| Skipping dry-run branch | Add `IsDryRun()` check before any side effect |
