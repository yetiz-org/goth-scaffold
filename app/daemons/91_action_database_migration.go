package daemons

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pkg/errors"
	kkdaemon "github.com/yetiz-org/goth-daemon"
	kklogger "github.com/yetiz-org/goth-kklogger"
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

	return d.migration(sqlDB, "schema")
}

func (d *ActionDatabaseMigration) migration(sqlDB *sql.DB, databaseName string) error {
	source, err := (&file.File{}).Open(fmt.Sprintf("file://app/database/migrate/%s", databaseName))
	if err != nil {
		return errors.Wrap(err, "failed to open migration source")
	}

	defer source.Close()
	driver, err := mysql.WithInstance(sqlDB, &mysql.Config{})
	if err != nil {
		return errors.Wrap(err, "failed to create mysql driver instance")
	}

	m, err := migrate.NewWithInstance("file", source, "mysql", driver)
	if err != nil {
		return errors.Wrap(err, "failed to create migrate instance")
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return errors.Wrap(err, "failed to run migration up")
	}

	if errors.Is(err, migrate.ErrNoChange) {
		kklogger.InfoJ("daemons:ActionDatabaseMigration.migration#migration!no_change", fmt.Sprintf("%s no migrations to apply", databaseName))
	}

	return nil
}
