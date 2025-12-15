package repositories

import (
	"context"

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
