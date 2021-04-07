package example

type Example struct {
	ID        uint64 `gorm:"column:id;primary_key"`
	CreatedAt uint   `gorm:"column:created_at"`
	UpdatedAt uint   `gorm:"column:updated_at"`
}

func (Example) TableName() string {
	return "example"
}
