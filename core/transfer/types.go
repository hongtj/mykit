package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	. "mykit/core/dsp"
	. "mykit/core/types"
	"net"
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	probing "github.com/prometheus-community/pro-bing"
)

func CheckHandler(raw interface{}) {
	vt := reflect.TypeOf(raw)

	if vt.NumIn() != 1 || vt.NumOut() != 1 {
		msg := fmt.Sprintf("%v is invalid", vt.Name())
		panic(msg)
	}
}

type InvocationParam struct {
	Src     string            `json:"src"`
	Dst     string            `json:"dst"`
	App     string            `json:"app"`
	Method  string            `json:"method"`
	Trace   string            `json:"trace"`
	Meta    map[string]string `json:"meta"`
	Payload json.RawMessage   `json:"payload"`
}

func (t InvocationParam) Ctx() context.Context {
	m := map[string]string{
		TagSrc:    t.Src,
		TagDst:    t.Dst,
		TagTrace:  t.Trace,
		TagSpan:   defaultSpanStr,
		TagUser:   t.Meta[TagUser],
		TagApp:    t.App,
		TagMethod: t.Method,
	}

	return NewIncomingContext(m)
}

func (t InvocationParam) Invalid(raw ...string) bool {
	if t.Trace == "" || t.Src == "" {
		return true
	}

	return t.Dst != ParseStrParam(raw, INST())
}

func (t InvocationParam) ToSend() []byte {
	return MustJsonMarshal(t)
}

func (t InvocationParam) Req() *Req {
	res := &Req{}
	json.Unmarshal(t.Payload, res)

	return res
}

func (t InvocationParam) Res() *Res {
	res := &Res{}
	json.Unmarshal(t.Payload, res)

	return res
}

func (t *InvocationParam) BadRequestRes(code ...int32) {
	res := BadRequestRes(ParseInt32Param(code, -1))
	t.Payload = res.ToSend()
}

func (t *InvocationParam) TimeoutRes(method string) {
	res := TimeoutRes(method)
	t.Payload = res.ToSend()
}

func (t *InvocationParam) Register(wait time.Duration) {
	t.Src = INST()
	cc.Register(t.Trace, wait)
}

func (t InvocationParam) GetResponse(ctx context.Context) (rsp *InvocationParam, err error) {
	return cc.GetResponse(ctx, t.Trace, t.Method)
}

func (t InvocationParam) Submit() {
	cc.Submit(t)
}

type InvocationReq struct {
	Src   string            `json:"src"`
	Dst   string            `json:"dst"`
	Trace string            `json:"trace"`
	Meta  map[string]string `json:"meta"`
	*Req
}

func NewInvocationReq(ctx context.Context, src, dst string, req *Req) InvocationReq {
	traceId := GetTrace(ctx)
	if len(traceId) == 0 {
		traceId = GetSpanContext(ctx).TraceID().String()
	}

	meta := map[string]string{
		TagUser: GetUser(ctx),
	}

	res := InvocationReq{
		Src:   src,
		Dst:   dst,
		Trace: traceId,
		Meta:  meta,
		Req:   req,
	}

	return res
}

func (t InvocationReq) InvocationParam() InvocationParam {
	res := InvocationParam{
		Src:     t.Src,
		Dst:     t.Dst,
		App:     t.App,
		Method:  t.Method,
		Trace:   t.Trace,
		Meta:    t.Meta,
		Payload: t.ToSend(),
	}

	return res
}

func NewReq(app, method string, obj interface{}) *Req {
	res := &Req{
		App:    app,
		Method: method,
		Param:  MustJsonMarshal(obj),
	}

	return res
}

func UtilsReq(method string, obj interface{}) *Req {
	return NewReq(RpcUtils, method, obj)
}

func IotReq(method string, obj interface{}) *Req {
	return NewReq(RpcIot, method, obj)
}

func NewReqFromGin(c *gin.Context, app, method string) *Req {
	res := &Req{
		App:    app,
		Method: method,
		Param:  GetBodyFromGin(c),
	}

	return res
}

func (x *Req) ToSend() []byte {
	return MustJsonMarshal(x)
}

func (x *Req) Payload() []byte {
	res := x.GetParam()
	if len(res) > 0 {
		return res
	}

	return ByteOfNullJson
}

func AppNotImplRes(app string) *Res {
	res := &Res{
		Code: CodeUnimplemented,
		Msg:  fmt.Sprintf(appNotImpl, app),
		Data: ByteOfNullJson,
	}

	return res
}

func MethodNotImplRes(method string) *Res {
	res := &Res{
		Code: CodeUnimplemented,
		Msg:  fmt.Sprintf(methodNotImpl, method),
		Data: ByteOfNullJson,
	}

	return res
}

func BadRequestRes(code int32) *Res {
	res := &Res{
		Code: code,
		Msg:  RspMsgBadRequest,
		Data: ByteOfNullJson,
	}

	return res
}

func TimeoutRes(app string) *Res {
	res := &Res{
		Code: CodeDeadlineExceeded,
		Msg:  fmt.Sprintf(callTimeout, app),
		Data: ByteOfNullJson,
	}

	return res
}

func RpcFail(proc, target string, err error) *Res {
	ZapFailed(LogS1,
		LogEvent("error occur @ rpc"),
		LogProcessor(proc),
		LogError(err),
	)

	if TimeoutErr(err) {
		return TimeoutRes(target)
	}

	res := &Res{
		Code: CodeInternal,
		Msg:  fmt.Sprintf(callFailed, target),
		Data: ByteOfNullJson,
	}

	return res
}

func (x *Res) CloneTo(raw *Res) {
	raw.Code = x.GetCode()
	raw.Msg = x.GetMsg()
	raw.Data = x.GetData()
}

func (x *Res) Dump() string {
	res := fmt.Sprintf(`{"code": %v, "msg": "%v", "data": %v}`,
		x.GetCode(), x.GetMsg(), BytesToString(x.GetData()))

	return res
}

func (x *Res) ToSend() []byte {
	return MustJsonMarshal(x)
}

type FinalRsp struct {
	Code int32           `json:"Code"`
	Msg  string          `json:"Msg"`
	Data json.RawMessage `json:"Data"`
}

func NewFinalRsp(msg string, code ...int32) FinalRsp {
	res := FinalRsp{
		Code: ParseInt32Param(code, http.StatusOK),
		Msg:  msg,
		Data: ByteOfNullJson,
	}

	return res
}

func DumpFinalRsp(msg string, obj interface{}, code ...int32) FinalRsp {
	res := FinalRsp{
		Code: ParseInt32Param(code, http.StatusOK),
		Msg:  msg,
		Data: EnsureJsonByte(MustJsonMarshal(obj)),
	}

	return res
}

func ErrorCodeFinalRsp(obj interface{}, err *ErrorCode) FinalRsp {
	res := FinalRsp{
		Code: err.Code(),
		Msg:  err.Error(),
		Data: EnsureJsonByte(RespJsonMarshal(obj)),
	}

	return res
}

type FinalRsp2 struct {
	Code int32           `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func NewFinalRsp2(msg string, code ...int32) FinalRsp2 {
	res := FinalRsp2{
		Code: ParseInt32Param(code, http.StatusOK),
		Msg:  msg,
		Data: ByteOfNullJson,
	}

	return res
}

func DumpFinalRsp2(msg string, obj interface{}, code ...int32) FinalRsp2 {
	res := FinalRsp2{
		Code: ParseInt32Param(code, http.StatusOK),
		Msg:  msg,
		Data: EnsureJsonByte(RespJsonMarshal(obj)),
	}

	return res
}

func SuccessFinalRsp2(obj interface{}) FinalRsp2 {
	return DumpFinalRsp2(RspMsgSuccess, obj, http.StatusOK)
}

func ErrorCodeFinalRsp2(obj interface{}, err *ErrorCode) FinalRsp2 {
	res := FinalRsp2{
		Code: err.Code(),
		Msg:  err.Error(),
		Data: EnsureJsonByte(RespJsonMarshal(obj)),
	}

	return res
}

var defaultUdpProcessor UdpProcessor = func(addr *net.UDPAddr, raw []byte) ([]byte, error) {
	return raw, nil
}

type PingResponse struct {
	PingRes
	LinkErr    error
	PersistErr error
}

type PingRes struct {
	Start time.Time
	End   time.Time
	Cost  time.Duration
	*probing.Statistics
}
