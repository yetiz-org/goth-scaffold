// ExampleService demonstrates the service-layer conventions used in this scaffold.
//
// Style highlights:
//   - DEP container (`_ExampleServiceDeps`) holds repositories and DB accessor
//     function references; `_DefaultExampleServiceDeps` is initialised eagerly at
//     package load (zero connections opened, only function pointers stored).
//   - `*F` repository constructors take `func() *gorm.DB` / `func() *gocql.Session`
//     so connection resolution happens lazily on the first query.
//   - `_Deps()` returns the active container; tests inject mocks by setting
//     `_Dependency` on the service struct directly.
//   - Method names follow `services:Struct.Method#section!action` for kklogger.
//   - `Tx` suffix marks methods that participate in an external transaction.
package services

import (
	"github.com/gocql/gocql"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/connector/database"
	"github.com/yetiz-org/goth-scaffold/app/connector/keyspaces"
	"github.com/yetiz-org/goth-scaffold/app/models"
	"github.com/yetiz-org/goth-scaffold/app/repositories"
	"gorm.io/gorm"
)

// ExampleService demonstrates the service layer for the SiteSetting and
// MaintenanceLog domains. Construct one per request or share at package level —
// the struct itself is stateless aside from its dependency container.
type ExampleService struct {
	_Dependency *_ExampleServiceDeps
}

// _ExampleServiceDeps groups the repositories and DB accessors that an
// ExampleService instance needs.
//
// All fields except *Repository are bare function references (no parens) so
// that connection lookup defers to the connectors instead of running at
// package init time.
type _ExampleServiceDeps struct {
	SiteSettingRepository    models.SiteSettingRepository
	SiteSettingTagRepository models.SiteSettingTagRepository
	MaintenanceLogRepository models.MaintenanceLogRepository
	ReaderDB                 func() *gorm.DB
	WriterDB                 func() *gorm.DB
	// CassandraSession resolves the writer keyspace session. The scaffold uses
	// a single Cassandra session per request path; if a read/write split is
	// needed later, add a separate ReaderSession field rather than aliasing.
	CassandraSession func() *gocql.Session
}

// _DefaultExampleServiceDeps is the process-wide default dependency set.
// Eagerly initialised at package load; tests bypass it via the `_Dependency`
// field on the ExampleService struct.
var _DefaultExampleServiceDeps = _ExampleServiceDeps{
	SiteSettingRepository:    repositories.NewSiteSettingRepositoryF(database.Writer),
	SiteSettingTagRepository: repositories.NewSiteSettingTagRepositoryF(database.Writer),
	MaintenanceLogRepository: repositories.NewMaintenanceLogRepositoryF(_KeyspacesWriterSession),
	ReaderDB:                 database.Reader,
	WriterDB:                 database.Writer,
	CassandraSession:         _KeyspacesWriterSession,
}

// _KeyspacesWriterSession adapts the keyspaces operator to a session function
// reference so a closure literal does not need to be inlined into the deps map.
func _KeyspacesWriterSession() *gocql.Session {
	if !keyspaces.Enabled() {
		return nil
	}

	return keyspaces.Writer().Session()
}

// NewExampleService returns a new ExampleService bound to the package-level
// default dependency set.
func NewExampleService() *ExampleService {
	return &ExampleService{_Dependency: &_DefaultExampleServiceDeps}
}

// _Deps returns the active dependency container. Tests should set
// `_Dependency` on the service to inject mocks; production code receives
// the eager-initialised `_DefaultExampleServiceDeps` automatically.
func (s *ExampleService) _Deps() *_ExampleServiceDeps {
	if s != nil && s._Dependency != nil {
		return s._Dependency
	}

	return &_DefaultExampleServiceDeps
}

// SiteSettingRepository exposes the site-setting repository.
func (s *ExampleService) SiteSettingRepository() models.SiteSettingRepository {
	return s._Deps().SiteSettingRepository
}

// SiteSettingTagRepository exposes the site-setting-tag repository.
func (s *ExampleService) SiteSettingTagRepository() models.SiteSettingTagRepository {
	return s._Deps().SiteSettingTagRepository
}

// MaintenanceLogRepository exposes the Cassandra maintenance-log repository.
func (s *ExampleService) MaintenanceLogRepository() models.MaintenanceLogRepository {
	return s._Deps().MaintenanceLogRepository
}

// ListSettings returns every site setting ordered by category and key.
//
// 業務邏輯：
//   - 單純委派給 SiteSettingRepository.List() 並回傳全量結果。
//   - 失敗時 repository 內部已記錄錯誤，本層回傳空 slice。
//
// 回傳：
//   - []*models.SiteSetting：所有 SiteSetting 列表，可能為空。
func (s *ExampleService) ListSettings() []*models.SiteSetting {
	return s.SiteSettingRepository().List()
}

// FindSettings 套用 DatabaseQueryOption 後查詢 SiteSetting。
//
// 業務邏輯：
//   - 將 opts 直接傳入 repository.Find，支援 PaginationOpt / EagerAll / GormOpt 等組合。
//   - 失敗時記錄錯誤後回傳 nil；成功回傳 repository 結果（可為空 slice）。
//
// 參數：
//   - opts: 可選的 DatabaseQueryOption[*models.SiteSetting] 變參。
//
// 回傳：
//   - []*models.SiteSetting：符合 opts 的查詢結果；查詢錯誤時為 nil。
func (s *ExampleService) FindSettings(opts ...models.DatabaseQueryOption[*models.SiteSetting]) []*models.SiteSetting {
	results, err := s.SiteSettingRepository().Find(opts...)
	if err != nil {
		kklogger.ErrorJ("services:ExampleService.FindSettings#fetch!db_error", err.Error())
		return nil
	}

	return results
}

// ListSettingsWithTags 一次取出所有 SiteSetting 並批次預載 Tags。
//
// 業務邏輯：
//   - 透過 EagerAll 觸發 repository 的批次預載，避免 N+1 查詢。
//   - DB 查詢錯誤時記錄後回傳 nil。
//   - EagerAll 內單一關聯預載失敗僅記錄 log，並把該關聯快取設為空 slice／nil；
//     後續 lazy getter 不會再重試該關聯。
//
// 回傳：
//   - []*models.SiteSetting：每筆已預載 Tags() 的結果，無資料時為空 slice。
func (s *ExampleService) ListSettingsWithTags() []*models.SiteSetting {
	results, err := s.SiteSettingRepository().Find(models.EagerAll[*models.SiteSetting]())
	if err != nil {
		kklogger.ErrorJ("services:ExampleService.ListSettingsWithTags#fetch!db_error", err.Error())
		return nil
	}

	return results
}

// UpdateSettingTx 在外部 transaction 內以 GORM Updates 更新 SiteSetting。
//
// 業務邏輯：
//   - 直接呼叫 tx.Model(setting).Updates(setting)；GORM 僅更新非零值欄位。
//   - 任何 DB error 都記錄後以 hasError=true 回傳，呼叫端可立即 rollback。
//
// 參數：
//   - tx: 外部 GORM transaction（非 nil）。
//   - setting: 要更新的 SiteSetting，ID 必須非零。
//
// 回傳：
//   - rowsAffected: 受影響的列數。
//   - hasError: true 表示更新時發生 DB error。
func (s *ExampleService) UpdateSettingTx(tx *gorm.DB, setting *models.SiteSetting) (rowsAffected int64, hasError bool) {
	result := tx.Model(setting).Updates(setting)
	if result.Error != nil {
		kklogger.ErrorJ("services:ExampleService.UpdateSettingTx#update!db_error", result.Error.Error())
		return 0, true
	}

	return result.RowsAffected, false
}

// GetMaintenanceLog 透過 Cassandra repository 以 (type, key) 取單筆 MaintenanceLog。
//
// 業務邏輯：
//   - keyspaces 未啟用時回傳 nil（不呼叫底層 session）。
//   - 啟用時委派給 MaintenanceLogRepository.Get；查無資料時回傳 nil。
//
// 參數：
//   - typ: log 類型欄位（例如 "deploy"、"alert"）。
//   - key: 同一類型下的唯一索引（例如時間戳或 ID）。
//
// 回傳：
//   - *models.MaintenanceLog：對應記錄；查無時為 nil。
func (s *ExampleService) GetMaintenanceLog(typ, key string) *models.MaintenanceLog {
	if s._Deps().CassandraSession == nil || s._Deps().CassandraSession() == nil {
		return nil
	}

	return s.MaintenanceLogRepository().Get(typ, key)
}

// ListMaintenanceLogs 取得指定類型的全部 MaintenanceLog 並回傳分頁狀態。
//
// 業務邏輯：
//   - keyspaces 未啟用時回傳空 slice 與零值 result（不呼叫底層 session）。
//   - 啟用時將 opts 傳入 MaintenanceLogRepository.GetAll，支援 PageState / Limit。
//
// 參數：
//   - typ: log 類型欄位。
//   - opts: 可選的 CassandraQueryOption（例如 CassandraQueryLimit、CassandraQueryNext）。
//
// 回傳：
//   - logs: 查詢結果切片，無資料時為空 slice。
//   - result: 包含 NextId 與分頁中繼資料，可呼叫 result.Next() 取下一頁。
func (s *ExampleService) ListMaintenanceLogs(typ string, opts ...models.CassandraQueryOption) (logs []*models.MaintenanceLog, result models.CassandraQueryResult[*models.MaintenanceLog]) {
	if s._Deps().CassandraSession == nil || s._Deps().CassandraSession() == nil {
		return []*models.MaintenanceLog{}, models.CassandraQueryResult[*models.MaintenanceLog]{}
	}

	return s.MaintenanceLogRepository().GetAll(typ, opts...)
}
