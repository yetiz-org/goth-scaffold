# app/models

## ID Type + Encryption (required for every model with integer PK)

```go
type FooId crypto.KeyId
const KeyTypeFooId crypto.KeyType = "foo_id"

func (k FooId) EncryptId() (string, error)            { return crypto.EncryptKeyId(KeyTypeFooId, crypto.KeyId(k)) }
func (k FooId) EncryptedId() string                   { return crypto.EncryptedKeyId(KeyTypeFooId, crypto.KeyId(k)) }
func (k FooId) UInt64() uint64                        { return uint64(k) }
func (k FooId) DecryptId(enc string) (FooId, error)   { v, err := crypto.DecryptKeyId[crypto.KeyId](KeyTypeFooId, enc); return FooId(v), err }
```

Place this block **before** the model struct. All five methods required. Never inline the crypto logic.

FK fields referencing another model must use that model's `<OtherModel>Id` type, never bare `int`/`uint`.

## GORM String Tags

Always include `size:<n>` or `type:<t>` matching the migration:

| DB type    | gorm tag       |
|------------|----------------|
| VARCHAR(n) | `size:n`       |
| CHAR(n)    | `type:char(n)` |
| TEXT       | `size:65535`   |
| JSON       | `type:json`    |

## Lazy Associations

Use helpers — no manual DB calls in getters:

```go
// struct field (private)
_Items []*Item `gorm:"foreignKey:FooID"`

// SetCache (required for batch preload)
func (m *Foo) SetCacheItems(v []*Item) { if m == nil { return }; m._Items = v }

// Getter
func (m *Foo) Items() []*Item {
	if m == nil { return nil }
	return LazyHasMany[Item](m, &m._Items)
}
```

| Helper             | Relation   | FK location |
|--------------------|------------|-------------|
| `LazyBelongsTo[T]` | belongs-to | this table  |
| `LazyHasMany[T]`   | has-many   | child table |
| `LazyHasOne[T]`    | has-one    | child table |

## Repository Interface Rule

Methods returning model instances must accept `opts ...DatabaseQueryOption[*ModelType]` as the final parameter (except
Tx methods or scalar returns).

## DatabaseQueryOption Constructors

Pass these as `opts` to any repository method that accepts `...DatabaseQueryOption[T]`:

| Constructor                       | Purpose                                   | Example                                                                          |
|-----------------------------------|-------------------------------------------|----------------------------------------------------------------------------------|
| `GormOpt[T](fn)`                  | 任意 GORM scope                             | `GormOpt[*Foo](func(db *gorm.DB) *gorm.DB { return db.Where("status = ?", 1) })` |
| `SelectOpt[T](cols...)`           | 限制回傳欄位（`id` 自動包含）                         | `SelectOpt[*Foo]("name", "status")`                                              |
| `PaginationOpt[T](offset, limit)` | LIMIT/OFFSET 分頁；`limit=0` 時自動補 `MaxInt32` | `PaginationOpt[*Foo](20, 10)`                                                    |
| `LimitOpt[T](limit)`              | 純 LIMIT，不含 OFFSET                         | `LimitOpt[*Foo](5)`                                                              |
| `EagerAll[T]()`                   | 自動 preload 所有 lazy associations           | `EagerAll[*Foo]()`                                                               |
| `QueryOpt[T](gormFn, filterFn)`   | queryfilter 元件用的複合 option                 | —                                                                                |

Multiple opts compose freely:

```go
repo.FindByOwner(id, EagerAll[*Foo](), PaginationOpt[*Foo](0, 20))
```
