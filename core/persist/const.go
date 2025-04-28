package persist

import (
	. "mykit/core/types"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
)

const (
	OperationSoftDel = -1
	OperationInsert  = 1
	OperationDelete  = 2
	OperationUpdate  = 3
	InsertIgnore     = true
)

const (
	EtcdEventPut    = mvccpb.PUT
	EtcdEventDELETE = mvccpb.DELETE
)

const (
	TagDb        = "db"
	TagSet       = "set"
	TagJson      = "json"
	IdDb         = "id"
	UuidDb       = "uuid"
	IsDelDb      = "is_del"
	CreatedAtDb  = "created_at"
	CreatedByDb  = "created_by"
	UpdatedAtDb  = "updated_at"
	UpdatedByDb  = "updated_by"
	OperatedAtDb = "operated_at"
	OperatedByDb = "operated_by"
	DeletedAtDb  = "deleted_at"
	DeletedByDb  = "deleted_by"
)

const (
	DriverMysql   = "mysql"
	DriverPgsql   = "postgres"
	DriverSqlite3 = "sqlite3"
)

const (
	maxArgNum                = 20
	sqlQuerySlowThreshold    = time.Second
	influxQuerySlowThreshold = time.Second
)

const (
	tagTx = "mac::@#!tx!#@"
)

const (
	tagDelimiter = " "
	tagAsc       = "asc"
	tagDesc      = "desc"
	tagLike      = "like"
	tagPage      = "page"
	tagSize      = "size"
	tagOr        = "_or"
	tagOrderBy   = "_orderby"
	TagLimit     = "_limit"
	TagGroupBy   = "_groupby"
	TagHaving    = "_having"
)

const (
	logFiledSql        = "sql"
	logFiledInflux     = "influx"
	logFiledCollection = "collection"
)

const (
	logFiledData         = "data"
	logFiledLike         = "like"
	logFiledArgs         = "args"
	logFiledArgNum       = "arg_num"
	logFiledResNum       = "res_num"
	logFiledHas          = "has"
	logFiledUmap         = "uMap"
	logFiledRowsAffected = "affected"
)

const (
	RedisZeroStr     = "0"
	RedisInf         = "+inf"
	RedisNegativeInf = "-inf"
)

const (
	CacheKeyRoot     = "CacheKey"
	RedisKeyTypeStr  = ""
	RedisKeyTypeHash = "H"
	RedisKeyTypeList = "L"
	RedisKeyTypeSet  = "S"
	RedisKeyTypeZset = "Z"
)

const (
	CacheKeyStr  = RedisKeyDelimiter + CacheKeyRoot + RedisKeyTypeStr + RedisKeyDelimiter
	CacheKeyHash = RedisKeyDelimiter + CacheKeyRoot + RedisKeyTypeHash + RedisKeyDelimiter
	CacheKeyList = RedisKeyDelimiter + CacheKeyRoot + RedisKeyTypeList + RedisKeyDelimiter
	CacheKeySet  = RedisKeyDelimiter + CacheKeyRoot + RedisKeyTypeSet + RedisKeyDelimiter
	CacheKeyZset = RedisKeyDelimiter + CacheKeyRoot + RedisKeyTypeZset + RedisKeyDelimiter
)

const (
	RedisKeyAlert = "alert" + RedisKeyDelimiter
	RedisKeyAuth  = "auth" + RedisKeyDelimiter
	RedisKeyBasic = "basic" + RedisKeyDelimiter
	RedisKeyChart = "chart" + RedisKeyDelimiter
	RedisKeyMsg   = "msg" + RedisKeyDelimiter
)

//auth

const (
	CacheKeySession = RedisKeyAuth + "Session" + CacheKeyStr
	CacheKeySecret  = RedisKeyAuth + "Secret" + CacheKeyHash
	CacheKeyAccess  = RedisKeyAuth + "Access" + CacheKeySet
)

//basic

const (
	CacheKeyRuntime  = RedisKeyBasic + "Runtime" + CacheKeyStr
	CacheKeyRuntimeH = RedisKeyBasic + "Runtime" + CacheKeyHash
	CacheKeyRuntimeS = RedisKeyBasic + "Runtime" + CacheKeySet
	CacheKeyTenant   = CacheKeyRuntimeH + "tenant"
	CacheKeyPage     = CacheKeyRuntimeH + "page"
	CacheKeyGC       = RedisKeyBasic + "GC" + CacheKeySet
	CacheKeyLog      = RedisKeyBasic + "Log" + CacheKeyList
	CacheKeyUserRole = RedisKeyBasic + "UserRole" + CacheKeyStr
	CacheKeyUserInfo = RedisKeyBasic + "UserInfo" + CacheKeyHash
)
