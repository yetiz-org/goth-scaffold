package conf

import (
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const EnvPrefix = "APP"

var (
	ConfigPath     = ""
	configInstance *Configuration
	configLock     sync.Mutex
)

type Configuration struct {
	DataStore struct {
		SecretPath    string `yaml:"secret_path"`
		DatabaseName  string `yaml:"database_name"`
		CassandraName string `yaml:"cassandra_name"`
		RedisName     string `yaml:"redis_name"`
	} `yaml:"datastore"`
	App struct {
		Name        String `yaml:"name"`
		DomainName  String `yaml:"domain_name"`
		Port        int    `yaml:"port"`
		Environment String `yaml:"environment"`
		Channel     String `yaml:"channel"`
	} `yaml:"app"`
	Credentials struct {
		SecretSlack struct {
			Webhook string `yaml:"webhook"`
		} `yaml:"secret_slack"`
		Recaptcha struct {
			SiteKey   string `yaml:"site_key"`
			SecretKey string `yaml:"secret_key"`
		} `yaml:"recaptcha"`
		GTMId string `yaml:"gtm_id"`
	} `yaml:"credentials"`
	Http struct {
		Scheme            string `yaml:"scheme"`
		SessionType       string `yaml:"session_type"`
		SessionKey        string `yaml:"session_key"`
		SessionDomain     String `yaml:"session_domain"`
		SessionExpireTime int    `yaml:"session_expire_time"`
	} `yaml:"http"`
	Lang struct {
		Default string `yaml:"default"`
	} `yaml:"lang"`
	Logger struct {
		LoggerPath   string `yaml:"logger_path"`
		LogLevel     string `yaml:"log_level"`
		StdioCapture bool   `yaml:"stdio_capture"`
		GCPLogging   struct {
			Enable             bool   `yaml:"enable"`
			JSONCredentialBody string `yaml:"json_credential_body"`
		} `yaml:"gcp_logging"`
	} `yaml:"logger"`
	Profiler struct {
		Enable               bool   `yaml:"enable"`
		ProjectID            string `yaml:"project_id"`
		MutexProfiling       bool   `yaml:"mutex_profiling"`
		NoAllocProfiling     bool   `yaml:"no_alloc_profiling"`
		NoHeapProfiling      bool   `yaml:"no_heap_profiling"`
		NoGoroutineProfiling bool   `yaml:"no_goroutine_profiling"`
		JSONCredentialBody   string `yaml:"json_credential_body"`
	} `yaml:"profiler"`
}

func Config() *Configuration {
	if configInstance == nil {
		configLock.Lock()
		defer configLock.Unlock()
		if configInstance == nil {
			configInstance = new(Configuration)
			configInstance._Init()
		}
	}

	return configInstance
}

func IsDebug() bool {
	return os.Getenv("APP_DEBUG") == "true"
}

func IsProduction() bool {
	return Config().App.Environment.Upper() == "PRODUCTION"
}

func IsStaging() bool {
	return Config().App.Environment.Upper() == "STAGING"
}

func IsLocal() bool {
	return Config().App.Environment.Upper() == "LOCAL"
}

type envMapping struct {
	EnvKey   string
	YAMLPath string
	Field    reflect.Value
}

// EnvMappingEntry is a public representation of an env var to YAML path mapping.
type EnvMappingEntry struct {
	EnvKey   string `json:"env_key"`
	YAMLPath string `json:"yaml_path"`
}

func (c *Configuration) _Init() {
	if bytes, err := os.ReadFile(ConfigPath); err == nil {
		_ = yaml.Unmarshal(bytes, configInstance)
	}

	c._LoadFromEnv()
}

func (c *Configuration) Reload() {
	configInstance = nil
}

// _LoadFromEnv overrides configuration fields with values from environment variables.
// Env var names follow the pattern: APP_<SECTION>_<FIELD> (all uppercase, underscores).
// Example: app.port → APP_APP_PORT, datastore.database_name → APP_DATASTORE_DATABASE_NAME
func (c *Configuration) _LoadFromEnv() {
	for _, mapping := range c.envMappingsWithFields() {
		loadFieldFromEnv(mapping.Field, mapping.EnvKey)
	}
}

// EnvMappings returns a map of env var name → YAML path for all supported config fields.
func (c *Configuration) EnvMappings() map[string]string {
	mappings := c.envMappingsWithFields()
	results := make(map[string]string, len(mappings))
	for _, mapping := range mappings {
		results[mapping.EnvKey] = mapping.YAMLPath
	}
	return results
}

// PrintEnvMappings returns a sorted list of all env var → YAML path mappings.
func (c *Configuration) PrintEnvMappings() []EnvMappingEntry {
	mappings := c.EnvMappings()
	keys := make([]string, 0, len(mappings))
	for key := range mappings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	entries := make([]EnvMappingEntry, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, EnvMappingEntry{
			EnvKey:   key,
			YAMLPath: mappings[key],
		})
	}
	return entries
}

func (c *Configuration) envMappingsWithFields() []envMapping {
	configValue := reflect.ValueOf(c).Elem()
	configType := configValue.Type()
	mappings := make([]envMapping, 0)

	for i := 0; i < configValue.NumField(); i++ {
		fieldType := configType.Field(i)
		yamlName := parseYAMLTagName(fieldType.Tag.Get("yaml"))
		if yamlName == "" || yamlName == "-" {
			continue
		}

		fieldValue := configValue.Field(i)
		nextYAMLPrefix := yamlName
		nextEnvPrefix := EnvPrefix + "_" + strings.ToUpper(yamlName)

		if fieldValue.Kind() == reflect.Struct {
			collectEnvMappings(fieldValue, nextEnvPrefix, nextYAMLPrefix, &mappings)
			continue
		}

		mappings = append(mappings, envMapping{
			EnvKey:   nextEnvPrefix,
			YAMLPath: nextYAMLPrefix,
			Field:    fieldValue,
		})
	}

	return mappings
}

func collectEnvMappings(current reflect.Value, envPrefix string, yamlPrefix string, mappings *[]envMapping) {
	currentType := current.Type()

	for i := 0; i < current.NumField(); i++ {
		fieldType := currentType.Field(i)
		yamlName := parseYAMLTagName(fieldType.Tag.Get("yaml"))
		if yamlName == "" || yamlName == "-" {
			continue
		}

		fieldValue := current.Field(i)
		nextEnvPrefix := envPrefix + "_" + strings.ToUpper(yamlName)
		nextYAMLPrefix := yamlName
		if yamlPrefix != "" {
			nextYAMLPrefix = yamlPrefix + "." + yamlName
		}

		if fieldValue.Kind() == reflect.Struct {
			collectEnvMappings(fieldValue, nextEnvPrefix, nextYAMLPrefix, mappings)
			continue
		}

		*mappings = append(*mappings, envMapping{
			EnvKey:   nextEnvPrefix,
			YAMLPath: nextYAMLPrefix,
			Field:    fieldValue,
		})
	}
}

func loadFieldFromEnv(fieldValue reflect.Value, envKey string) {
	envValue := os.Getenv(envKey)
	if envValue == "" || !fieldValue.CanSet() {
		return
	}

	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(envValue)
	case reflect.Bool:
		if parsedBool, err := strconv.ParseBool(envValue); err == nil {
			fieldValue.SetBool(parsedBool)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if parsedInt, err := strconv.ParseInt(envValue, 10, fieldValue.Type().Bits()); err == nil {
			fieldValue.SetInt(parsedInt)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if parsedUint, err := strconv.ParseUint(envValue, 10, fieldValue.Type().Bits()); err == nil {
			fieldValue.SetUint(parsedUint)
		}
	case reflect.Float32, reflect.Float64:
		if parsedFloat, err := strconv.ParseFloat(envValue, fieldValue.Type().Bits()); err == nil {
			fieldValue.SetFloat(parsedFloat)
		}
	case reflect.Slice:
		if fieldValue.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(envValue, ",")
			values := make([]string, 0, len(parts))
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					values = append(values, trimmed)
				}
			}
			fieldValue.Set(reflect.ValueOf(values))
		}
	}
}

func parseYAMLTagName(tag string) string {
	if tag == "" {
		return ""
	}
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

type String string

func (s String) String() string {
	return string(s)
}

func (s String) Upper() string {
	return strings.ToUpper(string(s))
}

func (s String) Lower() string {
	return strings.ToLower(string(s))
}

func (s String) Short() string {
	sb := ""
	for _, sp := range strings.Split(strings.ToUpper(string(s)), "-") {
		if len(sp) > 0 {
			sb = sb + sp[0:1]
		}
	}

	return sb
}

func (s String) Title() string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(string(s[0:1])) + strings.ToLower(string(s[1:]))
}
