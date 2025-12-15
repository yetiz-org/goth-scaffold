package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// SiteSettingValue represents a flexible JSON value
type SiteSettingValue json.RawMessage

// Scan implements sql.Scanner interface for GORM JSON handling (read from DB)
func (v *SiteSettingValue) Scan(value interface{}) error {
	if value == nil {
		*v = SiteSettingValue("null")
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	*v = SiteSettingValue(bytes)
	return nil
}

// Value implements driver.Valuer interface for GORM JSON handling (write to DB)
func (v SiteSettingValue) Value() (driver.Value, error) {
	if v == nil {
		return "null", nil
	}
	return []byte(v), nil
}

// Unmarshal parses the JSON value into the target
func (v SiteSettingValue) Unmarshal(target interface{}) error {
	if v == nil {
		return nil
	}
	return json.Unmarshal(v, target)
}

// NewSiteSettingValue creates a SiteSettingValue from any value
func NewSiteSettingValue(value interface{}) (SiteSettingValue, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return SiteSettingValue(bytes), nil
}

// MustNewSiteSettingValue creates a SiteSettingValue and panics on error
func MustNewSiteSettingValue(value interface{}) SiteSettingValue {
	v, err := NewSiteSettingValue(value)
	if err != nil {
		panic(err)
	}
	return v
}

// SiteSetting represents a site configuration setting
type SiteSetting struct {
	ID             int64            `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Category       string           `json:"category" gorm:"column:category;not null;index:idx_category_key_effective"`
	Key            string           `json:"key" gorm:"column:key;not null;index:idx_category_key_effective"`
	Value          SiteSettingValue `json:"value" gorm:"column:value;type:text;not null"`
	Default        bool             `json:"default" gorm:"column:default;not null;default:false"`
	EffectiveStart time.Time        `json:"effective_start" gorm:"column:effective_start;not null;default:CURRENT_TIMESTAMP;index:idx_category_key_effective"`
	EffectiveEnd   *time.Time       `json:"effective_end" gorm:"column:effective_end;index:idx_category_key_effective"`
	Description    *string          `json:"description" gorm:"column:description"`
	CreatedAt      time.Time        `json:"created_at" gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time        `json:"updated_at" gorm:"column:updated_at;not null;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt   `json:"deleted_at" gorm:"column:deleted_at"`
}

func (m *SiteSetting) TableName() string {
	return "site_settings"
}

// IsEffective checks if the setting is currently effective
func (s *SiteSetting) IsEffective(now time.Time) bool {
	if s.DeletedAt.Valid {
		return false
	}
	if now.Before(s.EffectiveStart) {
		return false
	}
	if s.EffectiveEnd != nil && now.After(*s.EffectiveEnd) {
		return false
	}
	return true
}

// GetTypedValue returns the value as the specified type
func (s *SiteSetting) GetTypedValue(target interface{}) error {
	return s.Value.Unmarshal(target)
}

// SiteSettingRepository defines the interface for site settings repository
type SiteSettingRepository interface {
	DatabaseRepository[*SiteSetting]
	Get(id int64) *SiteSetting
	GetByKey(category, key string) *SiteSetting
	GetEffectiveByKey(category, key string, now time.Time) *SiteSetting
	GetAllByCategory(category string) []*SiteSetting
	GetEffectiveByCategory(category string, now time.Time) []*SiteSetting
	List() []*SiteSetting
	ListEffective(now time.Time) []*SiteSetting
}
