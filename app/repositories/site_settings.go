package repositories

import (
	"time"

	"github.com/pkg/errors"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/models"
	"gorm.io/gorm"
)

type SiteSettingRepository struct {
	models.DatabaseRepository[*models.SiteSetting]
}

func NewSiteSettingRepository(db *gorm.DB) models.SiteSettingRepository {
	return &SiteSettingRepository{
		NewDatabaseDefaultRepository[*models.SiteSetting](db),
	}
}

func (r *SiteSettingRepository) Get(id int64) *models.SiteSetting {
	model := new(models.SiteSetting)
	if err := r.DB().First(model, id).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			kklogger.ErrorJ("repositories:SiteSettingRepository.Get#query!db_error", err.Error())
		}
		return nil
	}
	return model
}

func (r *SiteSettingRepository) GetByKey(category, key string) *models.SiteSetting {
	model := new(models.SiteSetting)
	if err := r.DB().Where("category = ? AND `key` = ?", category, key).
		Order("effective_start DESC").
		First(model).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			kklogger.ErrorJ("repositories:SiteSettingRepository.GetByKey#query!db_error", err.Error())
		}
		return nil
	}
	return model
}

func (r *SiteSettingRepository) GetEffectiveByKey(category, key string, now time.Time) *models.SiteSetting {
	model := new(models.SiteSetting)
	query := r.DB().Where("category = ? AND `key` = ?", category, key).
		Where("effective_start <= ?", now)

	// Handle effective_end - either it's NULL or it's after now
	query = query.Where("effective_end IS NULL OR effective_end >= ?", now)

	if err := query.Order("`default` ASC, effective_start DESC").
		First(model).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			kklogger.ErrorJ("repositories:SiteSettingRepository.GetEffectiveByKey#query!db_error", err.Error())
		}
		return nil
	}
	return model
}

func (r *SiteSettingRepository) GetAllByCategory(category string) []*models.SiteSetting {
	settings := make([]*models.SiteSetting, 0)
	if err := r.DB().Where("category = ?", category).
		Order("effective_start DESC").
		Find(&settings).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			kklogger.ErrorJ("repositories:SiteSettingRepository.GetAllByCategory#query!db_error", err.Error())
		}
		return []*models.SiteSetting{}
	}
	return settings
}

func (r *SiteSettingRepository) GetEffectiveByCategory(category string, now time.Time) []*models.SiteSetting {
	settings := make([]*models.SiteSetting, 0)
	query := r.DB().Where("category = ?", category).
		Where("effective_start <= ?", now).
		Where("effective_end IS NULL OR effective_end >= ?", now)

	if err := query.Order("`default` ASC, effective_start DESC").
		Find(&settings).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			kklogger.ErrorJ("repositories:SiteSettingRepository.GetEffectiveByCategory#query!db_error", err.Error())
		}
		return []*models.SiteSetting{}
	}
	return settings
}

func (r *SiteSettingRepository) List() []*models.SiteSetting {
	settings := make([]*models.SiteSetting, 0)
	if err := r.DB().
		Order("category ASC, `key` ASC, effective_start DESC").
		Find(&settings).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			kklogger.ErrorJ("repositories:SiteSettingRepository.List#query!db_error", err.Error())
		}
		return []*models.SiteSetting{}
	}
	return settings
}

func (r *SiteSettingRepository) ListEffective(now time.Time) []*models.SiteSetting {
	settings := make([]*models.SiteSetting, 0)
	query := r.DB().
		Where("effective_start <= ?", now).
		Where("effective_end IS NULL OR effective_end >= ?", now)

	if err := query.Order("category ASC, `key` ASC, `default` ASC, effective_start DESC").
		Find(&settings).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			kklogger.ErrorJ("repositories:SiteSettingRepository.ListEffective#query!db_error", err.Error())
		}
		return []*models.SiteSetting{}
	}
	return settings
}
