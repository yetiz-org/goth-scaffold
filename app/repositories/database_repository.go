package repositories

import (
	"context"
	"reflect"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/models"
	"gorm.io/gorm"
)

type DatabaseDefaultRepository[T models.Model] struct {
	db *gorm.DB
}

func NewDatabaseDefaultRepository[T models.Model](db *gorm.DB) *DatabaseDefaultRepository[T] {
	return &DatabaseDefaultRepository[T]{
		db: db,
	}
}

func (d *DatabaseDefaultRepository[T]) TableName() string {
	model := *new(T)
	return model.TableName()
}

func (d *DatabaseDefaultRepository[T]) Save(entity T) error {
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

func (d *DatabaseDefaultRepository[T]) SaveTx(tx *gorm.DB, entity T) error {
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

func (d *DatabaseDefaultRepository[T]) Delete(entity T) error {
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

func (d *DatabaseDefaultRepository[T]) DeleteTx(tx *gorm.DB, entity T) error {
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

func (d *DatabaseDefaultRepository[T]) DB() *gorm.DB {
	return d.db
}

// DefaultLimit returns the shared default page size for repositories
func (d *DatabaseDefaultRepository[T]) DefaultLimit() int {
	return 50
}

// Upsert updates or creates an entity based on the given conditions.
// Similar to Laravel's updateOrCreate: first parameter is unique key conditions,
// entity contains all values to save.
func (d *DatabaseDefaultRepository[T]) Upsert(entity T, conditions map[string]any) error {
	return d.UpsertTx(d.DB(), entity, conditions)
}

// UpsertTx updates or creates an entity within a transaction.
// Note: The tx parameter may contain Omit/Select clauses that should only affect
// Create/Save operations, not First queries. We use a clean session for queries.
func (d *DatabaseDefaultRepository[T]) UpsertTx(tx *gorm.DB, entity T, conditions map[string]any) error {
	if m, ok := any(entity).(models.ModelSavePreHook); ok {
		if err := m.PreSave(context.Background()); err != nil {
			return err
		}
	}

	var existing T
	err := tx.Where(conditions).First(&existing).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		kklogger.ErrorJ("repo:DatabaseDefaultRepository.UpsertTx#query!db_error", err.Error())
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		createErr := tx.Create(entity).Error
		if createErr != nil {
			if isMySQLDuplicateKeyError(createErr) {
				requeryErr := tx.Where(conditions).First(&existing).Error
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

func (d *DatabaseDefaultRepository[T]) FirstOrCreate(entity T, conditions map[string]any) (bool, error) {
	return d.FirstOrCreateTx(d.DB(), entity, conditions)
}

func (d *DatabaseDefaultRepository[T]) FirstOrCreateTx(tx *gorm.DB, entity T, conditions map[string]any) (bool, error) {
	err := tx.Where(conditions).First(entity).Error
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
		if isMySQLDuplicateKeyError(createErr) {
			requeryErr := tx.Where(conditions).First(entity).Error
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

func isMySQLDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}

	return false
}

// copyPrimaryKey copies the primary key value from src to dst.
// Supports multiple ID field naming conventions (ID, Id, id) and detects via gorm primaryKey tag.
func copyPrimaryKey(src, dst any) {
	srcVal := reflect.ValueOf(src).Elem()
	dstVal := reflect.ValueOf(dst).Elem()
	srcType := srcVal.Type()

	// Try to find primary key field by gorm tag first, then by common naming conventions
	var srcField, dstField reflect.Value
	var found bool

	// First: look for gorm primaryKey tag
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

	// Fallback: try common ID field names
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
