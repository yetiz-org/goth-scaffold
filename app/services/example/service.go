// Package example demonstrates the scaffold service-layer package pattern.
package example

import (
	"github.com/gocql/gocql"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/connector/database"
	"github.com/yetiz-org/goth-scaffold/app/connector/keyspaces"
	"github.com/yetiz-org/goth-scaffold/app/models"
	"github.com/yetiz-org/goth-scaffold/app/repositories"
	"gorm.io/gorm"
)

// Service coordinates the scaffold's example SQL and Cassandra repositories.
type Service struct{}

// _ServiceDeps groups repositories used by Service.
type _ServiceDeps struct {
	SiteSettingRepository    models.SiteSettingRepository
	SiteSettingTagRepository models.SiteSettingTagRepository
	MaintenanceLogRepository models.MaintenanceLogRepository
}

var _Deps = _ServiceDeps{
	SiteSettingRepository:    repositories.NewSiteSettingRepositoryF(database.Writer),
	SiteSettingTagRepository: repositories.NewSiteSettingTagRepositoryF(database.Writer),
	MaintenanceLogRepository: repositories.NewMaintenanceLogRepositoryF(_KeyspacesWriterSession),
}

func _KeyspacesWriterSession() (session *gocql.Session) {
	if !keyspaces.Enabled() {
		return nil
	}

	return keyspaces.Writer().Session()
}

// SiteSettingRepository returns the configured site-setting repository.
func (s *Service) SiteSettingRepository() (repository models.SiteSettingRepository) {
	return _Deps.SiteSettingRepository
}

// SiteSettingTagRepository returns the configured site-setting-tag repository.
func (s *Service) SiteSettingTagRepository() (repository models.SiteSettingTagRepository) {
	return _Deps.SiteSettingTagRepository
}

// MaintenanceLogRepository returns the configured Cassandra maintenance-log repository.
func (s *Service) MaintenanceLogRepository() (repository models.MaintenanceLogRepository) {
	return _Deps.MaintenanceLogRepository
}

// ListSettings returns all site settings in repository order.
// Repository failures follow the repository contract and return an empty slice.
func (s *Service) ListSettings() (settings []*models.SiteSetting) {
	return s.SiteSettingRepository().List()
}

// FindSettings returns site settings after applying the supplied typed options.
// A query failure is logged and returned as nil under the service's list contract.
func (s *Service) FindSettings(opts ...models.DatabaseQueryOption[*models.SiteSetting]) (settings []*models.SiteSetting) {
	results, err := s.SiteSettingRepository().Find(opts...)
	if err != nil {
		kklogger.ErrorJ("example:Service.FindSettings#fetch!db_error", err.Error())
		return nil
	}

	return results
}

// ListSettingsWithTags returns all site settings with lazy associations batch-loaded.
// A primary query failure is logged and returned as nil.
func (s *Service) ListSettingsWithTags() (settings []*models.SiteSetting) {
	results, err := s.SiteSettingRepository().Find(models.EagerAll[*models.SiteSetting]())
	if err != nil {
		kklogger.ErrorJ("example:Service.ListSettingsWithTags#fetch!db_error", err.Error())
		return nil
	}

	return results
}

// UpdateSettingTx updates non-zero fields on setting through the caller-owned transaction.
// It returns affected rows and reports database failure through hasError.
func (s *Service) UpdateSettingTx(tx *gorm.DB, setting *models.SiteSetting) (rowsAffected int64, hasError bool) {
	result := tx.Model(setting).Updates(setting)
	if result.Error != nil {
		kklogger.ErrorJ("example:Service.UpdateSettingTx#update!db_error", result.Error.Error())
		return 0, true
	}

	return result.RowsAffected, false
}

// GetMaintenanceLog returns one maintenance log by type and key.
// A disabled Cassandra connector or a missing record returns nil.
func (s *Service) GetMaintenanceLog(typ, key string) (log *models.MaintenanceLog) {
	if _KeyspacesWriterSession() == nil {
		return nil
	}

	return s.MaintenanceLogRepository().Get(typ, key)
}

// ListMaintenanceLogs returns logs for one type and their paging result.
// A disabled Cassandra connector returns an explicit empty slice and zero paging result.
func (s *Service) ListMaintenanceLogs(typ string, opts ...models.CassandraQueryOption) (logs []*models.MaintenanceLog, result models.CassandraQueryResult[*models.MaintenanceLog]) {
	if _KeyspacesWriterSession() == nil {
		return []*models.MaintenanceLog{}, models.CassandraQueryResult[*models.MaintenanceLog]{}
	}

	return s.MaintenanceLogRepository().GetAll(typ, opts...)
}
