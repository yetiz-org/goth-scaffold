# app/services

## Singleton Pattern

```go
var (
	_fooService     *FooService
	_fooServiceOnce sync.Once
)

func FooServiceInstance() *FooService {
	_fooServiceOnce.Do(func() { _fooService = &FooService{} })
	return _fooService
}
```

## Method Conventions

- Every public method has a doc comment summarising: purpose, parameter constraints, return values.
- Methods that participate in an external transaction take `tx *gorm.DB` as the first parameter and are named with a
  `Tx` suffix.
- Return types use named variables: `(result *Foo)`, `(rowsAffected int64, hasError bool)`.

## Logging

```
kklogger.InfoJ("services:FooService.DoThing#section!action", payload)
```

Format: `services:StructName.MethodName#section!action` — English only.

## Dependencies

Hold repositories and sibling services as private fields; initialise lazily.

Private accessor methods use `_PascalCase` (per project convention). Fields use `_camelCase`
to avoid naming collisions with the accessor method.

```go
type FooService struct {
	_fooRepo repositories.FooRepository  // lowercase field — avoids clash with _FooRepo() method
}

// _FooRepo returns the lazily-initialised repository (private method, _PascalCase).
func (s *FooService) _FooRepo() repositories.FooRepository {
	if s._fooRepo == nil {
		s._fooRepo = repositories.NewFooRepository(database.Writer())
	}
	return s._fooRepo
}
```

For testable services with injected mocks, prefer the **DEP injection pattern**
(see `app/services/example_service.go`):

```go
type _FooDeps struct {
	FooRepo repositories.FooRepository
}

type FooService struct {
	_Dependency *_FooDeps
}

func (s *FooService) _Deps() *_FooDeps {
	if s != nil && s._Dependency != nil {
		return s._Dependency
	}
	return &_FooDeps{FooRepo: repositories.NewFooRepository(database.Writer())}
}
```

### Testing with DEP injection

In unit tests, inject mocks via `_Dependency`:

```go
func TestFooService_DoThing(t *testing.T) {
	mockRepo := &mockFooRepo{...}
	svc := &FooService{
		_Dependency: &_FooDeps{FooRepo: mockRepo},
	}
	// test svc.DoThing(...)
}
```

Services using the lazy-init pattern (no `_Dependency` field) are harder to mock — prefer DEP injection for any service
that requires unit testing.
