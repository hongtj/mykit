package smarter

import (
	"fmt"
	. "mykit/core/dsp"
	. "mykit/core/internal"
	. "mykit/core/transfer"
	. "mykit/core/types"
	"net"
	"time"
)

var defaultStatusChecker StatusChecker = func() int32 {
	return RunningStatus()
}

func IsLocal() bool {
	return Env() == LocalMode
}

func IsUnitTest() bool {
	return Env() == UnitTestMode
}

func IsDev() bool {
	return Env() == DevMode
}

func IsAlpha() bool {
	return Env() == AlphaMode
}

func IsBeta() bool {
	return Env() == BetaMode
}

func IsRelease() bool {
	return Env() == ReleaseMode
}

func NoneLocal() bool {
	return Env() >= DevMode
}

func PrimaryEnv() bool {
	return IsBeta() || IsRelease()
}

func DevEnv() bool {
	return Env() < AlphaMode
}

func ParseEnvName(n ...int32) string {
	return envNameMap[ParseInt32Param(n, Env())]
}

func InitEnv(n int32) {
	if envNameMap[n] == "" {
		msg := fmt.Sprintf("env %v", n)
		HandleInitErr(msg, ErrInvalidParam)
	}

	SetEnv(n)
}

func initEnv(server *Server) {
	time.Local = CstTimeZone

	env = LoadEnv(server.Etcd)
	InitEnv(env.Release)

	SetEnvName(ParseEnvName())

	MaxMsgSize = DeIntParam(MaxMsgSize, server.MaxMsgSize)

	t, _ := ParseTimeStrWithLayout("2006-01-02_15:04:05", BuildTime)

	initMsg = fmt.Sprintf("build @ %v, commit: [%v]", t.Format(CompactTimeLayoutYyyyMmDdMm), Commit)
	addr := server.Address()
	if addr != "" {
		initMsg += fmt.Sprintf(", address [%v]", addr)
	}

	fmt.Println(fmt.Sprintf("init server [%v], %v env, ", INST(), EnvNAME()) + initMsg)
}

var (
	initMsg   = ""
	app       = Server{}
	env       = EnvConfig{}
	App       = ""
	Zone      = ""
	BuildTime = ""
	Commit    = ""
	Auth      = ""
)

func SERVER() Server {
	return app
}

func SetServer(raw Server) {
	app = raw

	app.CheckUri()

	InitTrace()

	InitLpc(PrimaryEnv())

	InitLog(app.LogConfig)

	SetRedisPrefix("r")

	SetStatus(StatusInited)
}

func ENV() EnvConfig {
	return env
}

func SETUP(raw Server) func() {
	app = raw

	app.init()

	return ServerTeardown
}

func ServerTeardown() {
	teardown()
}

func InitUnitTest(raw Server) func() {
	raw.LogConfig.SetDebug()

	app = raw

	return teardown
}

func SimpleInit(raw interface{}) func() {
	appName, ok := raw.(string)
	if ok {
		simpleServer := Server{
			Basic: Basic{
				App: appName,
			},
		}

		return InitUnitTest(simpleServer)
	}

	return SETUP(raw.(Server))
}

func UnitTestPath(raw []string, d string) {
	SetDeploy(ParseStrParam(raw, d))
}

func CheckApp() {
	if DevEnv() {
		return
	}

	if app.App != App || app.Zone != Zone {
		HandleInitErr("build mismatch", ErrInvalidParam)
	}
}

func UseSimpleAuth() bool {
	return Auth == "simple"
}

func init() {
	TeardownJobs = []func(){
		SyncLog,
	}
}

func AddTeardownJob(f ...func()) {
	TeardownJobs = append(TeardownJobs, f...)
}

func teardown() {
	n := len(TeardownJobs) - 1
	for i := n; i >= 0; i-- {
		TeardownJobs[i]()
	}
}

type Basic struct {
	App          string
	Zone         string       `json:",optional"`
	Version      string       `json:",optional"`
	StaticMeta   bool         `json:",default=true"`
	Module       int32        `json:",default=0"`
	OutputDetail bool         `json:",optional"`          //是否输出detail
	MaxMsgSize   int          `json:",default=134217728"` //max msg size
	Precision    int          `json:",default=3"`         //数据精度,小数点位数
	WritePoint   bool         `json:",default=true"`
	Gms          bool         `json:",default=true"`
	S            string       `json:",optional"`
	RootNode     uint64       `json:",default=1"`
	RootName     string       `json:",default=根"`
	RootDesc     string       `json:",default=根节点"`
	local        string       `json:",optional"`
	public       string       `json:",optional"`
	ip           *net.TCPAddr `json:",optional"`
	Param        `json:",optional"`
	TransNode
	ThirdServerControlConf
}

func (t Basic) InitRootNode(node *int64, name *string, description *string) {
	SetInt64Value(node, int64(t.RootNode))
	SetStrValue(name, t.RootName)
	SetStrValue(description, t.RootDesc)
}

func (t *Basic) Set(k string, v interface{}) {
	if t.Param == nil {
		t.Param = Param{}
	}

	t.Param.Set(k, v)
}

func (t *Basic) Get(k string) interface{} {
	if t.Param == nil {
		t.Param = Param{}
	}

	return t.Param.Get(k)
}

func (t *Basic) ApplyPlugin(k string, to []*string) {
	c, ok := t.Get(k).(PluginConf)
	if ok {
		c.Apply(to)
	}
}

func (t Basic) GetSmarter() string {
	return t.GetStr(KeySmarter)
}

func (t Basic) NoRpc() bool {
	return t.GetBool(keyNoRpc)
}

var (
	UrlDescMap = map[string]string{}
)

var (
	CommonLanguage = "Chinese"
	CommonAesKey   = "0123456789123456"
	SessionExpire  = time.Minute * 5
)

func SetUrlDescMap(raw map[string]string) {
	for k, v := range raw {
		UrlDescMap[k] = v
	}
}

func ParseUrlDesc(raw string) string {
	v, ok := UrlDescMap[raw]
	if ok {
		return v
	}

	return raw
}
