package core

import (
	"context"
	"mykit/core/dsp"
	"mykit/core/redis"
	"mykit/core/smarter"
	"mykit/core/types"
	"sync"

	"github.com/go-ini/ini"
	v8 "github.com/go-redis/redis/v8"
	"github.com/go-xorm/xorm"
	"github.com/jmoiron/sqlx"
)

func init() {
	smarter.App = App
	smarter.Zone = Zone
	smarter.BuildTime = BuildTime
	smarter.Commit = Commit
	smarter.Auth = Auth
}

var (
	Project    = "sf2"
	Tenant     = ""
	AppVersion = "0.1"
)

var (
	App       = ""
	Zone      = ""
	BuildTime = ""
	Commit    = ""
	Auth      = ""
)

var (
	RootNode        = int64(1)
	DefaultCompany  = int64(1)
	DefaultWorkshop = int64(1)
	NoTeam          = ""
)

var (
	DevDebugSwitch = dsp.DevDebugSwitch
	DevDebug       = dsp.DevDebug
	Debugs         = dsp.Debugs
	Debugf         = dsp.Debugf
	NewDebugBlock  = dsp.NewDebugBlock
)

var (
	HandleInitErr = types.HandleInitErr

	CertPem smarter.CertPem
)

var (
	InitMysql = func(conf smarter.MysqlConfig) {
		smarter.InitMysql(conf)

		UseDb(context.Background())
	}

	RedisCli v8.Cmdable

	InitRedis = func(conf smarter.RedisConfig) {
		smarter.InitRedis(conf)

		RedisCli = smarter.GetRedis(0)
	}
)

func UseDb(ctx context.Context, useMaster ...bool) *sqlx.DB {
	tenant, db := TenantDb(ctx)

	return smarter.UseDB(tenant, db, useMaster...)
}

func TenantDb(ctx context.Context) (tenant, db string) {
	tenant = ""
	db = "my_dev"

	return
}

func RedisKeyBuilder(ctx context.Context) *types.StrBuilder {
	return smarter.RedisKeyBuilder(ctx)
}

var (
	initDbOnce                sync.Once
	defaultMysqlSessionOption = "charset=utf8mb4&parseTime=true"
	redisProjectPrefix        = Project + RedisKeyDelimiter
)

var (
	SqlDbx    *xorm.Engine
	RedisConn *redis.ConnPool
)

func XormDb(ctx context.Context) *xorm.Engine {
	return SqlDbx
}

func OpenIni(f ...string) *ini.File {
	return smarter.OpenIni(types.ParseStrParam(f, IniFile))
}

func OpenRedis(cfg *ini.Section) *redis.ConnPool {
	host := cfg.Key("host").String()                //redis地址
	pwd := cfg.Key("pwd").String()                  //redis密码
	dbIndex, _ := cfg.Key("db_index").Int()         //数据库序号
	maxIdle, _ := cfg.Key("max_idle").Int()         //最大空闲连接数
	maxActive, _ := cfg.Key("max_active").Int()     //最大连接数
	idleTimeout, _ := cfg.Key("idle_timeout").Int() //空闲连接超时时间(秒)
	connTimeout, _ := cfg.Key("conn_timeout").Int() //连接超时时间(秒)

	res := redis.InitRedisPool(host, pwd, dbIndex, maxIdle, maxActive, idleTimeout, connTimeout)

	return res
}
