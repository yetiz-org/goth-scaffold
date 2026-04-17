package repositories

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/components/dialect"
	"github.com/yetiz-org/goth-scaffold/app/models"
	"gorm.io/gorm"
)

// quoteSQLIdentifier delegates to the active dialect so SQL identifiers are
// quoted with backticks on MySQL and double quotes on PostgreSQL.
func quoteSQLIdentifier(ident string) string {
	return dialect.Current().QuoteIdent(ident)
}

func _ApplyWhereConditions(tx *gorm.DB, conditions map[string]any) *gorm.DB {
	q := tx
	for k, v := range conditions {
		col := quoteSQLIdentifier(k)
		if v == nil {
			q = q.Where(fmt.Sprintf("%s IS NULL", col))
			continue
		}

		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Pointer && rv.IsNil() {
			q = q.Where(fmt.Sprintf("%s IS NULL", col))
			continue
		}

		q = q.Where(fmt.Sprintf("%s = ?", col), v)
	}

	return q
}

type DatabaseDefaultRepository[K any, T models.Model] struct {
	db      *gorm.DB
	_DBFunc func() *gorm.DB
}

func (d *DatabaseDefaultRepository[K, T]) _ApplyGorm(db *gorm.DB, opts []models.DatabaseQueryOption[T]) *gorm.DB {
	for _, opt := range opts {
		db = opt.ApplyGorm(db)
	}

	return db
}

func (d *DatabaseDefaultRepository[K, T]) _ApplyEager(items []T, opts []models.DatabaseQueryOption[T]) {
	for _, opt := range opts {
		opt.ApplyEager(items)
	}
}

func (d *DatabaseDefaultRepository[K, T]) _ApplyFilter(items []T, opts []models.DatabaseQueryOption[T]) []T {
	for _, opt := range opts {
		items = opt.ApplyFilter(items)
	}

	return items
}

// FindWhere executes a query built by the caller-provided function.
func (d *DatabaseDefaultRepository[K, T]) FindWhere(
	build func(db *gorm.DB) *gorm.DB,
	opts ...models.DatabaseQueryOption[T],
) ([]T, error) {
	var items []T
	db := d._ApplyGorm(build(d.DB()), opts)
	if err := db.Find(&items).Error; err != nil {
		kklogger.ErrorJ("repo:DatabaseDefaultRepository.FindWhere#query!db_error", err.Error())
		return nil, err
	}

	items = d._ApplyFilter(items, opts)
	d._ApplyEager(items, opts)
	return items, nil
}

// FirstWhere executes a query built by the caller-provided function and returns the first match.
// Returns (zero, nil) when no record is found.
func (d *DatabaseDefaultRepository[K, T]) FirstWhere(
	build func(db *gorm.DB) *gorm.DB,
	opts ...models.DatabaseQueryOption[T],
) (T, error) {
	entityType := reflect.TypeFor[T]().Elem()
	entity := reflect.New(entityType).Interface().(T)

	db := d._ApplyGorm(build(d.DB()), opts)
	if err := db.First(entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return *new(T), nil
		}

		kklogger.ErrorJ("repo:DatabaseDefaultRepository.FirstWhere#query!db_error", err.Error())
		return *new(T), err
	}

	filtered := d._ApplyFilter([]T{entity}, opts)
	if len(filtered) == 0 {
		return *new(T), nil
	}

	d._ApplyEager(filtered, opts)
	return filtered[0], nil
}

// First retrieves the first entity matching the given options.
// Returns (zero, nil) when no record is found.
func (d *DatabaseDefaultRepository[K, T]) First(opts ...models.DatabaseQueryOption[T]) (T, error) {
	entityType := reflect.TypeFor[T]().Elem()
	entity := reflect.New(entityType).Interface().(T)

	db := d._ApplyGorm(d.DB(), opts)
	if err := db.First(entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return *new(T), nil
		}

		kklogger.ErrorJ("repo:DatabaseDefaultRepository.First#query!db_error", err.Error())
		return *new(T), err
	}

	filtered := d._ApplyFilter([]T{entity}, opts)
	if len(filtered) == 0 {
		return *new(T), nil
	}

	d._ApplyEager(filtered, opts)
	return filtered[0], nil
}

// Find retrieves all entities matching the given options.
func (d *DatabaseDefaultRepository[K, T]) Find(opts ...models.DatabaseQueryOption[T]) ([]T, error) {
	var items []T
	db := d._ApplyGorm(d.DB(), opts)
	if err := db.Find(&items).Error; err != nil {
		kklogger.ErrorJ("repo:DatabaseDefaultRepository.Find#query!db_error", err.Error())
		return nil, err
	}

	items = d._ApplyFilter(items, opts)
	d._ApplyEager(items, opts)

	return items, nil
}

type TxBeginFunc func() (*gorm.DB, error)

type TransactionFunc func(tx *gorm.DB) error

func NewDatabaseDefaultRepository[K any, T models.Model](db *gorm.DB) *DatabaseDefaultRepository[K, T] {
	return &DatabaseDefaultRepository[K, T]{
		db: db,
	}
}

func NewDatabaseDefaultRepositoryF[K any, T models.Model](dbFunc func() *gorm.DB) *DatabaseDefaultRepository[K, T] {
	return &DatabaseDefaultRepository[K, T]{
		_DBFunc: dbFunc,
	}
}

func (d *DatabaseDefaultRepository[K, T]) TableName() string {
	model := *new(T)
	return model.TableName()
}

func (d *DatabaseDefaultRepository[K, T]) Save(entity T) error {
	if m, ok := any(entity).(models.ModelSavePreHook); ok {
		if err := m.PreSave(context.Background()); err != nil {
			return err
		}
	}

	err := d.DB().Save(entity).Error
	if err == nil {
		if m, ok := any(entity).(models.ModelSavePostHook); ok {
			if err := m.PostSave(context.Background()); err != nil {
				return err
			}
		}
	}

	return err
}

func (d *DatabaseDefaultRepository[K, T]) SaveTx(tx *gorm.DB, entity T) error {
	if m, ok := any(entity).(models.ModelSavePreHook); ok {
		if err := m.PreSave(context.Background()); err != nil {
			return err
		}
	}

	err := tx.Save(entity).Error
	if err == nil {
		if m, ok := any(entity).(models.ModelSavePostHook); ok {
			if err := m.PostSave(context.Background()); err != nil {
				return err
			}
		}
	}

	return err
}

func (d *DatabaseDefaultRepository[K, T]) SaveRetry(entity T) error {
	return d.saveWithRetry(d.DB(), entity)
}

func (d *DatabaseDefaultRepository[K, T]) SaveRetryTx(tx *gorm.DB, entity T) error {
	return d.saveWithRetry(tx, entity)
}

func WithTransactionRetry(maxRetries int, begin TxBeginFunc, fn TransactionFunc) error {
	if maxRetries <= 0 {
		maxRetries = 1
	}

	backoffs := transactionRetryBackoffs(maxRetries)
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		tx, err := begin()
		if err != nil {
			lastErr = err
			if !isRetryableError(err) || attempt == maxRetries {
				return err
			}

			kklogger.WarnJ("repo:WithTransactionRetry#retry!begin_error", map[string]any{
				"attempt": attempt,
				"error":   err.Error(),
			})
			time.Sleep(backoffs[attempt-1])
			continue
		}

		err = withTransactionRecovery(func() error {
			return fn(tx)
		}, func() {
			if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
				kklogger.ErrorJ("repo:WithTransactionRetry#rollback!rollback_error", rollbackErr.Error())
			}
		})
		if err != nil {
			lastErr = err
			if !isRetryableError(err) || attempt == maxRetries {
				return err
			}

			kklogger.WarnJ("repo:WithTransactionRetry#retry!execute_error", map[string]any{
				"attempt": attempt,
				"error":   err.Error(),
			})
			time.Sleep(backoffs[attempt-1])
			continue
		}

		if commitErr := tx.Commit().Error; commitErr != nil {
			lastErr = commitErr
			if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
				kklogger.ErrorJ("repo:WithTransactionRetry#commit!rollback_error", rollbackErr.Error())
			}

			if !isRetryableError(commitErr) || attempt == maxRetries {
				return commitErr
			}

			kklogger.WarnJ("repo:WithTransactionRetry#retry!commit_error", map[string]any{
				"attempt": attempt,
				"error":   commitErr.Error(),
			})
			time.Sleep(backoffs[attempt-1])
			continue
		}

		return nil
	}

	return lastErr
}

func (d *DatabaseDefaultRepository[K, T]) saveWithRetry(tx *gorm.DB, entity T) error {
	backoffs := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		500 * time.Millisecond,
	}

	var lastErr error
	for attempt := 0; attempt <= len(backoffs); attempt++ {
		if m, ok := any(entity).(models.ModelSavePreHook); ok {
			if err := m.PreSave(context.Background()); err != nil {
				return err
			}
		}

		err := tx.Save(entity).Error
		if err == nil {
			if m, ok := any(entity).(models.ModelSavePostHook); ok {
				if err := m.PostSave(context.Background()); err != nil {
					return err
				}
			}

			return nil
		}

		lastErr = err
		if !isRetryableError(err) || attempt == len(backoffs) {
			return err
		}

		time.Sleep(backoffs[attempt])
	}

	return lastErr
}

func withTransactionRecovery(fn func() error, rollback func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			rollback()
			err = fmt.Errorf("panic occurred")
			kklogger.ErrorJ("repo:withTransactionRecovery#panic!recover", "panic occurred")
		}
	}()

	if err = fn(); err != nil {
		rollback()
		return err
	}

	return nil
}

func transactionRetryBackoffs(maxRetries int) []time.Duration {
	baseBackoffs := []time.Duration{
		50 * time.Millisecond,
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
	}

	if maxRetries <= len(baseBackoffs) {
		return baseBackoffs[:maxRetries]
	}

	backoffs := make([]time.Duration, 0, maxRetries)
	backoffs = append(backoffs, baseBackoffs...)
	last := baseBackoffs[len(baseBackoffs)-1]
	for len(backoffs) < maxRetries {
		last = last * 2
		backoffs = append(backoffs, last)
	}

	return backoffs
}

func (d *DatabaseDefaultRepository[K, T]) Delete(entity T) error {
	if m, ok := any(entity).(models.ModelDeletePreHook); ok {
		if err := m.PreDelete(context.Background()); err != nil {
			return err
		}
	}

	err := d.DB().Delete(entity).Error
	if err == nil {
		if m, ok := any(entity).(models.ModelDeletePostHook); ok {
			if err := m.PostDelete(context.Background()); err != nil {
				return err
			}
		}
	}

	return err
}

func (d *DatabaseDefaultRepository[K, T]) DeleteTx(tx *gorm.DB, entity T) error {
	if m, ok := any(entity).(models.ModelDeletePreHook); ok {
		if err := m.PreDelete(context.Background()); err != nil {
			return err
		}
	}

	err := tx.Delete(entity).Error
	if err == nil {
		if m, ok := any(entity).(models.ModelDeletePostHook); ok {
			if err := m.PostDelete(context.Background()); err != nil {
				return err
			}
		}
	}

	return err
}

func (d *DatabaseDefaultRepository[K, T]) DB() *gorm.DB {
	if d._DBFunc != nil {
		return d._DBFunc()
	}

	return d.db
}

// DefaultLimit returns the shared default page size for repositories.
func (d *DatabaseDefaultRepository[K, T]) DefaultLimit() int {
	return 50
}

// Fetch retrieves a single entity by primary key.
// Returns (zero, nil) when the record is not found.
func (d *DatabaseDefaultRepository[K, T]) Fetch(id any, opts ...models.DatabaseQueryOption[T]) (model T, err error) {
	entityType := reflect.TypeFor[T]().Elem()
	entity := reflect.New(entityType).Interface().(T)

	db := d._ApplyGorm(d.DB(), opts)

	var er error
	if strID, ok := id.(string); ok {
		er = db.Where("id = ?", strID).First(entity).Error
	} else {
		er = db.First(entity, id).Error
	}

	if er != nil {
		if errors.Is(er, gorm.ErrRecordNotFound) {
			return *new(T), nil
		}

		kklogger.ErrorJ("repo:DatabaseDefaultRepository.Fetch#query!db_error", er.Error())
		return *new(T), er
	}

	filtered := d._ApplyFilter([]T{entity}, opts)
	if len(filtered) == 0 {
		return *new(T), nil
	}

	d._ApplyEager(filtered, opts)
	return filtered[0], nil
}

// Get retrieves a single entity by primary key.
// Returns zero value when the record is not found or any error occurs.
func (d *DatabaseDefaultRepository[K, T]) Get(id K, opts ...models.DatabaseQueryOption[T]) T {
	model, _ := d.Fetch(id, opts...)
	return model
}

// Upsert updates or creates an entity based on the given conditions.
func (d *DatabaseDefaultRepository[K, T]) Upsert(entity T, conditions map[string]any) error {
	return d.UpsertTx(d.DB(), entity, conditions)
}

// UpsertTx updates or creates an entity within a transaction.
func (d *DatabaseDefaultRepository[K, T]) UpsertTx(tx *gorm.DB, entity T, conditions map[string]any) error {
	if m, ok := any(entity).(models.ModelSavePreHook); ok {
		if err := m.PreSave(context.Background()); err != nil {
			return err
		}
	}

	var existing T
	err := _ApplyWhereConditions(tx, conditions).First(&existing).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		kklogger.ErrorJ("repo:DatabaseDefaultRepository.UpsertTx#query!db_error", err.Error())
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		createErr := tx.Create(entity).Error
		if createErr != nil {
			if dialect.Current().IsDuplicateKeyErr(createErr) {
				requeryErr := _ApplyWhereConditions(tx, conditions).First(&existing).Error
				if requeryErr != nil {
					kklogger.ErrorJ("repo:DatabaseDefaultRepository.UpsertTx#query_after_dup!db_error", requeryErr.Error())
					return requeryErr
				}

				copyPrimaryKey(existing, entity)

				saveErr := tx.Omit("created_at").Save(entity).Error
				if saveErr != nil {
					kklogger.ErrorJ("repo:DatabaseDefaultRepository.UpsertTx#save_after_dup!db_error", saveErr.Error())
					return saveErr
				}
			} else {
				kklogger.ErrorJ("repo:DatabaseDefaultRepository.UpsertTx#create!db_error", createErr.Error())
				return createErr
			}
		}
	} else {
		copyPrimaryKey(existing, entity)

		saveErr := tx.Omit("created_at").Save(entity).Error
		if saveErr != nil {
			kklogger.ErrorJ("repo:DatabaseDefaultRepository.UpsertTx#save!db_error", saveErr.Error())
			return saveErr
		}
	}

	if m, ok := any(entity).(models.ModelSavePostHook); ok {
		if err := m.PostSave(context.Background()); err != nil {
			return err
		}
	}

	return nil
}

func (d *DatabaseDefaultRepository[K, T]) FirstOrCreate(entity T, conditions map[string]any) (bool, error) {
	return d.FirstOrCreateTx(d.DB(), entity, conditions)
}

func (d *DatabaseDefaultRepository[K, T]) FirstOrCreateTx(tx *gorm.DB, entity T, conditions map[string]any) (bool, error) {
	err := _ApplyWhereConditions(tx, conditions).First(entity).Error
	if err == nil {
		return false, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		kklogger.ErrorJ("repo:DatabaseDefaultRepository.FirstOrCreateTx#query!db_error", err.Error())
		return false, err
	}

	if m, ok := any(entity).(models.ModelSavePreHook); ok {
		if err := m.PreSave(context.Background()); err != nil {
			return false, err
		}
	}

	createErr := tx.Create(entity).Error
	if createErr != nil {
		if dialect.Current().IsDuplicateKeyErr(createErr) {
			requeryErr := _ApplyWhereConditions(tx, conditions).First(entity).Error
			if requeryErr == nil {
				return false, nil
			}

			if !errors.Is(requeryErr, gorm.ErrRecordNotFound) {
				kklogger.ErrorJ("repo:DatabaseDefaultRepository.FirstOrCreateTx#requery!db_error", requeryErr.Error())
			}
			return false, createErr
		}

		kklogger.ErrorJ("repo:DatabaseDefaultRepository.FirstOrCreateTx#create!db_error", createErr.Error())
		return false, createErr
	}

	if m, ok := any(entity).(models.ModelSavePostHook); ok {
		if err := m.PostSave(context.Background()); err != nil {
			return true, err
		}
	}

	return true, nil
}

// IsLockNoWaitError reports whether err is a lock-not-available error on the
// active dialect (MySQL 3572 / Postgres 55P03).
func IsLockNoWaitError(err error) bool {
	return dialect.Current().IsLockNoWaitErr(err)
}

func isRetryableError(err error) bool {
	return dialect.Current().IsRetryableErr(err)
}

// copyPrimaryKey copies the primary key value from src to dst.
func copyPrimaryKey(src, dst any) {
	srcVal := reflect.ValueOf(src).Elem()
	dstVal := reflect.ValueOf(dst).Elem()
	srcType := srcVal.Type()

	var srcField, dstField reflect.Value
	var found bool

	for i := 0; i < srcType.NumField(); i++ {
		field := srcType.Field(i)
		gormTag := field.Tag.Get("gorm")
		if strings.Contains(strings.ToLower(gormTag), "primarykey") {
			srcField = srcVal.Field(i)
			dstField = dstVal.FieldByName(field.Name)
			found = true
			break
		}
	}

	if !found {
		for _, name := range []string{"ID", "Id", "id"} {
			srcField = srcVal.FieldByName(name)
			dstField = dstVal.FieldByName(name)
			if srcField.IsValid() && dstField.IsValid() {
				found = true
				break
			}
		}
	}

	if !found || !srcField.IsValid() || !dstField.IsValid() || !dstField.CanSet() {
		return
	}

	switch srcField.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		dstField.SetInt(srcField.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		dstField.SetUint(srcField.Uint())
	case reflect.String:
		dstField.SetString(srcField.String())
	}
}
