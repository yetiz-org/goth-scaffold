---
description: Repository Integration Test Generation - Complete AI Agent Workflow
auto_execution_mode: 3
---

# Repository Integration Testing Workflow

## GOLDEN RULES
1. **Parallel execution**: ALL tests MUST call `t.Parallel()` at the first line
2. **Test isolation**: Use unique IDs via atomic counter - NEVER hardcode IDs
3. **Copy existing patterns**: Reference actual tests in `tests/units/repositories/*_test.go`
4. **Run tests**: Work incomplete until `go test -v -parallel 4 ./tests/units/repositories/...` shows 100% PASS
5. **All content in English**: All comments, logs, and documentation must be in English
6. **Cleanup required**: Use `defer` for cleanup immediately after data creation

---

## WORKFLOW

### 1. Verify Repository Under Test
// turbo
```bash
grep -r "func.*Repository" app/repositories/ --include="*.go" | grep -i "{repository_name}"
```

### 2. Check Existing Tests
// turbo
```bash
ls -la tests/units/repositories/
cat tests/units/repositories/{repository_name}_test.go 2>/dev/null || echo "No existing test file"
```

### 3. Reference Model Structure
// turbo
```bash
grep -A 50 "type {Model}.*struct" app/models/*.go
```

### 4. Generate Test File
// turbo
1. Create `tests/units/repositories/{repository}_test.go`
2. Copy template from Section A
3. Name: `Test{Repository}_{Operation}_{Scenario}`

### 5. Run & Fix
// turbo
```bash
go test -v -parallel 4 ./tests/units/repositories/...
go test -v -parallel 8 -count 3 ./tests/units/repositories/...  # Race check
```
Fix failures → check Section C → re-run until 100% PASS

---

## REFERENCE

### A. Templates

**Package Header**:
```go
/*
Package repositories - {Repository} Integration Tests

Test Case List:
  - Test{Repository}_{Operation}_{Scenario}: {Description}
*/
package repositories
```

**Imports**:
```go
import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"{module}/app/connector/database"
	"{module}/app/models"
	"{module}/app/repositories"
)
```

**Unique ID Generator**:
```go
var {resource}TestIDCounter int64 = {starting_value}

func next{Resource}TestID() int64 {
	return atomic.AddInt64(&{resource}TestIDCounter, 1)
}

func unique{Resource}String(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, next{Resource}TestID())
}
```

**Cleanup Helper**:
```go
func delete{Resource}(t *testing.T, id int64) {
	t.Helper()
	database.SSWriter().Unscoped().Delete(&models.{Model}{}, id)
}

func delete{Resources}(t *testing.T, ids []int64) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	database.SSWriter().Unscoped().Where("id IN ?", ids).Delete(&models.{Model}{})
}
```

**Test Function**:
```go
func Test{Repository}_{Operation}_{Scenario}(t *testing.T) {
	t.Parallel()
	
	// Arrange
	db := database.SSWriter()
	repo := repositories.New{Repository}(db)
	targetID := next{Resource}TestID()
	
	record := &models.{Model}{ /* fields */ }
	require.NoError(t, db.Create(record).Error, "setup: create record")
	defer delete{Resource}(t, record.ID)
	
	// Act
	result := repo.Get(record.ID)
	
	// Assert
	require.NotNil(t, result, "should return record")
	assert.Equal(t, record.ID, result.ID, "ID should match")
}
```

### B. ID Generator Rules

- Each test file uses unique starting value to avoid conflicts
- Decrement by 100 for each new file (e.g., 9999900, 9999800, 9999700...)
- Check existing files to determine next available range

### C. Common Fixes

| ❌ Wrong | ✅ Correct |
|---------|-----------|
| Missing `t.Parallel()` | Add as first line |
| `targetID := int64(123)` | `targetID := nextTestID()` |
| `assert.NoError()` for setup | `require.NoError(t, err, "setup: ...")` |
| Missing `defer cleanup()` | Add immediately after create |
| Assertion without message | Add descriptive message |

### D. Test Categories

**Required**: Create, Get, Update, Delete - success cases
**Boundary**: Nil fields, empty results, pagination
**Error**: Non-existent record, soft-deleted record
**Filtering**: Exclude by user, exclude expired, exclude deleted
**Batch**: Multiple operations, empty slice handling

### E. Table-Driven Pattern

```go
func Test{Repo}_Create_AllTypes(t *testing.T) {
	t.Parallel()
	db := database.SSWriter()
	
	testCases := []struct {
		name  string
		input SomeType
		want  SomeType
	}{
		{"case_1", input1, expected1},
		{"case_2", input2, expected2},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Do NOT call t.Parallel() in sub-tests
			targetID := next{Resource}TestID()
			record := &models.{Model}{TargetID: targetID}
			
			require.NoError(t, db.Create(record).Error)
			defer delete{Resource}(t, record.ID)
			
			assert.Equal(t, tc.want, record.Field)
		})
	}
}
```

---

## CHECKLIST

- [ ] Package doc with Test Case List
- [ ] All functions: `t.Parallel()` first, `Test{Repo}_{Op}_{Scenario}` naming
- [ ] Atomic ID generator with unique starting value
- [ ] `defer cleanup()` immediately after creation
- [ ] `require` for setup, `assert` for verification, all with messages
- [ ] `go test -v -parallel 4` passes
- [ ] `go test -v -parallel 8 -count 3` passes (race check)
