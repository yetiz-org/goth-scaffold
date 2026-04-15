package conf_test

import (
	"os"
	"testing"

	"github.com/yetiz-org/goth-scaffold/app/conf"
)

func resetConfig() {
	conf.ConfigPath = ""
	conf.Config().Reload()
}

func TestEnvPrefixConstant(t *testing.T) {
	if conf.EnvPrefix != "APP" {
		t.Errorf("expected EnvPrefix to be APP, got %q", conf.EnvPrefix)
	}
}

func TestLoadFromEnvOverridesAppPort(t *testing.T) {
	resetConfig()
	t.Setenv("APP_APP_PORT", "9999")

	c := conf.Config()
	if c.App.Port != 9999 {
		t.Errorf("expected App.Port=9999, got %d", c.App.Port)
	}
	resetConfig()
}

func TestLoadFromEnvOverridesDatastoreDatabaseName(t *testing.T) {
	resetConfig()
	t.Setenv("APP_DATASTORE_DATABASE_NAME", "mydb")

	c := conf.Config()
	if c.DataStore.DatabaseName != "mydb" {
		t.Errorf("expected DataStore.DatabaseName=mydb, got %q", c.DataStore.DatabaseName)
	}
	resetConfig()
}

func TestLoadFromEnvOverridesAppEnvironment(t *testing.T) {
	resetConfig()
	t.Setenv("APP_APP_ENVIRONMENT", "production")

	c := conf.Config()
	if c.App.Environment.Upper() != "PRODUCTION" {
		t.Errorf("expected App.Environment=PRODUCTION, got %q", c.App.Environment.Upper())
	}
	resetConfig()
}

func TestLoadFromEnvOverridesLoggerStdioCapture(t *testing.T) {
	resetConfig()
	t.Setenv("APP_LOGGER_STDIO_CAPTURE", "true")

	c := conf.Config()
	if !c.Logger.StdioCapture {
		t.Errorf("expected Logger.StdioCapture=true")
	}
	resetConfig()
}

func TestLoadFromEnvIgnoresEmptyValue(t *testing.T) {
	resetConfig()
	os.Unsetenv("APP_APP_PORT")

	c := conf.Config()
	if c.App.Port != 0 {
		t.Errorf("expected App.Port=0 when env var absent, got %d", c.App.Port)
	}
	resetConfig()
}

func TestEnvMappingsContainsExpectedKeys(t *testing.T) {
	resetConfig()
	c := conf.Config()
	mappings := c.EnvMappings()

	expectedKeys := []string{
		"APP_APP_PORT",
		"APP_APP_ENVIRONMENT",
		"APP_APP_NAME",
		"APP_DATASTORE_DATABASE_NAME",
		"APP_LOGGER_LOG_LEVEL",
		"APP_HTTP_SESSION_KEY",
	}

	for _, key := range expectedKeys {
		if _, ok := mappings[key]; !ok {
			t.Errorf("expected env key %q to be present in EnvMappings", key)
		}
	}
	resetConfig()
}

func TestPrintEnvMappingsIsSorted(t *testing.T) {
	resetConfig()
	c := conf.Config()
	entries := c.PrintEnvMappings()

	for i := 1; i < len(entries); i++ {
		if entries[i].EnvKey < entries[i-1].EnvKey {
			t.Errorf("PrintEnvMappings not sorted: %q before %q", entries[i-1].EnvKey, entries[i].EnvKey)
		}
	}
	resetConfig()
}

func TestStringMethods(t *testing.T) {
	s := conf.String("hello-world")

	if s.Upper() != "HELLO-WORLD" {
		t.Errorf("Upper() = %q, want HELLO-WORLD", s.Upper())
	}
	if s.Lower() != "hello-world" {
		t.Errorf("Lower() = %q, want hello-world", s.Lower())
	}
	if s.Short() != "HW" {
		t.Errorf("Short() = %q, want HW", s.Short())
	}
	if s.Title() != "Hello-world" {
		t.Errorf("Title() = %q, want Hello-world", s.Title())
	}
}
