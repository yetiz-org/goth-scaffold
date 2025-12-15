package repositories

import (
	"github.com/gocql/gocql"
	"github.com/yetiz-org/goth-scaffold/app/models"
)

type MaintenanceLogRepository struct {
	models.CassandraRepository[*models.MaintenanceLog]
}

func NewMaintenanceLogRepository(session *gocql.Session) models.MaintenanceLogRepository {
	return &MaintenanceLogRepository{
		NewCassandraDefaultRepository[*models.MaintenanceLog](session),
	}
}

func (r *MaintenanceLogRepository) Get(typ, key string) (maintenanceLog *models.MaintenanceLog) {
	return r.QueryBuilder().Where("type = ?", typ).Where("key = ?", key).First()
}

func (r *MaintenanceLogRepository) GetAll(typ string, opts ...models.QueryOption) (maintenanceLogs []*models.MaintenanceLog, queryResult models.QueryResult[*models.MaintenanceLog]) {
	return r.QueryBuilder().Where("type = ?", typ).Fetch(opts...)
}

func (r *MaintenanceLogRepository) GetByKeyRange(typ, keyStart, keyEnd string, includeEnd bool, opts ...models.QueryOption) (maintenanceLogs []*models.MaintenanceLog, queryResult models.QueryResult[*models.MaintenanceLog]) {
	endCondition := "key < ?"
	if includeEnd {
		endCondition = "key <= ?"
	}

	return r.QueryBuilder().Where("type = ?", typ).Where("key >= ?", keyStart).Where(endCondition, keyEnd).Fetch(opts...)
}
