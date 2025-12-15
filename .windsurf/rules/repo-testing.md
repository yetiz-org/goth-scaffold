---
trigger: model_decision
description: Repository integration tests guidelines with parallel execution
globs: 
---

# Repository Integration Tests Guidelines

## Core Principles (MANDATORY)

1. **Parallel Execution**: All tests MUST use `t.Parallel()` - no exceptions
2. **Test Isolation**: Each test uses unique IDs via atomic counter
3. **Auto Cleanup**: Use `defer` immediately after creation
4. **Language**: English only

---

## Naming & Structure

### Test Function: `Test{Repository}_{Operation}_{Scenario}`
```go
TestNoticeRepository_Create_AllFields
TestNoticeRepository_Get_NonExistentNotice
TestEmailLogRepository_FindByMessageID_EmptyString
```

### Helper Functions
```go
func deleteNotice(t *testing.T, id int64)      // cleanup
func nextNoticeTestID() int64                   // unique ID generator
```

---

## File Template

```go
/*
Package repositories - {Repository} Integration Tests

Features:
  - {Feature 1}
  - {Feature 2}

Test Case List:
  - Test{Repo}_{Op}_{Scenario}: {Description}
*/
package repositories

import (
    "sync/atomic"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// =============================================================================
// Test ID Generator - Unique IDs for parallel isolation
// =============================================================================

var noticeTestIDCounter int64 = 9999800  // Each file uses different starting value

func nextNoticeTestID() int64 {
    return atomic.AddInt64(&noticeTestIDCounter, 1)
}

// =============================================================================
// Cleanup Helpers
// =============================================================================

func deleteNotice(t *testing.T, id int64) {
    t.Helper()
    db := database.SSWriter()
    db.Unscoped().Delete(&models.Notice{}, id)
}

func deleteNotices(t *testing.T, ids []int64) {
    t.Helper()
    if len(ids) == 0 {
        return
    }
    db := database.SSWriter()
    db.Unscoped().Where("id IN ?", ids).Delete(&models.Notice{})
}
```

---

## Test Function Template

```go
// TestNoticeRepository_Get_ExistingNotice tests retrieval by ID
//
// Test scenario:
//   - Create a notice in database
//   - Call repo.Get() with the notice ID
//   - Verify the returned notice matches
func TestNoticeRepository_Get_ExistingNotice(t *testing.T) {
    t.Parallel()  // ✅ MANDATORY first line
    
    // ==================== Arrange ====================
    db := database.SSWriter()
    repo := repositories.NewNoticeRepository(db)
    targetID := nextNoticeTestID()  // ✅ Unique ID
    
    notice := &models.Notice{
        Category: "test_category",
        Title:    "Test Notice",
        TargetID: targetID,
    }
    require.NoError(t, db.Create(notice).Error, "setup: must create notice")
    defer deleteNotice(t, notice.ID)  // ✅ Immediate cleanup
    
    // ==================== Act ====================
    result := repo.Get(notice.ID)
    
    // ==================== Assert ====================
    require.NotNil(t, result, "should return existing notice")
    assert.Equal(t, notice.ID, result.ID, "ID should match")
}
```

---

## Assertion Rules

| Function | Use Case | On Failure |
|----------|----------|------------|
| `require` | Setup, fatal conditions | Stop test |
| `assert` | Verification | Continue test |

**All assertions MUST include descriptive message.**

---

## Test Case Categories

Design tests covering these categories:

1. **Happy Path**: Basic CRUD operations
2. **Boundary**: Empty values, nil, pagination limits
3. **Error Cases**: Non-existent records, soft-deleted records
4. **Filtering**: Exclude other users, expired, soft-deleted
5. **Batch Operations**: Empty slice handling

---

## ❌ Common Mistakes

```go
// ❌ Missing t.Parallel()
func TestSomething(t *testing.T) {
    // wrong
}

// ❌ Hardcoded ID
notice := &models.Notice{TargetID: 123}

// ❌ Missing cleanup
result := createTestData()
// no defer

// ❌ assert for setup (should use require)
assert.NoError(t, err)

// ❌ No error message
require.NoError(t, err)
```

---

## Verification Commands

```bash
# Run all repo tests
go test -v ./tests/units/repositories/...

# Run specific test
go test -v -run TestNoticeRepository_Get ./tests/units/repositories/...

# Parallel execution test
go test -v -parallel 8 -count 3 ./tests/units/repositories/...
```
