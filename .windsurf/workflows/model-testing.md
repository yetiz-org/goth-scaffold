---
description: Model Unit Test Generation - Complete AI Agent Workflow
auto_execution_mode: 3
---

# Model Unit Testing Workflow

## RULES
1. **Pure logic only** - no DB, network, HTTP, daemons
2. **Deterministic** - no environment-dependent assertions
3. **Global state** - restore after modification, no `t.Parallel()` if mutated
4. **Documentation** - package doc with test list, function doc with scenario
5. **English only** - all comments and documentation
6. **Verification** - `go test -v ./tests/units/models/...` must 100% PASS

---

## WORKFLOW

### 1. Classify
```
Pure method on struct       → Model Unit Test
DB/GORM operations          → Repository Integration Test
HTTP/session/user flow      → E2E
```

### 2. Inspect Model
// turbo
```bash
grep -n "type {Model} struct" app/models/*.go
grep -n "func (m \\*{Model})" app/models/*.go
```

### 3. Check Existing Tests
// turbo
```bash
ls tests/units/models/ 2>/dev/null
```

### 4. Generate Test File
// turbo
Create `tests/units/models/{model}_test.go`

### 5. Design Cases
- **TableName()** - prevent table mapping regression
- **Computed methods** - string building, fallback behavior
- **Crypto wrappers** - round-trip (encrypt → decrypt)
- Skip trivial getters

### 6. Run & Fix
// turbo
```bash
go test -v ./tests/units/models/...
go test -v ./...
```

---

## TEMPLATE

**Package doc (required):**
```go
/*
Package models - {Model} Unit Tests
Test Case List:
  - Test{Model}_{Method}_{Scenario}: {description}
*/
package models_test
```

**Function doc (required):**
```go
// Test{Model}_{Method}_{Scenario} tests {description}
// Scenario: Arrange → Act → Assert
// Note: Restore global state if modified; no t.Parallel() with shared state
func Test{Model}_{Method}_{Scenario}(t *testing.T) {}
```
