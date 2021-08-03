package app

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/profiler"
	"github.com/kklab-com/gone/http"
	"github.com/kklab-com/gone/http/httpsession/redis"
	daemon "github.com/kklab-com/goth-daemon"
	datastore "github.com/kklab-com/goth-kkdatastore"
	kkgeoip "github.com/kklab-com/goth-kkgeoip"
	"github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-kksecret"
	"github.com/kklab-com/goth-kkstdcatcher"
	"github.com/kklab-com/goth-scaffold/app/build_info"
	"github.com/kklab-com/goth-scaffold/app/conf"
)

var (
	help       bool
	configPath string
)

func Init() {
	defer deferInitPanic()
	FlagParse()
	EnvironmentInit()
	DatabaseInit()
	RedisInit()
	LoggerInit()
	StdRedirectInit()
	HttpSessionInit()
	ProfilerInit()
	RegisterService()
	daemon.Start()
}

func RegisterService() {
}

func deferInitPanic() {
	if e := recover(); e != nil {
		fmt.Sprintln(e)
	}
}

func FlagParse() {
	flag.StringVar(&configPath, "c", "config.yaml", "config file")
	flag.BoolVar(&help, "h", false, "help")
	flag.Parse()

	if help {
		fmt.Println(fmt.Sprintf("BuildGitBranch: %s", build_info.BuildGitBranch))
		fmt.Println(fmt.Sprintf("BuildGitVersion: %s", build_info.BuildGitVersion))
		fmt.Println(fmt.Sprintf("BuildGoVersion: %s", build_info.BuildGoVersion))
		fmt.Println(fmt.Sprintf("BuildTimestamp: %s", build_info.BuildTimestamp))
		flag.Usage()
		os.Exit(0)
	}

	if configPath == "" {
		println("config path can't be set empty")
		os.Exit(1)
	}

	conf.ConfigPath = configPath
}

func EnvironmentInit() {
	kksecret.PATH = conf.Config().DataStore.KKSecretPath
	kkgeoip.GeoIPDBDirPath = conf.Config().DataStore.GeoIPPath
	os.Setenv("KKAPP_ENVIRONMENT", conf.Config().App.Environment.String())
	if strings.ToUpper(conf.Config().App.Environment.String()) != "PRODUCTION" {
		os.Setenv("KKAPP_DEBUG", "TRUE")
	}
}

func DatabaseInit() {
	datastore.KKDBParamDialTimeout = "3s"
	datastore.KKDBParamReaderMaxOpenConn = 64
	datastore.KKDBParamReaderMaxIdleConn = 32
	datastore.KKDBParamWriterMaxOpenConn = 64
	datastore.KKDBParamWriterMaxIdleConn = 32
	datastore.KKDBParamConnMaxLifetime = 3600000
}

func RedisInit() {
	redis.RedisName = conf.Config().DataStore.RedisName
}

func StdRedirectInit() {
	kkstdcatcher.StdoutWriteFunc = func(s string) {
		kklogger.InfoJ("STDOUT", s)
	}

	kkstdcatcher.StderrWriteFunc = func(s string) {
		kklogger.ErrorJ("STDERR", s)
	}

	kkstdcatcher.Start()
}

func LoggerInit() {
	// create logger path
	if conf.Config().Logger.LoggerPath != "" {
		kklogger.LoggerPath = conf.Config().Logger.LoggerPath
	}

	if _, err := os.Stat(kklogger.LoggerPath); os.IsNotExist(err) {
		if err := os.MkdirAll(kklogger.LoggerPath, 0755); err != nil {
			fmt.Println("logger path create fail")
			panic("logger path create fail")
		}
	}

	kklogger.Environment = conf.Config().App.Environment.String()
	kklogger.SetLogLevel(conf.Config().Logger.LogLevel)
	kklogger.Info("Logger Initialization")
}

func HttpSessionInit() {
	switch strings.ToUpper(conf.Config().Http.SessionType) {
	case string(http.SessionTypeMemory):
		http.DefaultSessionType = http.SessionTypeMemory
	case string(http.SessionTypeRedis):
		http.DefaultSessionType = http.SessionTypeRedis
		redis.RedisSessionPrefix = fmt.Sprintf("%s:%s:hs", conf.Config().Http.SessionKey, conf.Config().App.Environment)
	}

	http.SessionKey = conf.Config().Http.SessionKey
	http.SessionDomain = conf.Config().Http.SessionDomain.String()
}

func ProfilerInit() {
	if conf.Config().Profiler.Enable {
		timer := time.NewTimer(time.Second)
		config := profiler.Config{
			Service:              "go-scaffold",
			ServiceVersion:       fmt.Sprintf("%s(%s-%s)", conf.Config().App.Environment, build_info.BuildGitVersion[:8], build_info.BuildTimestamp),
			MutexProfiling:       conf.Config().Profiler.MutexProfiling,
			NoAllocProfiling:     conf.Config().Profiler.NoAllocProfiling,
			NoHeapProfiling:      conf.Config().Profiler.NoHeapProfiling,
			NoGoroutineProfiling: conf.Config().Profiler.NoGoroutineProfiling,
			ProjectID:            conf.Config().Profiler.ProjectID,
		}
		for {
			<-timer.C
			if err := profiler.Start(config); err != nil {
				kklogger.ErrorJ("GCPProfiler", map[string]interface{}{"status": "fail", "error": err.Error(), "config": config})
				timer.Reset(time.Second * 1)
			} else {
				kklogger.InfoJ("GCPProfiler", map[string]interface{}{"status": "success", "config": config})
				return
			}
		}
	}
}
