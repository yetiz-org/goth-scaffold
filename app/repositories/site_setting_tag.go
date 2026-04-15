package repositories

import (
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/models"
	"gorm.io/gorm"
)

type SiteSettingTagRepository struct {
	*DatabaseDefaultRepository[models.SiteSettingTagId, *models.SiteSettingTag]
}

func NewSiteSettingTagRepository(db *gorm.DB) models.SiteSettingTagRepository {
	return &SiteSettingTagRepository{
		NewDatabaseDefaultRepository[models.SiteSettingTagId, *models.SiteSettingTag](db),
	}
}

// FindBySetting returns all tags for the given site setting ID.
func (r *SiteSettingTagRepository) FindBySetting(settingID models.SiteSettingId, opts ...models.DatabaseQueryOption[*models.SiteSettingTag]) []*models.SiteSettingTag {
	items, err := r.FindWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where("site_setting_id = ?", settingID)
	}, opts...)
	if err != nil {
		kklogger.ErrorJ("repositories:SiteSettingTagRepository.FindBySetting#query!failed", err.Error())
		return []*models.SiteSettingTag{}
	}

	return items
}
