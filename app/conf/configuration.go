package conf

import (
	"io/ioutil"
	"sync"

	"github.com/yetiz-org/goth-kkutil/xtype"
	"gopkg.in/yaml.v2"
)

var (
	ConfigPath     = ""
	configInstance *Configuration
	configLock     sync.Mutex
)

type Configuration struct {
	DataStore struct {
		KKSecretPath string `yaml:"kksecret_path"`
		GeoIPPath    string `yaml:"geoip_path"`
		DatabaseName string `yaml:"database_name"`
		RedisName    string `yaml:"redis_name"`
	} `yaml:"datastore"`
	App struct {
		Name        xtype.String `yaml:"name"`
		DomainName  xtype.String `yaml:"domain_name"`
		Port        int          `yaml:"port"`
		Environment xtype.String `yaml:"environment"`
	} `yaml:"app"`
	Credentials struct {
		Rollbar struct {
			Token xtype.String `yaml:"token"`
		} `yaml:"rollbar"`
	} `yaml:"credentials"`
	Http struct {
		Scheme            string       `yaml:"scheme"`
		SessionType       string       `yaml:"session_type"`
		SessionKey        string       `yaml:"session_key"`
		SessionDomain     xtype.String `yaml:"session_domain"`
		SessionExpireTime int          `yaml:"session_expire_time"`
	} `yaml:"http"`
	Websocket struct {
		Scheme      string `yaml:"scheme"`
		CheckOrigin bool   `yaml:"check_origin"`
	} `yaml:"websocket"`
	Lang struct {
		Default string `yaml:"default"`
	} `yaml:"lang"`
	Logger struct {
		LoggerPath string `yaml:"logger_path"`
		LogLevel   string `yaml:"log_level"`
	} `yaml:"logger"`
	Profiler struct {
		Enable               bool   `yaml:"enable"`
		ProjectID            string `yaml:"project_id"`
		MutexProfiling       bool   `yaml:"mutex_profiling"`
		NoAllocProfiling     bool   `yaml:"no_alloc_profiling"`
		NoHeapProfiling      bool   `yaml:"no_heap_profiling"`
		NoGoroutineProfiling bool   `yaml:"no_goroutine_profiling"`
	} `yaml:"profiler"`
}

func Config() *Configuration {
	if configInstance == nil {
		configLock.Lock()
		defer configLock.Unlock()
		if configInstance == nil {
			configInstance = new(Configuration)
			configInstance._Init()
			if configInstance.Logger.LoggerPath == "" {
				configInstance.Logger.LoggerPath = "alloc/logs/"
			}
		}
	}

	return configInstance
}

func (c *Configuration) _Init() {
	if bytes, err := ioutil.ReadFile(ConfigPath); err == nil {
		yaml.Unmarshal(bytes, configInstance)
	}
}

func (c *Configuration) Reload() {
	configInstance = nil
}
