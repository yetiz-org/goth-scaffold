# app/repositories

## Pattern

Embed `*DatabaseDefaultRepository[K, *models.Model]` — do not reimplement Find/First/Save.

```go
type FooRepository struct {
	*DatabaseDefaultRepository[FooId, *models.Foo]
}

func NewFooRepository(db *gorm.DB) *FooRepository {
	return &FooRepository{
		DatabaseDefaultRepository: NewDatabaseDefaultRepository[FooId, *models.Foo](db),
	}
}
```

## Method Rules

- **Never return errors** — return `nil` / empty slice on failure; log via `kklogger.ErrorJ`.
- Use `FirstWhere` / `FindWhere` for all filtered queries — they apply gorm scope + in-memory filter + eager loading
  automatically.
- Pass `opts ...models.DatabaseQueryOption[*models.Foo]` as the last parameter on every method that returns model
  instances.

```go
func (r *FooRepository) FindByOwner(ownerID int64, opts ...models.DatabaseQueryOption[*models.Foo]) []*models.Foo {
	items, err := r.FindWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where("owner_id = ?", ownerID)
	}, opts...)
	if err != nil {
		kklogger.ErrorJ("repositories:FooRepository.FindByOwner#query!failed", err.Error())
		return []*models.Foo{}
	}
	return items
}
```

## SaveRetry

Use `SaveRetry` / `SaveRetryTx` for rows with optimistic-lock fields (`Version`).
It retries up to 5 times on duplicate-key conflicts with exponential backoff.

## Tx Suffix

Methods that participate in an external transaction take `tx *gorm.DB` as the first parameter and are named with a `Tx`
suffix.

## Built-in Methods (DatabaseDefaultRepository)

Do NOT reimplement these — they are provided by the embedded struct:

| Method                                    | Returns         | Notes                          |
|-------------------------------------------|-----------------|--------------------------------|
| `FindWhere(build, opts...)`               | `([]T, error)`  | 主要查詢入口，接受任意 GORM scope         |
| `FirstWhere(build, opts...)`              | `(T, error)`    | 主要單筆查詢                         |
| `Find(opts...)`                           | `([]T, error)`  | 無條件查全表                         |
| `First(opts...)`                          | `(T, error)`    | 無條件取第一筆                        |
| `Get(id K, opts...)`                      | `T`             | 不返回 error（符合 repo 慣例，失敗回 nil）  |
| `Fetch(id any, opts...)`                  | `(T, error)`    | 需要 error 時使用                   |
| `Save(entity)`                            | `error`         | INSERT OR UPDATE               |
| `SaveTx(tx, entity)`                      | `error`         | 在外部 Tx 中執行                     |
| `SaveRetry(entity)`                       | `error`         | 樂觀鎖（`Version` 欄位）重試，最多 5 次     |
| `SaveRetryTx(tx, entity)`                 | `error`         | 同上，在外部 Tx                      |
| `Delete(entity)`                          | `error`         |                                |
| `DeleteTx(tx, entity)`                    | `error`         |                                |
| `Upsert(entity, conditions)`              | `error`         | 依 `conditions` map 決定 WHERE 條件 |
| `UpsertTx(tx, entity, conditions)`        | `error`         |                                |
| `FirstOrCreate(entity, conditions)`       | `(bool, error)` | `bool=true` 表示新建               |
| `FirstOrCreateTx(tx, entity, conditions)` | `(bool, error)` |                                |

## Cassandra Repositories

For Cassandra models, embed `*CassandraDefaultRepository[T]` instead of `DatabaseDefaultRepository`:

```go
type FooRepository struct {
	*repositories.CassandraDefaultRepository[*models.Foo]
}

func NewFooRepository(session *gocql.Session) *FooRepository {
	return &FooRepository{
		CassandraDefaultRepository: repositories.NewCassandraDefaultRepository[*models.Foo](session),
	}
}
```

Built-in methods: `Save`, `Delete`, `UniqueCreate`, `QueryBuilder()`, `Session()`.
Unlike MySQL repositories, Cassandra methods **return `error` directly** — do not suppress them.
