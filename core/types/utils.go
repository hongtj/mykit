package types

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

func ApplyIgnoreLog(raw ...IgnoreLog) {
	for _, v := range raw {
		v.IgnoreLog()
	}
}

func ApplyInit(conf interface{}, raw ...InitProgress) {
	for _, v := range raw {
		v(conf)
	}
}

func ApplyDeleteAll(raw ...DeleteAll) (err error) {
	for _, v := range raw {
		err = v.DeleteAll()
		if err != nil {
			return
		}
	}

	return
}

func ApplyTracedLog(ctx context.Context, raw ...TracedLog) {
	for _, v := range raw {
		v.NewTrace(ctx)
	}
}

func Fn(raw interface{}) *runtime.Func {
	if reflect.TypeOf(raw).Kind() != reflect.Func {
		return nil
	}

	pc := reflect.ValueOf(raw).Pointer()
	fn := runtime.FuncForPC(pc)

	return fn
}

func FnName(raw interface{}) string {
	fn := Fn(raw)
	if fn == nil {
		return ""
	}

	return fn.Name()
}

func RpcMethodName(raw interface{}) string {
	pc := reflect.ValueOf(raw).Pointer()
	fn := runtime.FuncForPC(pc)
	res := fn.Name()

	res = extractFnName(res)
	res = FirstLower(res)

	return res
}

func extractFnName(raw string) string {
	// fire/api.Fire.Ptz-fm -> Ptz

	tmp := strings.Split(raw, ".")
	l := len(tmp)

	m := tmp[l-1]
	tmp = strings.Split(m, "-")

	return tmp[0]
}

func GenAddress(ip string, port int) string {
	return fmt.Sprintf("%v:%v", ip, port)
}

func VersionToInt(raw string) int {
	raw = strings.Replace(raw, ".", "", -1)
	n, _ := strconv.Atoi(raw)

	return n
}

func Version2Flt(raw string) (version float64) {
	arr := strings.SplitN(raw, ".", 2)
	if len(arr) > 1 {
		arr[1] = strings.Replace(arr[1], ".", "", -1)
		raw = arr[0] + "." + arr[1]
	}

	version, _ = ParseFloat64FromStr(raw)

	return
}

func IntPtr(raw int) *int {
	return &raw
}

func Int32Ptr(raw int32) *int32 {
	return &raw
}

func Int64Ptr(raw int64) *int64 {
	return &raw
}

var TearDownWait = func() {
	time.Sleep(time.Second)
}

func SetStrValue(raw *string, value string) {
	if value == "" {
		return
	}

	*raw = value
}

func SetInt64Value(raw *int64, value int64) {
	if value == 0 {
		return
	}

	*raw = value
}

func SetFload64Value(raw *float64, value float64) {
	if value == 0 {
		return
	}

	*raw = value
}

func SetDuration(raw *time.Duration, value time.Duration) {
	if value == 0 {
		return
	}

	*raw = value
}

func SetMinuteDuration(raw *time.Duration, m int64) {
	if m == 0 {
		return
	}

	SetDuration(raw, time.Minute*time.Duration(m))
}

func IsInt64OrInt64Slice(raw interface{}) bool {
	_, ok := raw.(int64)
	if ok {
		return true
	}

	_, ok = raw.([]int64)

	return ok
}

func IsStrOrStrSlice(raw interface{}) bool {
	_, ok := raw.(string)
	if ok {
		return true
	}

	_, ok = raw.([]string)

	return ok
}

func NewDecimal(v float64) decimal.Decimal {
	return decimal.NewFromFloat(v)
}

func Signal32(raw int32) bool {
	return raw == SignalOn
}

func Signal64(raw int64) bool {
	return raw == SignalOn
}

func NewWriter() *SimpleWriter {
	return &SimpleWriter{}
}

func (t *SimpleWriter) Write(p []byte) (n int, err error) {
	t.b = append(t.b, p...)

	return len(t.b), nil
}

func (t *SimpleWriter) Bytes() []byte {
	return t.b
}

func (t *SimpleWriter) String() string {
	return BytesToString(t.b)
}

func (t *SimpleWriter) B64() string {
	return base64.StdEncoding.EncodeToString(t.b)
}

type SenderInst struct {
	protocol string
	remote   string
	id       ConnectionId
	send     SendFunc
}

func NewSenderInst(protocol, remote string, id ConnectionId, f SendFunc, debug ...bool) *SenderInst {
	res := &SenderInst{
		protocol: protocol,
		remote:   remote,
		id:       id,
		send:     f,
	}

	return res
}

func (t SenderInst) Protocol() string {
	return t.protocol
}

func (t SenderInst) RemoteAddr() string {
	return t.remote
}

func (t SenderInst) ID() ConnectionId {
	return t.id
}

func (t SenderInst) Send(ctx context.Context, raw []byte) error {
	return t.send(ctx, raw)
}

func GenerateOrderNo(prefix string) string {
	now := time.Now()
	year := fmt.Sprintf("%02d", now.Year()%100)
	month := fmt.Sprintf("%02d", int(now.Month()))
	day := fmt.Sprintf("%02d", now.Day())
	hour := fmt.Sprintf("%02d", now.Hour())
	minute := fmt.Sprintf("%02d", now.Minute())
	second := fmt.Sprintf("%02d", now.Second())

	rand.Seed(time.Now().UnixNano())
	randomNum := fmt.Sprintf("%02d", rand.Intn(90)+10)

	return fmt.Sprintf(prefix+"%s%s%s%s%s%s%s", year, month, day, hour, minute, second, randomNum)
}

func NumberToCol(num int) string {
	if num <= 0 {
		return ""
	}

	var result []byte
	for num > 0 {
		num--
		letter := byte((num % 26) + 'A')
		result = append([]byte{letter}, result...)
		num /= 26
	}

	return string(result)
}
