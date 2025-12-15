---
trigger: always_on
description: 
globs: 
---

# Go Project Coding Guidelines

## 🎯 CRITICAL REQUIREMENTS
- **NO compile errors** - Every change must compile
- **NO useless methods** - Only implement necessary functionality
- **Clean test artifacts** - Remove `alloc/` folder after testing

## 📄 YAML STANDARDS
- **Validate syntax**: `ruby -ryaml -e "YAML.safe_load(File.read('file.yaml'))"`
- Strings with `:` or `,` **MUST be quoted**

## 📋 LOGGING (kklogger)

**Format**: `package:Struct.Method#section!action`
- `section` and `action` are optional; `action` requires `section`
- **English only** in log messages

```go
kklogger.ErrorJ("auth:Handler.Post#load!cache", err.Error())
kklogger.InfoJ("service:Handler.method", "message")
```

## 🌐 HTTP HANDLERS

### Method Mapping
`Index`→GET(list), `Get`→GET(single), `Post`→POST, `Patch`→PATCH, `Delete`→DELETE

### Test Requirements
Before modifying `app/handlers/endpoints/`:
1. Search tests in `tests/units/handlers/` and `tests/e2e/`
2. Update tests if behavior changes
3. Run `go test -v ./tests/...` - all must pass

### Validation Pattern
```go
if len(req.Body().Bytes()) == 0 { return erresponse.InvalidRequestWrongBodyFormat }
if err := json.Unmarshal(req.Body().Bytes(), body); err != nil { return erresponse.InvalidRequestWrongBodyFormat }
if body.Field == "" { return erresponse.InvalidRequestCantBeEmptyOfName("field") }
```

### Response Rules
- **NEVER** return `success: true` - HTTP 200 already indicates success
- Return `nil` for success, or `resp.JsonResponse(&Data{})` with data

## 🛣️ ROUTE REGISTRATION
**Child routes inherit parent acceptances** - DO NOT duplicate.
```go
route.SetEndpoint("/api/v1/me", me.Handler, CheckAuth, acceptances.HEnsureUserExist)
route.SetEndpoint("/api/v1/me/contact", me.HandlerContact)  // Inherits parent acceptances
```

## 🔧 WORKER PATTERNS
- **MUST return error** on DB Save/Update/Delete failures
- Log errors before returning

## 🗄️ DATABASE (MySQL 8.0)

### Schema
- Location: `app/database/migrate/`, one change per file

### Model Rules
```go
type Model struct {
    ID        int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
    Name      string    `gorm:"column:name;not null" json:"name"`    // NOT NULL = no pointer
    Desc      *string   `gorm:"column:desc" json:"desc"`             // Nullable = pointer
    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}
```
- **MUST include `column`** in gorm tags
- Time: `2006-01-02 15:04:05` (string) or unix timestamp (int)

### Repository Rules
- Methods inheriting `DatabaseRepository` **MUST NOT return errors**
- Return `nil` or empty slice for no-data; log errors on Save/Delete

## 🎨 TEMPLATES & LOCALIZATION
- Templates: `./resources/template/default/` or `./resources/template/<lang>/`
- Translations: `./resources/translation/`

## 📝 ERROR RESPONSES
```go
erresponse.InvalidRequestWrongBodyFormat  // Bad JSON
erresponse.InvalidRequestExpired          // Token expired
erresponse.ServerError                    // Generic error
erresponse.InvalidRequestCantBeEmptyOfName("field")  // Validation
```

## 🏗️ PROJECT STRUCTURE
```
app/
├── conf/          # Configuration
├── connector/     # DB & external connectors
├── database/      # Schemas & migrations
├── handlers/      # HTTP handlers (acceptances/, endpoints/, minortasks/)
├── models/        # Data models
├── repositories/  # Data access layer
├── services/      # Business logic
└── worker/        # Background jobs

tests/
├── units/         # Unit tests (handlers/, helpers/, repositories/, services/)
└── e2e/           # End-to-end tests
```

## ✅ CHECKLIST
- [ ] `go build` succeeds
- [ ] All tests pass
- [ ] `alloc/` folders removed
- [ ] kklogger format correct (English only)
- [ ] GORM tags include `column`
- [ ] Repository methods don't return errors
