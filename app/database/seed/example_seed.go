package seed

import (
	"fmt"
	"time"

	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/connector/database"
	"github.com/yetiz-org/goth-scaffold/app/models"
	"github.com/yetiz-org/goth-scaffold/app/repositories"
)

func init() {
	Register(&SiteSettingSeed{})
}

// SiteSettingSeed seeds the default site settings.
// Safe to run multiple times — uses Upsert to avoid duplicates.
type SiteSettingSeed struct{}

func (s *SiteSettingSeed) Name() string { return "site_setting_seed" }
func (s *SiteSettingSeed) Order() int   { return 10 }

var _defaultSettings = []struct {
	category    string
	key         string
	value       any
	description string
}{
	{"app", "maintenance_mode", false, "Toggle maintenance mode — true to block all requests"},
	{"app", "version", "1.0.0", "Current application version string"},
	{"feature", "registration_enabled", true, "Allow new user registration"},
}

// Run seeds default site settings using Upsert (idempotent).
func (s *SiteSettingSeed) Run() error {
	kklogger.InfoJ("seed:SiteSettingSeed.Run#start!seed", "seeding site settings")

	repo := repositories.NewSiteSettingRepository(database.Writer())
	now := time.Now()
	count := 0

	for _, d := range _defaultSettings {
		val, err := models.NewSiteSettingValue(d.value)
		if err != nil {
			kklogger.ErrorJ("seed:SiteSettingSeed.Run#encode!json_error",
				fmt.Sprintf("key=%s/%s: %v", d.category, d.key, err))
			return err
		}

		desc := d.description
		setting := &models.SiteSetting{
			Category:       d.category,
			Key:            d.key,
			Value:          val,
			Default:        true,
			EffectiveStart: now,
			Description:    &desc,
		}

		if err := repo.Upsert(setting, map[string]any{
			"category": d.category,
			"key":      d.key,
			"default":  true,
		}); err != nil {
			kklogger.ErrorJ("seed:SiteSettingSeed.Run#upsert!db_error",
				fmt.Sprintf("key=%s/%s: %v", d.category, d.key, err))
			return err
		}

		count++
	}

	kklogger.InfoJ("seed:SiteSettingSeed.Run#done!success",
		fmt.Sprintf("seeded %d site settings", count))

	return nil
}
