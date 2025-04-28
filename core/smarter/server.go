package smarter

import (
	. "mykit/core/dsp"
	. "mykit/core/internal"
	. "mykit/core/persist"
	. "mykit/core/transfer"
	. "mykit/core/types"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	Basic
	LogConfig

	Etcd  EtcdConfig  `json:",optional" help:"Etcd配置"`
	Mysql MysqlConfig `json:",optional"`
	Redis RedisConfig `json:",optional"`

	Gin GinConfig `json:",optional"`
	Rpc RpcConfig `json:",optional"`
}

func (t *Server) CheckUri() {
	if t.Param == nil {
		t.Param = map[string]interface{}{}
	}

	if t.Gin.Uri != "" {
		t.ip = ParseTcp(&t.Gin.Uri)
	} else if t.Rpc.Uri != "" {
		t.ip = ParseTcp(&t.Rpc.Uri)
	} else {
		return
	}

	t.local = GenAddress(GetIp(), t.Port())
	t.public = GenAddress(GetPublic(), t.Port())

	SetAddress(t.local)
}

func (t *Server) init() {
	if INITED() {
		return
	}

	t.CheckUri()

	SetModule(t.Module)

	initEnv(t)

	InitTrace()

	InitLpc(PrimaryEnv())

	InitLog(t.LogConfig)

	Tracking(LogMsgTracking,
		LogEvent("system"),
		LogProcessor("init"),
		LogDetail(initMsg),
	)

	if t.Mysql.Db != "" {
		SetSqlxGlobalUnsafe(t.Mysql.SqlxUnsafe)
		SetSqlxGlobalIgnoreLog(t.Mysql.SqlxIgnoreLog)
		SetSqlxDebug(t.Mysql.Debug)
	}

	if PrimaryEnv() {
		DISP.ForbidApp(appDebug)
	}

	CheckApp()

	SetStatus(StatusInited)

	t.putMeta()
}

func (t Server) RunRpc(raw ...SmarterHandler) {
	rpc := t.Rpc
	rpc.Endpoint = DeStrParam(rpc.Endpoint, RpcEndpoint(t.App))
	rpc.MaxMsgSize = DeIntParam(rpc.MaxMsgSize, t.MaxMsgSize)
	rpc.Etcd = t.ETCD(rpc.Etcd)

	rpc.RunRpc(raw...)
}

func (t Server) Add(app ...string) {
	Router.Add(t.Etcd, app...)
}

func (t Server) OpenMysql(raw ...MysqlConfig) *sqlx.DB {
	conf := t.MYSQL(raw...)
	conf.Db = conf.Db
	return OpenMysql(t.Etcd, conf)
}

func (t Server) OpenRedis(raw ...RedisConfig) redis.Cmdable {
	return openRedis(t.Etcd, t.REDIS(raw...))
}
