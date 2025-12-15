package conf

import (
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

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

func (c *Configuration) _Init() {
	if bytes, err := os.ReadFile(ConfigPath); err == nil {
		_ = yaml.Unmarshal(bytes, configInstance)
	}
}

func (c *Configuration) Reload() {
	configInstance = nil
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
