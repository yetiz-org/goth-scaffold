package models

import (
	"time"

	"github.com/yetiz-org/goth-scaffold/app/components/crypto"
)

// ─── ID type ─────────────────────────────────────────────────────────────────

type SiteSettingTagId crypto.KeyId

const KeyTypeSiteSettingTagId crypto.KeyType = "site_setting_tag_id"

func (k SiteSettingTagId) EncryptId() (string, error) {
	return crypto.EncryptKeyId(KeyTypeSiteSettingTagId, crypto.KeyId(k))
}

func (k SiteSettingTagId) EncryptedId() string {
	return crypto.EncryptedKeyId(KeyTypeSiteSettingTagId, crypto.KeyId(k))
}

func (k SiteSettingTagId) UInt64() uint64 { return uint64(k) }

func (k SiteSettingTagId) DecryptId(enc string) (SiteSettingTagId, error) {
	v, err := crypto.DecryptKeyId[crypto.KeyId](KeyTypeSiteSettingTagId, enc)
	return SiteSettingTagId(v), err
}

// ─── Model ───────────────────────────────────────────────────────────────────

// SiteSettingTag attaches a label to a SiteSetting record.
//
// Associations:
//   - Setting() *SiteSetting — lazy belongs-to; resolved via SiteSettingId.Resolve().
//
// Usage (lazy, loads on first access):
//
//	tag.Setting()
//
// Usage (eager, N+1-free when loading many SiteSettings):
//
//	settings, _ := repo.Find(models.EagerAll[*models.SiteSetting]())
//	// Each setting.Tags() is pre-populated; no extra queries.
type SiteSettingTag struct {
	ID            SiteSettingTagId `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	SiteSettingID SiteSettingId    `json:"site_setting_id" gorm:"column:site_setting_id;not null;index"`
	Name          string           `json:"name" gorm:"column:name;size:100;not null"`
	CreatedAt     time.Time        `json:"created_at" gorm:"column:created_at;not null;autoCreateTime"`

	// Lazy belongs-to association — do NOT use in GORM Preload; use Setting() accessor.
	_Setting *SiteSetting `gorm:"foreignKey:SiteSettingID;references:ID"`
}

func (m *SiteSettingTag) TableName() string { return "site_setting_tags" }

// SetCacheSetting pre-populates the setting cache (called by batch eager loading via EagerAll).
func (m *SiteSettingTag) SetCacheSetting(v *SiteSetting) {
	if m == nil {
		return
	}

	m._Setting = v
}

// Setting returns the parent SiteSetting, loading it lazily on first access.
func (m *SiteSettingTag) Setting() *SiteSetting {
	if m == nil {
		return nil
	}

	return LazyBelongsTo[SiteSetting](m, &m._Setting)
}

// ─── Repository interface ─────────────────────────────────────────────────────

// SiteSettingTagRepository defines data-access operations for SiteSettingTag.
type SiteSettingTagRepository interface {
	DatabaseRepository[SiteSettingTagId, *SiteSettingTag]
	FindBySetting(settingID SiteSettingId, opts ...DatabaseQueryOption[*SiteSettingTag]) []*SiteSettingTag
}
