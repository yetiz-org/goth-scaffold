package database

type Example struct {
	ID        int64 `gorm:"column:id;primary_key"`
	ExpiresAt uint  `gorm:"column:expires_at;default:4294967295"`
	CreatedAt uint  `gorm:"column:created_at"`
	UpdatedAt uint  `gorm:"column:updated_at"`
	DeletedAt *uint `gorm:"column:deleted_at"`
}
