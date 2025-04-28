package smarter

import (
	"context"
	"fmt"
	"io/ioutil"
	. "mykit/core/dsp"
	. "mykit/core/internal"
	. "mykit/core/persist"
	. "mykit/core/transfer"
	. "mykit/core/types"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
	"github.com/zeromicro/go-zero/core/conf"
)

var (
	Router = SmarterRouter{}
)

func RegRpc(service string, m ...string) {
	if SERVER().NoRpc() {
		return
	}

	RegAppDispatch(service, m...)
}

func AddHandler(app string, h ...Handler) {
	DISP.Add(app, h...)
}

func Override(app string, h ...Handler) {
	DISP.Override(app, h...)
}

func Impact(app string, h Handler, method ...string) {
	DISP.Impact(app, Impactor{H: h, M: method})
}

func Forbid(app string, method ...string) {
	DISP.Forbid(app, method...)
}

func IgnoreHandlerLog(app string, f ...string) {
	DISP.IgnoreLog(app, f...)
}

func IgnoreHandlerLogData(app string, f ...string) {
	DISP.IgnoreLogData(app, f...)
}

func AddDebugHandler(f ...Handler) {
	AddHandler(appDebug, f...)
}

func AddIotHandler(f ...Handler) {
	AddHandler(RpcIot, f...)
}

func LogDisp() {
	DISP.Log()
}

func METHOD() map[string][]string {
	return DISP.Method()
}

func InitApp(raw ...string) {
	BatchSet(raw...)

	CheckNullStr(
		PROJECT(),
		TENANT(),
		INST(),
		Version(),
	)
}

func ParseEtcdConfig(param []EtcdConfig, v EtcdConfig) EtcdConfig {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ParseMysqlConfig(param []MysqlConfig, v MysqlConfig) MysqlConfig {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ParseRedisConfig(param []RedisConfig, v RedisConfig) RedisConfig {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ConcurrencyJob(ctx context.Context, timeout time.Duration, key string, f Checker) (err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	//lock := NewEtcdLock(EtcdCli(), key, int64(timeout.Seconds())+2)
	lock, err := NewEtcdLockV2(EtcdCli(), key)
	if err != nil {
		return
	}

	err = lock.Lock(ctx)

	err = f(ctx)

	lock.Unlock(ctx)
	lock.Close()

	return
}

func GinHandler() func(c *gin.Context) {
	return Router.GinHandler()
}

func GinHandler2() func(c *gin.Context) {
	return Router.GinHandler2()
}

func LpcHandler() func(c *gin.Context) {
	return Router.LpcHandler()
}

func RPC(ctx context.Context,
	app, method string, obj interface{},
	res ...interface{}) (code int32, err error) {
	return Router.Rpc(ctx, app, method, obj, res...)
}

func SafeGo(ctx context.Context, f JOB, msg ...string) {
	Go(ctx, f, msg...)
}

func TracedContext(c ...context.Context) context.Context {
	return GetTracedContext(c...)
}

func TracedGo(f JOB, msg ...string) {
	Go(GetTracedContext(), f, msg...)
}

func TracedRun(f func(), msg ...string) {
	TracedGo(
		func(ctx context.Context) {
			f()
		},
		msg...,
	)
}

func MustLoad(f string, obj interface{}) {
	f = FilePath(f)

	err := conf.Load(f, obj)
	if err != nil {
		msg := fmt.Sprintf("load %v", f)
		HandleInitErr(msg, err)
	}
}

func LoadFile(f string, must ...bool) []byte {
	f = FilePath(f)
	data, err := ioutil.ReadFile(f)
	if err != nil {
		if ParseBool(must) {
			msg := fmt.Sprintf("load %v", f)
			HandleInitErr(msg, err)
		}
	}

	return data
}

func OpenIni(f string, must ...bool) *ini.File {
	f = FilePath(f)
	cfg, err := ini.Load(f)
	if err != nil {
		if ParseBool(must) {
			msg := fmt.Sprintf("load %v", f)
			HandleInitErr(msg, err)
		}
	}

	return cfg
}

func LoadYamlStr(raw string, v interface{}) {
	LoadYaml(StringToBytes(raw), v)
}

func LoadYaml(content []byte, v interface{}) {
	err := conf.LoadFromYamlBytes(content, v)
	if err != nil {
		msg := fmt.Sprintf("load yaml, data len %v\n%v\n", len(content), content)
		HandleInitErr(msg, err)
	}
}

func SimpleServerConfig(name string, v interface{}, suffix ...string) {
	yamlStr := fmt.Sprintf("App: %v\nUri: 0.0.0.0:777", name)

	if len(suffix) > 0 {
		yamlStr += "\n" + strings.Join(suffix, "\n")
	}

	LoadYamlStr(yamlStr, v)
}

var (
	keyM        = map[string]string{}
	redisPrefix string
)

func DbKey() string {
	return keyM[KeyDb]
}

func RegMK(k, v string) {
	keyM[k] = v
}

func SetRedisPrefix(raw string) {
	if raw == "" {
		return
	}

	if !strings.HasSuffix(raw, RedisKeyDelimiter) {
		raw += RedisKeyDelimiter
	}

	redisPrefix = raw
}

func PaddingRedisKey(raw ...*string) {
	AddPrefix(redisPrefix, raw...)
}

func DecodeKey(key, raw string) []byte {
	res := DeStrParam(raw, StrOfNullJson)

	return StringToBytes(res)
}
