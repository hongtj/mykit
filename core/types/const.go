package types

import (
	"context"
	"encoding/json"
	"math"
	"time"
)

var (
	StrOfNullJson                        = "{}"
	StrOfNullList                        = "[]"
	ByteOfNull                           = []byte{110, 117, 108, 108}
	ByteOfNullJson                       = []byte(StrOfNullJson)
	ByteOfNullList                       = []byte(StrOfNullList)
	RawMessageOfNullJson json.RawMessage = ByteOfNullJson
	RawMessageOfNullList json.RawMessage = ByteOfNullList
)

const (
	StrOfTrue      = "true"
	StrOfFalse     = "false"
	FlagForTrue    = 1
	FlagForFalse   = -1
	SignalOn       = 1
	SignalOff      = 0
	SignalForTrue  = 100
	SignalForFalse = 0
)

var (
	Ctx         = context.Background()
	CstTimeZone = time.FixedZone("CST", 8*3600)
)

const (
	OutputPanicStackSize = 64 << 10
	DefaultSpanStr       = "8888888888888888"
)

const (
	MinusMsTick = 100_000_000_000
	epsilon     = 1e-9
	deg2rad     = math.Pi / 180
)

const (
	EarthRadius           = 6_371_000.0    // 地球平均半径，单位为米
	EarthEquatorialRadius = 6_378_137.0    // 地球赤道半径，单位为米.0
	EarthPolarRadius      = 6_356_752.3142 // 地球极半径，单位为米.0
)

var (
	Decimal0d5   = NewDecimal(0.5)
	Decimal2     = NewDecimal(2)
	Decimal3     = NewDecimal(3)
	EarthRadiusD = NewDecimal(EarthRadius)
	deg2radD     = NewDecimal(deg2rad)
)

var (
	//e2  = (math.Pow(EarthEquatorialRadius, 2) - math.Pow(EarthPolarRadius, 2)) / math.Pow(EarthEquatorialRadius, 2)
	e2, _ = NewDecimal(EarthEquatorialRadius).Pow(Decimal2).Sub(NewDecimal(EarthPolarRadius).Pow(Decimal2)).
		Div(NewDecimal(EarthEquatorialRadius).Pow(Decimal2)).
		Float64() //第一偏心率平方
	e2D = NewDecimal(e2)
)

const (
	MarkDelimiter     = "_"
	NumDelimiter      = ","
	TagDelimiter      = ","
	HeaderDelimiter   = ", "
	KeyDelimiter      = "::"
	RedisKeyDelimiter = ":"
)

const (
	ChartTypeLine = "line"
	ChartTypeBar  = "bar"
)

const (
	LineTypeDashed = "dashed"
)

const (
	KindAccumulationCurve = 1
	KindFlatCurve         = 0
)
