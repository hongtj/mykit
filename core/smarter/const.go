package smarter

const (
	mega         = 1024 * 1024
	stats        = true
	dbDelimiter  = "_"
	rpcDelimiter = "."
)

const (
	ConfigFile   = "config.yml"
	mysqlUriPat  = "%v:%v@tcp(%v:%v)/%v?%v"
	influxUriPat = "http://%v:%v"
)

const (
	LocalMode = iota
	DevMode
	AlphaMode
	BetaMode
	ReleaseMode

	UnitTestMode = -2
)

const (
	KeyDb      = "db"
	appDebug   = "debug"
	KeySmarter = "Smarter"
	keyNoRpc   = "NoRpc"
)

var envNameMap = map[int32]string{
	LocalMode:    "local",
	UnitTestMode: "unit test",
	DevMode:      "dev",
	AlphaMode:    "alpha",
	BetaMode:     "beta",
	ReleaseMode:  "release",
}
