package models

type MaintenanceLog struct {
	Type      string   `json:"type" cql:"type"`
	Key       string   `json:"key" cql:"key"`
	Metadata  Metadata `json:"metadata" cql:"metadata"`
	Value     string   `json:"value" cql:"value"`
	CreatedAt int64    `json:"created_at" cql:"created_at"`
	UpdatedAt int64    `json:"updated_at" cql:"updated_at"`
}

func (m *MaintenanceLog) TableName() string {
	return "maintenance_logs"
}

type MaintenanceLogRepository interface {
	CassandraRepository[*MaintenanceLog]
	Get(typ, key string) (maintenanceLog *MaintenanceLog)
	GetAll(typ string, opts ...QueryOption) (maintenanceLogs []*MaintenanceLog, queryResult QueryResult[*MaintenanceLog])
	GetByKeyRange(typ, keyStart, keyEnd string, includeEnd bool, opts ...QueryOption) (maintenanceLogs []*MaintenanceLog, queryResult QueryResult[*MaintenanceLog])
}
