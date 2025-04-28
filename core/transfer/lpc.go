package transfer

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"
	. "utils/dsp"
	. "utils/persist"
	. "utils/types"

	"go.uber.org/zap"
)

type LpcDispatch map[string]*Lpc

var (
	DISP = LpcDispatch{}
)

func (t LpcDispatch) getLpc(app string) *Lpc {
	lpc, ok := t[app]
	if !ok || lpc == nil {
		lpc = NewLpc(app)
		t[app] = lpc
	}

	return lpc
}

func (t LpcDispatch) Add(app string, f ...Handler) {
	lpc := t.getLpc(app)

	for _, v := range f {
		lpc.add(1, v)
	}
}

func (t LpcDispatch) Override(app string, f ...Handler) {
	lpc := t.getLpc(app)

	for _, v := range f {
		lpc.override(1, v)
	}
}

func (t LpcDispatch) Impact(app string, imp ...Impactor) {
	lpc := t.getLpc(app)

	lpc.Impact(imp...)
}

func (t LpcDispatch) Forbid(app string, method ...string) {
	lpc := t.getLpc(app)

	lpc.Forbid(method...)
}

func (t LpcDispatch) ForbidApp(app ...string) {
	for _, v := range app {
		delete(t, v)
	}
}

func (t LpcDispatch) Call(ctx context.Context, input *Req, output *Res) (err error) {
	ctx = HandleMd(ctx)
	ctx = context.WithValue(ctx, TagMethod, input.GetMethod())

	f, ok := GetProxy(input)
	if ok {
		proxyRes, _ := callProxy(f, ctx, input)
		proxyRes.CloneTo(output)
		return
	}

	app := input.GetApp()

	lpc, ok := t[app]
	if !ok {
		AppNotImplRes(app).CloneTo(output)
		return
	}

	lpcRes, _ := lpcCall(lpc, ctx, input)
	lpcRes.CloneTo(output)

	return
}

func (t LpcDispatch) Log() {
	app := []string{}
	for k := range t {
		app = append(app, k)
	}
	sort.Strings(app)

	iLog := []string{}
	for _, v := range app {
		a, o := t[v].Abstract()

		LogS1.Info(LogMsgSetup,
			LogEvent("lpc"),
			LogProcessor(v),
			LogDetail(map[string]interface{}{
				logImport:   a,
				logOverride: o,
			}),
		)

		for _, v2 := range a {
			iLog = append(iLog, ImportMsg(0, v, v2))
		}

		for _, v2 := range o {
			iLog = append(iLog, ImportMsg(1, v, v2))
		}
	}

	if len(iLog) > 0 {
		fmt.Println(strings.Join(iLog, "\n"))
	}
}

func (t LpcDispatch) IgnoreLog(app string, f ...string) {
	lpc, ok := t[app]
	if !ok || lpc == nil {
		return
	}

	lpc.SetLogOut(-9, f...)
}

func (t LpcDispatch) IgnoreLogData(app string, f ...string) {
	lpc, ok := t[app]
	if !ok || lpc == nil {
		return
	}

	lpc.SetLogOut(-1, f...)
}

func (t LpcDispatch) Method() map[string][]string {
	res := map[string][]string{}

	for k, v := range t {
		m := v.Method()
		res[k] = append(res[k], m...)
	}

	return res
}

type LpcMeta struct {
	o     int
	n     int
	param reflect.Type
	out   int32
	name  string
	Handler
}

func (t LpcMeta) ImportMsg(app string) string {
	return ImportMsg(t.o, app, t.name)
}

func ParseLpcMeta(h Handler, method ...string) (res []LpcMeta) {
	ht := reflect.TypeOf(h)
	hv := reflect.ValueOf(h)

	n := hv.NumMethod()
	m := map[string]int{}

	for i := 0; i < n; i++ {
		name := FirstLower(ht.Method(i).Name)
		m[name] = i
	}

	for _, v := range method {
		i, ok := m[v]
		if !ok {
			msg := fmt.Sprintf(unknownMethod, method)
			fmt.Println(msg)
			continue
		}

		mt := hv.Method(i).Type()
		code := DefaultMethodCheck(mt)
		if code > 0 {
			msg := fmt.Sprintf(invalidMethod, method, code)
			fmt.Println(msg)
			continue
		}

		item := LpcMeta{
			name:    v,
			n:       n,
			Handler: h,
			param:   mt.In(1).Elem(),
			out:     10,
		}

		res = append(res, item)
	}

	return
}

func ParseHandlerMethod(h Handler) (res []string) {
	n := reflect.ValueOf(h).NumMethod()

	for i := 0; i < n; i++ {
		meta := parseLpcMeta(h, i)
		if meta.name == "" {
			continue
		}

		res = append(res, meta.name)
	}

	return
}

func GetMethodList(m ...interface{}) []string {
	res := []string{}

	for _, v := range m {
		switch v.(type) {
		case Handler:
			h := v.(Handler)
			res = append(res, ParseHandlerMethod(h)...)

		case string:
			res = append(res, v.(string))

		default:
			res = append(res, RpcMethodName(v))
		}
	}

	return res
}

type Impactor struct {
	H Handler
	M []string
}

var (
	lpcCall        func(lpc *Lpc, ctx context.Context, req *Req) (res *Res, err error)
	parseDecodeErr = func(payload []byte, err error) string {
		return fmt.Sprintf(decodeErr, len(payload))
	}
)

var (
	DefaultMethodCheck = func(m reflect.Type) (code int32) {
		//校验入参出参数量
		code++
		if m.NumIn() != 2 || m.NumOut() != 2 {
			return
		}

		var t reflect.Type

		//校验入参1
		code++
		t = m.In(0)
		if !IsContextKind(t) {
			return
		}

		//校验入参2
		code++
		t = m.In(1)

		if !validInput(t) {
			return
		}

		//校验出参1
		code++
		t = m.Out(0)
		if !validOutput(t) {
			return
		}

		//校验出参2
		code++
		t = m.Out(1)
		if !IsErrorKind(t) {
			return
		}

		return 0
	}
)

func validInput(t reflect.Type) bool {
	if IsStructPtrKind(t) {
		return true
	}

	_, ok := reflect.New(t).Interface().(**DbContext)

	return ok
}

func validOutput(t reflect.Type) bool {
	return IsStructPtrKind(t)
}

func parseLpcMeta(h Handler, n int) (meta LpcMeta) {
	ht := reflect.TypeOf(h)
	hv := reflect.ValueOf(h)

	name := FirstLower(ht.Method(n).Name)
	method := hv.Method(n).Type()

	code := DefaultMethodCheck(method)
	if code > 0 {
		return
	}

	meta = LpcMeta{
		name:    name,
		n:       n,
		Handler: h,
		param:   method.In(1).Elem(),
		out:     10,
	}

	return
}

func InitLpc(release bool) {
	if release {
		lpcCall = callRelease

	} else {

		lpcCall = callDev
		parseDecodeErr = func(payload []byte, err error) string {
			return err.Error()
		}
	}
}

type Lpc struct {
	app string
	f   map[string]LpcMeta
}

func NewLpc(app string) *Lpc {
	res := &Lpc{
		app: app,
		f:   map[string]LpcMeta{},
	}

	return res
}

func (t *Lpc) Method() []string {
	m := []string{}
	for k := range t.f {
		m = append(m, k)
	}
	sort.Strings(m)

	return m
}

func (t *Lpc) Abstract() (a, o []string) {
	a = []string{}
	o = []string{}

	m := t.Method()

	for _, v := range m {
		meta := t.f[v]
		if meta.o == 0 {
			a = append(a, meta.name)
		} else {
			o = append(o, meta.name)
		}
	}

	return
}

func (t *Lpc) Add(h ...Handler) {
	for _, v := range h {
		t.add(1, v)
	}
}

func (t *Lpc) add(k int, h Handler) {
	k++

	n := reflect.ValueOf(h).NumMethod()
	for i := 0; i < n; i++ {
		t.addMethod(k, h, i)
	}
}

func (t *Lpc) AddMethod(h Handler, n int) {
	t.addMethod(1, h, n)
}

func (t *Lpc) addMethod(k int, h Handler, n int) (res string) {
	k++

	meta := t.parseMeta(k, h, n)
	if meta.name == "" {
		return
	}

	return t.Reg(meta)
}

func (t *Lpc) parseMeta(k int, h Handler, n int) LpcMeta {
	k++

	ht := reflect.TypeOf(h)
	hv := reflect.ValueOf(h)

	name := FirstLower(ht.Method(n).Name)
	method := hv.Method(n).Type()

	code := DefaultMethodCheck(method)
	if code > 0 {
		if code > 1 {
			detail := map[string]interface{}{
				TagApp:      t.app,
				TagMethod:   name,
				"check res": code,
			}

			Logger(Ctx).Skip(k).Info(LogMsgSetup,
				LogEvent("lpc"),
				LogProcessor("parse meta"),
				LogDetail(detail),
			)
		}

		return LpcMeta{}
	}

	meta := LpcMeta{
		name:    name,
		n:       n,
		Handler: h,
		param:   method.In(1).Elem(),
		out:     10,
	}

	return meta
}

func (t *Lpc) Reg(meta LpcMeta) (res string) {
	if meta.name == "" {
		msg := fmt.Sprintf(invalidLpcMeta, t.app)
		HandleInitErr(msg, ErrInvalidParam)
	}

	_, ok := t.f[meta.name]
	if ok {
		msg := fmt.Sprintf(dupedLpcMeta, meta.name, t.app)
		HandleInitErr(msg, ErrInvalidParam)
	}

	t.f[meta.name] = meta

	return meta.ImportMsg(t.app)
}

func (t *Lpc) Override(meta ...LpcMeta) {
	for _, v := range meta {
		if v.name == "" {
			continue
		}

		old, ok := t.f[v.name]
		if ok {
			v.o = old.o + 1
		}

		t.f[v.name] = v
	}
}

func (t *Lpc) override(k int, h Handler) {
	k++

	n := reflect.ValueOf(h).NumMethod()
	for i := 0; i < n; i++ {
		meta := t.parseMeta(k, h, i)
		t.Override(meta)
	}
}

func (t *Lpc) Impact(imp ...Impactor) {
	for _, v := range imp {
		meta := ParseLpcMeta(v.H, v.M...)
		t.Override(meta...)
	}
}

func (t *Lpc) Forbid(method ...string) {
	for _, v := range method {
		delete(t.f, v)
	}
}

func (t *Lpc) IgnoreLog(f ...string) {
	t.SetLogOut(0, f...)
}

func (t *Lpc) SetLogOut(n int32, f ...string) {
	for _, v := range f {
		meta, ok := t.f[v]
		if ok {
			meta.out = n
			t.f[v] = meta
		}
	}
}

func (t *Lpc) GetMethod(ctx context.Context, method string) (
	f reflect.Value, in reflect.Type, ok bool) {
	fn, ok := t.f[method]
	if !ok {
		return reflect.Value{}, nil, false
	}

	h := fn.New(ctx)
	f = reflect.ValueOf(h).Method(fn.n)

	in = fn.param

	return
}

func (t *Lpc) Call(ctx context.Context, req *Req) (res *Res, err error) {
	code, msg, data, err := t.call(ctx, req)

	res = &Res{Code: code, Msg: msg, Data: EnsureJsonByte(data)}

	return
}

func (t *Lpc) call(ctx context.Context, req *Req) (
	code int32, msg string, data []byte, err error) {
	method := req.GetMethod()
	f, in, ok := t.GetMethod(ctx, method)
	if !ok {
		code = CodeUnimplemented
		msg = fmt.Sprintf(methodNotImpl, method)
		return
	}

	//解析入参
	payload := req.Payload()
	param := reflect.New(in).Interface()
	err = UnmarshalJson(payload, param)
	if err != nil {
		code = CodeInvalidArgument
		msg = parseDecodeErr(payload, err)
		return
	}

	//校验入参
	if in.Kind() == reflect.Struct {
		err = ValidateStruct(param)
		if err != nil {
			code = CodeFailedOnRequired
			msg = err.Error()
			return
		}
	}

	//执行方法
	callRes := f.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(param),
	})

	//处理执行结果
	err, ok = callRes[1].Interface().(error)
	if ok && err != nil {
		var v *ErrorCode
		v, ok = err.(*ErrorCode)
		if ok {
			code = v.Code()
			msg = v.Error()
			data = RespJsonMarshalValue(callRes[0])
		} else {
			code = CodeInternal
			msg = fmt.Sprintf(callFailed, method)
		}

		return
	}

	code = http.StatusOK
	msg = RspMsgSuccess
	data = RespJsonMarshalValue(callRes[0])

	return
}

func callDev(lpc *Lpc, ctx context.Context, req *Req) (res *Res, err error) {
	var cost time.Duration

	method := req.GetMethod()
	out := lpc.f[method].out
	user := GetUser(ctx)
	from := GetFrom(ctx)

	f := []zap.Field{
		LogEventGrpc(),
	}

	defer func() {
		failed := false
		var panicErr error
		if r := recover(); r != nil {
			failed = true
			var ok bool
			panicErr, ok = r.(error)
			if !ok {
				panicErr = fmt.Errorf("%v", r)
			}

			res.Code = CodeInternal
			res.Msg = fmt.Sprintf(callFailed, method)
		}

		recordRequest(req, res, cost)

		if failed {
			buf := make([]byte, OutputPanicStackSize)
			buf = buf[:runtime.Stack(buf, false)]

			Logger(ctx).Failed(
				LogEvent("panic occur @ lpc"),
				LogProcessor(method),
				LogUser(user), LogFrom(from),
				LogBinary(buf),
				LogError(panicErr),
			)
			return
		}

		if out > 0 {
			f = append(f, LogBinary(res.GetData()))
		}

		if cost > rpcQuerySlowThreshold {
			f = append(f, LogPerformanceSlow())
		}

		Logger(ctx).Output(method, err, f...)
	}()

	Logger(ctx).Info(method,
		append(f,
			LogProcRecv(),
			LogUser(user), LogFrom(from),
			LogBinary(req.GetParam()),
		)...,
	)

	t0 := time.Now()
	res, err = lpc.Call(ctx, req)
	cost = time.Now().Sub(t0)

	f = append(f,
		LogProcSend(),
		LogContent(res.GetMsg()),
		LogDuration(cost),
		LogCode(res.GetCode()),
		LogUser(user), LogFrom(from),
	)

	return
}

func callRelease(lpc *Lpc, ctx context.Context, req *Req) (res *Res, err error) {
	var cost time.Duration

	method := req.GetMethod()
	out := lpc.f[method].out
	user := GetUser(ctx)
	from := GetFrom(ctx)

	f := []zap.Field{
		LogEventGrpc(),
	}

	defer func() {
		failed := false
		var panicErr error
		if r := recover(); r != nil {
			failed = true
			var ok bool
			panicErr, ok = r.(error)
			if !ok {
				panicErr = fmt.Errorf("%v", r)
			}

			res.Code = CodeInternal
			res.Msg = fmt.Sprintf(callFailed, method)
		}

		recordRequest(req, res, cost)

		if failed {
			buf := make([]byte, OutputPanicStackSize)
			buf = buf[:runtime.Stack(buf, false)]

			Logger(ctx).Failed(
				LogEvent("panic occur @ lpc"),
				LogProcessor(method),
				LogUser(user), LogFrom(from),
				LogBinary(buf),
				LogError(panicErr),
			)
			return
		}

		if out > 0 {
			f = append(f, LogBinary(res.GetData()))
		}

		if cost > rpcQuerySlowThreshold {
			f = append(f, LogPerformanceSlow())
		}

		Logger(ctx).Output(method, err, f...)
	}()

	if out > 0 {
		Logger(ctx).Info(method,
			append(f,
				LogProcRecv(),
				LogUser(user), LogFrom(from),
				LogBinary(req.GetParam()),
			)...,
		)
	}

	t0 := time.Now()
	res, err = lpc.Call(ctx, req)
	cost = time.Now().Sub(t0)

	if err != nil {
		//todo: 处理msg
		if v, ok := err.(*ErrorCode); ok {
			_ = v

		} else {

		}
	}

	f = append(f,
		LogProcSend(),
		LogContent(res.GetMsg()),
		LogDuration(cost),
		LogCode(res.GetCode()),
		LogUser(user), LogFrom(from),
	)

	return
}
