package daemons

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pkg/errors"
	kkdaemon "github.com/yetiz-org/goth-daemon"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/components/dialect"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"github.com/yetiz-org/goth-scaffold/app/connector/database"
)

type ActionDatabaseMigration struct {
	kkdaemon.DefaultDaemon
}

func (d *ActionDatabaseMigration) Start() {
	if conf.Config().DataStore.DatabaseName == "" {
		kklogger.InfoJ("daemons:ActionDatabaseMigration.Start", "database name is empty, skipping migration")
		return
	}

	kklogger.InfoJ("daemons:ActionDatabaseMigration.Start", "starting database migration")

	if err := d.runMigration(); err != nil {
		kklogger.ErrorJ("daemons:ActionDatabaseMigration.Start#scaffoldMigration!run", err.Error())
		panic(errors.Wrap(err, "failed to run scaffold migration"))
	}

	kklogger.InfoJ("daemons:ActionDatabaseMigration.Start", "database migration completed successfully")
}

func (d *ActionDatabaseMigration) runMigration() error {
	sqlDB, _ := database.Writer().DB()
	if sqlDB == nil {
		return errors.Errorf("scaffold database connection is nil")
	}

	return d.migration(sqlDB)
}

func (d *ActionDatabaseMigration) migration(sqlDB *sql.DB) error {
	active := dialect.Current()
	schemaDir := active.MigrationSchemaDir()

	source, err := (&file.File{}).Open(fmt.Sprintf("file://app/database/migrate/schema/%s", schemaDir))
	if err != nil {
		return errors.Wrap(err, "failed to open migration source")
	}

	defer source.Close()
	driver, dialectName, err := active.MigrationDriver(sqlDB)
	if err != nil {
		return errors.Wrapf(err, "failed to create %s driver instance", active.Name())
	}

	m, err := migrate.NewWithInstance("file", source, dialectName, driver)
	if err != nil {
		return errors.Wrap(err, "failed to create migrate instance")
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return errors.Wrap(err, "failed to run migration up")
	} else if errors.Is(err, migrate.ErrNoChange) {
		kklogger.InfoJ("daemons:ActionDatabaseMigration.migration#migration!no_change", fmt.Sprintf("%s no migrations to apply", schemaDir))
	}

	return nil
}
