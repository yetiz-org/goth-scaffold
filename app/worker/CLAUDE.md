# app/worker

## Task Type Names & Payloads

All task type name constants **and** shared Payload structs live in **one place only**: `tasks/taskdefs/`.

This is the single source of truth for the entire task system. `app/worker/tasks/`, `app/services/`, and `app/handlers/` all import from here — no duplication.

When adding a new task:
1. **First** add the name constant (and Payload struct if needed) in `tasks/taskdefs/` under the appropriate group.
2. **Then** implement the handler in `tasks/`.

The `Name()` method MUST return the constant directly:

```go
func (t *Xxx) Name() string { return taskdefs.Xxx }
```

Never hardcode a task name string directly in `Name()` or anywhere else.

---

## Job File Structure

Each job lives in `tasks/` as a single `.go` file. The file contains:
1. A comment block at the top describing task name, goal, and flow.
2. The Handler struct.
3. Handler interface method overrides.
4. `Run` and any private helper methods.

## Handler

```go
type Xxx struct {
    internal.DefaultHandler
}
```

Override only methods that differ from `DefaultHandler` defaults:
- `Name() string` — MUST return `taskdefs.Xxx`.
- `Description() string` — one-line English description.
- `Payload() internal.Payload` — return `&taskdefs.XxxPayload{}`.
- `Retention() time.Duration` — set when result must be inspectable after completion.
- `TaskID() string` — return `t.Name()` to enforce single-instance dedup; return `""` for parallel execution.
- `Timeout() time.Duration` — override only when the default 60-minute limit is wrong.
- `MaxRetry() int` — override only when retry behaviour differs (default: 3).
- `Queue() string` — override to route to a specific queue (default: asynq default).

Do not override `Before`, `After`, or `ErrorCaught` unless the handler needs specific lifecycle behaviour.

## Payload

```go
// In tasks/taskdefs/xxx.go:
type XxxPayload struct {
    internal.BasePayload
    Field *Type `json:"field,omitempty" required:"false" desc:"Description."`
}

func (p *XxxPayload) Encode() (string, error) { return p.BasePayload.EncodePayload(p) }
func (p *XxxPayload) Decode(data string) error { return p.BasePayload.DecodePayload(data, p) }
```

Rules:
- MUST embed `internal.BasePayload`.
- MUST override `Encode` and `Decode` when the struct has fields beyond `BasePayload`.
- Use pointer types (`*string`, `*int`) for optional fields.
- Include `json`, `required`, and `desc` struct tags on every field.
- Tasks with no parameters may omit a custom Payload; the handler's `Payload()` then returns `&internal.BasePayload{}`.

## DryRun Support

`DryRun bool` is already declared in `internal.BasePayload` and is available on every payload via embedding. Do NOT redeclare it.

When a job writes to the database or triggers external side effects, it MUST support DryRun.

### Step 1 — Expose the field via admin API

Implement this empty method on the Payload so the admin API shows the `dry_run` field to callers:

```go
func (p *XxxPayload) SupportDryRun() {}
```

### Step 2 — Guard all writes in Run

`payload.DryRun` is accessible directly via the embedded `BasePayload`.

```go
if payload != nil && payload.DryRun {
    taskInfo.WriteDryRunSuccess(map[string]any{
        "would_affect": count,
    })
    return nil
}

// real writes happen here
```

Rules:
- DryRun MUST skip all DB writes, external API calls, and file mutations.
- DryRun MUST still execute the full read and computation path.
- Call `taskInfo.WriteDryRunSuccess(data)` — do not call `WriteSuccess` for dry run paths.

## Result Reporting

Every `Run` exit path MUST call exactly one of:

| Outcome | Call |
|---------|------|
| Success | `taskInfo.WriteSuccess(message, data)` |
| DryRun | `taskInfo.WriteDryRunSuccess(data)` |
| Error | `taskInfo.WriteError(errCode, message, nil)` then `return err` |
| Nothing to do | `taskInfo.WriteSuccess("no_data", map[string]any{...})` then `return nil` |

Never return without calling one of the above when `Retention() > 0`.

## Task Creation

To create a task for submission, use `worker.NewTask` with the taskdefs constant:

```go
// Outside tasks package (services, handlers, bootstrap/register.go):
task := worker.NewTask(taskdefs.Xxx, &taskdefs.XxxPayload{Field: value})
worker.Submit(task, asynq.ProcessIn(delay))

// Inside a task's Run method (self-dispatch or chaining):
task := worker.NewTask(taskdefs.Xxx, &taskdefs.XxxPayload{Field: value})
internal.Submit(task, asynq.ProcessIn(delay))
```

Pass `nil` as payload when callers use all defaults (Run will apply defaults internally).

Do NOT define `NewXxxTask` factory functions — they create import cycles when services or handlers need to dispatch tasks.

## Registration

After adding a task:

1. Add `worker.Register(&tasks.Xxx{})` to `RegisterTasks()` in `bootstrap/register.go`.
2. If the task runs on a schedule, add `worker.RegisterSchedule(scheduler, "<cron>", worker.NewTask(taskdefs.Xxx, nil))` to `RegisterScheduledTasks()` in `bootstrap/register.go`.
3. Wrap scheduled registrations in environment guards when the job must not run in dev/staging.

## Submission Helpers

| Function | Purpose |
|----------|---------|
| `worker.Submit(task, opts...)` | Enqueue with caller opts overriding handler defaults |
| `worker.SubmitCritical(task, opts...)` | Enqueue into the critical queue |
| `worker.SubmitUnique(task, ttl, opts...)` | Enqueue with deduplication |

`SubmitUnique` uses Redis for distributed dedup. Pick a `uniqueTTL` that covers the expected run time.

## Queues

| Queue      | Priority | Use for                         |
|------------|----------|---------------------------------|
| `critical` | 4        | User-facing real-time ops       |
| `share`    | 3        | Cross-service interactions      |
| `default`  | 2        | Standard background work        |
| `low`      | 1        | Bulk / non-urgent jobs          |
| `emails`   | 1        | Serialised mail delivery        |
| `unique`   | 1        | Single-instance scheduled tasks |

Override queue with `Queue() string` in the handler, or pass `asynq.Queue("critical")` at Submit time.

## Asynqmon

Web UI for inspecting task queues: **http://localhost:8081** (when `make local-env-start` is running).

Displays queued / active / scheduled / retry / dead tasks with full payload and result.

## Error Handling & Retry

- Return `nil` → task marked **completed**.
- Return non-nil `error` → asynq schedules a retry with exponential backoff.
- Default max retries: **3** (set in `DefaultHandler.MaxRetry`). Override per-handler:

```go
func (h *FooHandler) MaxRetry() int { return 0 } // no retry
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

```go
kklogger.InfoJ("worker:Xxx.Run#start", map[string]any{"param": value})
kklogger.ErrorJ("worker:Xxx.Run#save!db_error", err.Error())
```

Log at `InfoJ` before and after each major phase. Log at `ErrorJ` before returning any error.

## Validation Commands

```bash
# Compile check after adding a new task
go build ./app/worker/...

# Check all tasks are registered in bootstrap/register.go
rg -n "worker.Register\(" app/worker/bootstrap/register.go

# Check all tasks implement Name() method
rg -n "func \(.*\) Name\(\)" app/worker/tasks/ --type go

# Check taskdefs constants used (no hardcoded strings in Name())
rg -n "return \"" app/worker/tasks/ --type go

# Check DryRun guard pattern used in side-effecting tasks
rg -l "SupportDryRun\|WriteDryRunSuccess" app/worker/tasks/ --type go
```
