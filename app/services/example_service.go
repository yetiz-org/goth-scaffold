package services

// ExampleService demonstrates the service layer conventions.
//
// Three usage patterns are shown:
//
//  1. Singleton (global shared instance via ExampleServiceInstance)
//  2. Direct construction: &ExampleService{}
//  3. Struct embedding: embed ExampleService inside another struct
//
// The DEP injection pattern (_Dependency / _Deps()) keeps the service testable:
// tests can inject a mock repo via _Dependency; production falls back to a lazily
// initialised package-level default via _Deps().

import (
	"sync"

	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/connector/database"
	"github.com/yetiz-org/goth-scaffold/app/models"
	"github.com/yetiz-org/goth-scaffold/app/repositories"
	"gorm.io/gorm"
)

// =============================================================================
// Dependency container
// =============================================================================

// _ExampleDeps holds injected repositories for ExampleService.
// Replace any field with a mock implementation during testing.
type _ExampleDeps struct {
	SiteSettingRepo models.SiteSettingRepository
}

// _defaultExampleDeps is the process-wide default dependency set.
// It is initialised once on first production use via _defaultExampleDepsOnce.
// Tests bypass this by setting _Dependency on the ExampleService struct directly.
var (
	_defaultExampleDeps     *_ExampleDeps
	_defaultExampleDepsOnce sync.Once
)

// =============================================================================
// Singleton (optional — convenient for handlers and daemons)
// =============================================================================

var (
	_exampleService     *ExampleService
	_exampleServiceOnce sync.Once
)

// ExampleServiceInstance returns the process-wide singleton.
func ExampleServiceInstance() *ExampleService {
	_exampleServiceOnce.Do(func() {
		_exampleService = &ExampleService{}
	})

	return _exampleService
}

// =============================================================================
// Service struct
// =============================================================================

// ExampleService handles business logic for the Example domain.
//
// Direct construction (handler, no singleton required):
//
//	svc := &ExampleService{}
//
// Inject test doubles:
//
//	svc := &ExampleService{_Dependency: &_ExampleDeps{SiteSettingRepo: mockRepo}}
//
// Struct embedding (embed directly into another struct):
//
//	type FooHandler struct{ ExampleService }
//	h.ListSettings()
type ExampleService struct {
	// _Dependency is nil in production.
	// Set this field in tests to inject mocks without modifying the service.
	_Dependency *_ExampleDeps
}

// _Deps returns the active dependency set.
// In production (_Dependency == nil) the package-level default is lazily initialised once
// and reused for the lifetime of the process.
// In tests, set _Dependency on the struct to inject mocks without touching the package default.
func (s *ExampleService) _Deps() *_ExampleDeps {
	if s != nil && s._Dependency != nil {
		return s._Dependency
	}

	_defaultExampleDepsOnce.Do(func() {
		_defaultExampleDeps = &_ExampleDeps{
			SiteSettingRepo: repositories.NewSiteSettingRepository(database.Writer()),
		}
	})

	return _defaultExampleDeps
}

// =============================================================================
// Public methods
// =============================================================================

// ListSettings returns all site settings, ordered by category/key.
// Returns an empty slice on failure; errors are logged internally.
func (s *ExampleService) ListSettings() []*models.SiteSetting {
	kklogger.InfoJ("services:ExampleService.ListSettings#fetch!start", nil)

	return s._Deps().SiteSettingRepo.List()
}

// FindSettings returns site settings that satisfy the given query options.
// Returns nil on failure; errors are logged internally.
func (s *ExampleService) FindSettings(opts ...models.DatabaseQueryOption[*models.SiteSetting]) []*models.SiteSetting {
	kklogger.InfoJ("services:ExampleService.FindSettings#fetch!start", nil)

	results, err := s._Deps().SiteSettingRepo.Find(opts...)
	if err != nil {
		kklogger.ErrorJ("services:ExampleService.FindSettings#fetch!db_error", err.Error())
		return nil
	}

	return results
}

// ListSettingsWithTags returns all settings with their Tags pre-loaded in batch.
// Uses EagerAll to fetch Settings + Tags in 2 queries total (N+1-free).
//
// Example:
//
//	svc := &ExampleService{}
//	for _, s := range svc.ListSettingsWithTags() {
//	    fmt.Println(s.Category, s.Key, s.Tags())
//	}
func (s *ExampleService) ListSettingsWithTags() []*models.SiteSetting {
	kklogger.InfoJ("services:ExampleService.ListSettingsWithTags#fetch!start", nil)

	results, err := s._Deps().SiteSettingRepo.Find(models.EagerAll[*models.SiteSetting]())
	if err != nil {
		kklogger.ErrorJ("services:ExampleService.ListSettingsWithTags#fetch!db_error", err.Error())
		return nil
	}

	return results
}

// UpdateTx updates a site setting within the provided transaction.
// The "Tx" suffix signals that this method participates in an external transaction.
//
// Returns (rowsAffected, hasError).
func (s *ExampleService) UpdateTx(tx *gorm.DB, setting *models.SiteSetting) (rowsAffected int64, hasError bool) {
	kklogger.InfoJ("services:ExampleService.UpdateTx#update!start", map[string]any{"id": setting.ID})

	result := tx.Model(setting).Updates(setting)
	if result.Error != nil {
		kklogger.ErrorJ("services:ExampleService.UpdateTx#update!db_error", result.Error.Error())
		return 0, true
	}

	return result.RowsAffected, false
}
