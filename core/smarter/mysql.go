package smarter

import (
	"context"
	"fmt"
	. "mykit/core/dsp"
	. "mykit/core/persist"
	. "mykit/core/types"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
)

type MysqlConfig struct {
	Db                string
	Driver            string `json:",default=mysql"`
	Access            string `json:",default=main"`
	Option            string `json:",optional"`
	Debug             bool   `json:",optional"`
	MaxIdle           uint64 `json:",default=5"`
	MaxConn           uint64 `json:",default=20"`
	MaxIdleTimeMinute uint64 `json:",default=120"`
	SlowThresholdMs   uint64 `json:",default=1000"`
	SqlxUnsafe        bool   `json:",default=true"`
	SqlxIgnoreLog     bool   `json:",default=false"`
}

func (t MysqlConfig) URI(access ACCESS) string {
	option := DeStrParam(t.Option, defaultMysqlSessionOption)
	res := fmt.Sprintf(mysqlUriPat,
		access.User,
		access.Pwd,
		access.Host,
		access.Port,
		t.Db,
		option,
	)

	return res
}

func (t MysqlConfig) Name(project ...string) string {
	db := t.Db
	if strings.Contains(db, "@") && strings.Contains(db, "/") {
		db = strings.Split(db, "/")[1]
		if strings.Contains(db, "?") {
			db = strings.Split(db, "?")[0]
		}
	}

	if t.Access == db {
		return t.Access
	}

	name := t.Access + KeyDelimiter + db
	if len(project) > 0 {
		if project[0] != "" {
			name = project[0] + KeyDelimiter + name
		}
	}

	return name
}

func (t MysqlConfig) OpenMysql(raw ...EtcdConfig) *sqlx.DB {
	etcd := SERVER().ETCD(raw...)
	return OpenMysql(etcd, t)
}

var (
	initSqlxOnce              sync.Once
	initXormOnce              sync.Once
	defaultMysqlSessionOption = "charset=utf8mb4&parseTime=true"
	TenantDb                  TenantDbParser
)

var (
	sqlxDb   *sqlx.DB
	sqlxDisp = NewSqlxDisp()
)

func SetDb(db *sqlx.DB) {
	sqlxDb = db
}

func GetDb(useMaster ...bool) *sqlx.DB {
	return sqlxDb
}

func RegSqlxDb(mark string, db *sqlx.DB) {
	sqlxDisp.Reg(mark, db)
}

type SqlxDisp struct {
	sync.RWMutex
	m map[string]*sqlx.DB
}

func NewSqlxDisp() *SqlxDisp {
	res := &SqlxDisp{
		m: map[string]*sqlx.DB{},
	}

	return res
}

func (t *SqlxDisp) Reg(mark string, db *sqlx.DB) {
	t.RWMutex.Lock()
	t.m[mark] = db
	t.RWMutex.Unlock()
}

func (t *SqlxDisp) Use(tenant, db string, useMaster ...bool) *sqlx.DB {
	t.RWMutex.RLock()
	res := t.m[db]
	t.RWMutex.RUnlock()

	if res != nil {
		return res
	}

	server := SERVER()
	mysqlConf := server.Mysql
	mysqlConf.Db = db
	res = TenantMysql(tenant, server.ETCD(), mysqlConf)
	if res == nil {
		return nil
	}

	t.RWMutex.Lock()
	t.m[db] = res
	t.RWMutex.Unlock()

	return res
}

func UseDB(tenant, db string, useMaster ...bool) *sqlx.DB {
	return sqlxDisp.Use(tenant, db, useMaster...)
}

func TenantDB(ctx context.Context, useMaster ...bool) *sqlx.DB {
	tenant, db := TenantDb(ctx)

	return UseDB(tenant, db, useMaster...)
}

func InitTenantDB(f TenantDbRouter, c ...context.Context) {
	ctx := ParseContextParam(c)
	cli := f(ctx)

	if cli == nil {
		HandleInitErr("tenant db init failed", ErrNotFound)
	}
}

func InitMysql(conf MysqlConfig) {
	initSqlxOnce.Do(func() {
		if conf.Db == "" {
			return
		}

		conf.Db = conf.Db
		initMysql(conf)
		sqlxDisp.Reg(conf.Db, sqlxDb)
	})
}

func initMysql(conf MysqlConfig) {
	sqlxDb = conf.OpenMysql()
}

func SqlxOpenMysql(driver string, conf MysqlConfig, uri string) *sqlx.DB {
	cli, err := sqlx.Connect(driver, uri)
	HandleInitErr("sqlx init: "+uri, err)

	cli.SetMaxOpenConns(int(conf.MaxConn))
	cli.SetMaxIdleConns(int(conf.MaxIdle))
	cli.SetConnMaxIdleTime(MinuteTimeout(conf.MaxIdleTimeMinute))

	return cli
}

func OpenMysql(etcd EtcdConfig, raw ...MysqlConfig) *sqlx.DB {
	conf := SERVER().MYSQL(raw...)
	access := LoadAccess(etcd, conf.Access)
	driver := conf.Driver
	uri := conf.URI(access)

	return SqlxOpenMysql(driver, conf, uri)
}

func TenantMysql(tenant string, etcd EtcdConfig, conf MysqlConfig) *sqlx.DB {
	driver := conf.Driver
	access := LoadAccess(etcd, conf.Access)
	uri := conf.URI(access)

	return SqlxOpenMysql(driver, conf, uri)
}
