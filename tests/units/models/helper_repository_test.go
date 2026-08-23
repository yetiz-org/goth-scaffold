/*
Tests that the public GettableRepository contract remains source-compatible with get-only implementations.
The generic DatabaseDefaultRepository may expose additive batch methods without expanding this interface.
*/

package models_test

import "github.com/yetiz-org/goth-scaffold/app/models"

type gettableCompatibilityModel struct{}

func (*gettableCompatibilityModel) TableName() (name string) {
	return "gettable_compatibility_models"
}

type getOnlyRepository struct{}

func (*getOnlyRepository) Get(id uint64, opts ...models.DatabaseQueryOption[*gettableCompatibilityModel]) (model *gettableCompatibilityModel) {
	return nil
}

var _ models.GettableRepository[uint64, *gettableCompatibilityModel] = (*getOnlyRepository)(nil)
