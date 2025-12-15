package keyspaces

import (
	"fmt"
	"os"
	"sync"

	"github.com/gocql/gocql"
	datastore "github.com/yetiz-org/goth-datastore"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/conf"
)

var once sync.Once
var cassandra *datastore.Cassandra

func _Init() {
	once.Do(func() {
		if !Enabled() {
			return
		}
		
		cassandra = datastore.NewCassandra(conf.Config().DataStore.CassandraName)
		cassandra.Writer().Config().DisableInitialHostLookup = true
		cassandra.Reader().Config().DisableInitialHostLookup = true
		cassandra.Writer().Config().ConnectObserver = &connectObserver{}
		cassandra.Reader().Config().ConnectObserver = &connectObserver{}
		if _, err := os.Stat("/usr/local/share/ca-certificates/sf-class2-root.crt"); err == nil {
			if cassandra.Writer() != nil {
				cassandra.Writer().Config().SslOpts.CaPath = ""
			}

			if cassandra.Reader() != nil {
				cassandra.Reader().Config().SslOpts.CaPath = ""
			}
		}
	})
}

type connectObserver struct{}

func (c *connectObserver) ObserveConnect(connect gocql.ObservedConnect) {
	if connect.Err != nil {
		kklogger.WarnJ("datastore:CassandraOp.ObserveConnect", connect)
	} else {
		kklogger.DebugJ("datastore:CassandraOp.ObserveConnect", fmt.Sprintf("new connection to %s", connect.Host))
	}
}

func Enabled() bool {
	return conf.Config().DataStore.CassandraName != ""
}

func Instance() *datastore.Cassandra {
	_Init()
	return cassandra
}

func HealthCheck() error {
	if err := Writer().Session().Query("select table_name from system_schema.tables limit 1").PageSize(1).Exec(); err != nil {
		return err
	}

	if err := Reader().Session().Query("select table_name from system_schema.tables limit 1").PageSize(1).Exec(); err != nil {
		return err
	}

	return nil
}

func Writer() datastore.CassandraOperator {
	return Instance().Writer()
}

func Reader() datastore.CassandraOperator {
	return Instance().Reader()
}
