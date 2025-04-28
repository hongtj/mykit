package dsp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	. "mykit/core/types"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	zhLocale "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhT "github.com/go-playground/validator/v10/translations/zh"
	"go.uber.org/zap/zapcore"
)

func Go(ctx context.Context, f JOB, msg ...string) {
	desc := ParseStrParam(msg, ShortCaller(1))

	go func() {
		defer Recover(desc)
		f(NewContext(ctx))
	}()
}

func Start(ctx context.Context, f JOB, msg ...string) {
	desc := ParseStrParam(msg, ShortCaller(1))

	s := desc
	PadSuffix(&s, " start")
	fmt.Println(s)

	go func() {
		defer Recover(desc)
		f(NewContext(ctx))
	}()
}

func Recover(desc string, outputPanicStack ...bool) {
	out := ParseBool(outputPanicStack)

	if r := recover(); r != nil {
		err, ok := r.(error)
		if !ok {
			err = fmt.Errorf("%v", r)
		}

		if out {
			buf := make([]byte, OutputPanicStackSize)
			buf = buf[:runtime.Stack(buf, false)]

			ZapFailed(LogS1,
				LogEvent(LogMsgRecover),
				LogProcessor(desc),
				LogBinary(buf),
				LogError(err),
			)

		} else {

			ZapFailed(LogS1,
				LogEvent(LogMsgRecover),
				LogProcessor(desc),
				LogError(err),
			)
		}
	}
}

func ShortCaller(k int) string {
	k++

	pc, file, line, ok := runtime.Caller(k)

	entry := zapcore.NewEntryCaller(pc, file, line, ok)

	return entry.TrimmedPath()
}

type Spokesman struct {
	msg string
	Speech
}

func NewSpokesman(msg string) *Spokesman {
	res := &Spokesman{
		msg: msg,
	}

	return res
}

func (t *Spokesman) SetSpeaker(f Speech) {
	t.Speech = f
}

func (t *Spokesman) DoSpeech(ctx context.Context) string {
	if t.Speech == nil {
		return t.msg
	}

	return t.Speech(ctx)
}

type JsonEncoder struct {
	l sync.RWMutex
	m map[string]reflect.Type
}

func NewJsonEncoder() *JsonEncoder {
	res := &JsonEncoder{
		m: map[string]reflect.Type{},
	}

	return res
}

func (t *JsonEncoder) Valid(raw []byte) bool {
	return json.Valid(raw)
}

func (t *JsonEncoder) Encode(raw interface{}) (res []byte, err error) {
	return json.Marshal(raw)
}

func (t *JsonEncoder) Decode(raw []byte, obj interface{}) (err error) {
	return json.Unmarshal(raw, obj)
}

func (t *JsonEncoder) Reg(k string, obj interface{}) error {
	t.l.RLock()
	_, ok := t.m[k]
	t.l.RUnlock()

	if ok {
		return ErrAlreadyExist
	}

	m := reflect.TypeOf(obj)

	t.l.Lock()
	t.m[k] = m
	t.l.Unlock()

	return nil
}

func (t *JsonEncoder) Load(k string, raw []byte) (obj interface{}, err error) {
	t.l.RLock()
	m, ok := t.m[k]
	t.l.RUnlock()

	if !ok {
		return raw, nil
	}

	obj = reflect.New(m).Interface()
	err = t.Decode(raw, obj)

	return
}

type FnvSelector struct {
	n int
}

func NewFnvSelector(n int) Selector {
	res := FnvSelector{
		n: n,
	}

	return res
}

func (t FnvSelector) Select(raw string) int {
	return Fnv32(raw) % t.n
}

type KV struct {
	K string
	V string
}

type PageCacheObj struct {
	K string
	O string
	V string
}

func (t IotDataPack) Payload() []byte {
	return MustJsonMarshal(t)
}

func (t ACCESS) Address() string {
	return fmt.Sprintf("%v:%v", t.Host, t.Port)
}

func (t *WatchCat) Set(n int64) {
	atomic.StoreInt64((*int64)(t), n)
}

func (t *WatchCat) Load() int64 {
	return atomic.LoadInt64((*int64)(t))
}

func (t *WatchCat) Closed() bool {
	return t.Load() == -9999
}

func (t *WatchCat) Help(raw ...uint16) {
	n := ParseUint16Param(raw, 0)
	n++

	t.Set(int64(n) * -1)
}

func (t *WatchCat) KeepAlive(f func(*WatchCat), d ...time.Duration) {
	scan := time.Millisecond * 200
	wait := ParseTimeDuration(d, time.Second*10)

	t.KeepAliveAdv(scan, wait, f)
}

func (t *WatchCat) KeepAliveAdv(scan, wait time.Duration, f func(*WatchCat)) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			time.Sleep(wait)
			t.KeepAliveAdv(scan, wait, f)
		}
	}()

start:
	fmt.Println("KeepAlive:", Fn(f).Name())

	t.Set(1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println(r)
				t.Help(99)
			}
		}()

		f(t)
	}()

	for {
		if t.Closed() {
			return
		}

		if t.Load() < 0 {
			time.Sleep(wait)
			goto start
		}

		CHECK(scan)
	}
}

func KeepAlive(f func(cat *WatchCat), d ...time.Duration) {
	cat := new(WatchCat)
	cat.KeepAlive(f, d...)
}

func (t BaseQuery) I64(raw int64) int64 {
	return DeInt64Param(t.Id, raw)
}

func (t BaseQuery) UUID(raw string) string {
	return DeStrParam(t.Uuid, raw)
}

func (t BaseMeta) Values() []interface{} {
	values := []interface{}{
		"id", t.Id,
		"name", t.Name,
		"description", t.Description,
		"status", t.Status,
		"sort", t.Sort,
	}

	return values
}

func NewBaseMeta(raw map[string]string) BaseMeta {
	statusInt, _ := ParseInt64FromStr(raw["status"])
	sort, _ := ParseInt64FromStr(raw["sort"])

	res := BaseMeta{
		Id:          raw["id"],
		Name:        raw["name"],
		Description: raw["description"],
		Status:      int32(statusInt),
		Sort:        sort,
	}

	return res
}

func RunUntilError(raw ...ErrorFunc) (err error) {
	for _, v := range raw {
		err = v()
		if err != nil {
			return
		}
	}

	return
}

func CtxRunUntilError(ctx context.Context, raw ...ErrorCtxFunc) (err error) {
	for _, v := range raw {
		err = v(ctx)
		if err != nil {
			return
		}
	}

	return
}

func GenPoint(station, kind string, n ...int64) string {
	return fmt.Sprintf("%v_%v_%d", station, kind, ParseInt64Param(n, 1))
}

var (
	validate = validator.New()
	transM   = map[string]ut.Translator{}
	transZh  ut.Translator
)

func ValidateStruct(s interface{}, l ...string) (err error) {
	err = validate.Struct(s)
	if err == nil {
		return nil
	}

	var errs validator.ValidationErrors
	ok := errors.As(err, &errs)
	if !ok {
		return err
	}

	trans := GetTrans(l...)

	b := NewStrBuilder()
	for _, v := range errs {
		b.WriteString(v.Translate(trans))
		b.WriteString(";")
	}

	err = errors.New(b.String())

	return err
}

func init() {
	initValidate()
}

func initValidate() {
	initTransLocal()

	zhT.RegisterDefaultTranslations(validate, transZh)
}

func initTransLocal() {
	initTransZh()
}

func initTransZh() {
	chT := zhLocale.New()
	uni := ut.New(chT)
	transZh, _ = uni.GetTranslator("zh")
	transM["zh"] = transZh
}

func GetTrans(l ...string) ut.Translator {
	language := "zh"
	if len(l) > 0 && l[0] != "" {
		language = strings.ToLower(l[0])
	}

	res, ok := transM[language]
	if !ok {
		res = transZh
	}

	return res
}

func ProcTag(raw string) string {
	if strings.HasPrefix(raw, "<") && strings.HasSuffix(raw, ">") {
		return raw
	}

	return strings.ToUpper("<" + raw + ">")
}

func AppLogger(ctx context.Context) *ZLogger {
	method := GetStringFromContext(ctx, TagMethod)
	if method == "" {
		method = INST()
	}

	return NewZLogger(ctx, method)
}

func Core(ctx context.Context) *CORE {
	res := &CORE{
		ZLogger: AppLogger(ctx),
	}

	return res
}

func (t *CORE) Setup(ctx context.Context, method string, param []byte) (err error) {
	return nil
}

func (t IotDataPack) Merge(raw ...IotDataPack) {
	for _, idp := range raw {
		for k, v := range idp {
			t[k] = v
		}
	}
}

func (t Bgd) Invalid() bool {
	return t.Company == 0 && t.Workshop == 0 && t.Scx == 0 && t.Team == 0 && t.Operator == 0
}

func OpenFile(f string) (*os.File, error) {
	return os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
}

func SampleRunning(s uint64, flag *uint64, pat string) {
	times := atomic.LoadUint64(flag)
	if (times-1)%s == 0 || STOPPED() {
		msg := fmt.Sprintf(pat, times)
		log.Println(msg)
	}
}

func NewTextTemplate(name, text string) *template.Template {
	return template.Must(template.New(name).Parse(text))
}

func GetTag(raw reflect.StructTag, tag ...string) string {
	for _, v := range tag {
		res, ok := raw.Lookup(v)
		if ok {
			return res
		}

	}

	return ""
}

func GetMacAddress(e ...string) (string, error) {
	eth := ParseStrParam(e, "eth0")
	ipCmd := exec.Command("ip", "link", "show", eth)

	var out bytes.Buffer
	ipCmd.Stdout = &out
	err := ipCmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run 'ip link show eth0': %w", err)
	}

	output := out.Bytes()

	for _, line := range bytes.Split(output, []byte("\n")) {
		if bytes.Contains(line, []byte("ether")) {
			parts := bytes.Fields(line)
			if len(parts) >= 2 {
				return string(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf("could not find MAC address in 'ip link show eth0' output")
}
