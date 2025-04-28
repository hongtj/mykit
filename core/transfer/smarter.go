package transfer

import (
	"context"
	"fmt"
	. "mykit/core/dsp"
	. "mykit/core/persist"
	. "mykit/core/types"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var AppDispatch = map[string]string{}

var BeforeSend = func(c *gin.Context) {}

func RegAppDispatch(service string, m ...string) {
	if len(m) == 0 {
		return
	}

	log := []string{}

	for _, v := range m {
		line := fmt.Sprintf("%v -> %v", v, service)
		log = append(log, line)

		AppDispatch[v] = service
	}

	fmt.Println(strings.Join(log, "\n"))
}

type SmarterRouter map[string]SmarterService

func (t SmarterRouter) Add(etcd EtcdConfig, app ...string) {
	for _, v := range app {
		t[v] = NewSmarterClient(v, etcd)
	}
}

func (t SmarterRouter) Rpc(ctx context.Context,
	app, method string, obj interface{},
	res ...interface{}) (code int32, err error) {

	req := NewReq(app, method, obj)
	c := RpcCtx(ctx)

	rsp, err := t.Call(c, req)
	if err != nil || rsp == nil {
		return
	}

	code = rsp.GetCode()

	if len(res) > 0 {
		err = UnmarshalJson(rsp.GetData(), &res[0])
	}

	return
}

func (t SmarterRouter) Call(ctx context.Context, req *Req) (rsp *Res, err error) {
	app := req.GetApp()

	target := AppDispatch[app]
	client, ok := t[target]
	if !ok {
		msg := fmt.Sprintf(appNotImpl, app)
		rsp = &Res{Code: http.StatusNotFound, Msg: msg, Data: ByteOfNullJson}
		return
	}

	rsp, err = client.Call(ctx, req)
	if err != nil {
		msg := fmt.Sprintf(callFailed, app)
		rsp = &Res{Code: http.StatusInternalServerError, Msg: msg, Data: ByteOfNullJson}
		return
	}

	return
}

func (t SmarterRouter) Handle(c *gin.Context, app, method string) {
	ctx := CallCtx(c)
	req := NewReqFromGin(c, app, method)
	rsp, err := t.Call(ctx, req)

	BeforeSend(c)

	SendRsp(c, req, rsp, err)
}

func (t SmarterRouter) GinHandler() func(c *gin.Context) {
	var h = func(c *gin.Context) {
		t.Handle(c, c.Param(TagApp), c.Param(TagMethod))
	}

	return h
}

func (t SmarterRouter) Handle2(c *gin.Context, app, method string) {
	ctx := CallCtx(c)
	req := NewReqFromGin(c, app, method)

	rsp, err := t.Call(ctx, req)

	BeforeSend(c)

	SendRsp2(c, req, rsp, err)
}

func (t SmarterRouter) GinHandler2() func(c *gin.Context) {
	var h = func(c *gin.Context) {
		t.Handle2(c, c.Param(TagApp), c.Param(TagMethod))
	}

	return h
}

func (t SmarterRouter) LpcCall(c *gin.Context, app, method string) {
	ctx := CallCtx(c)
	req := NewReqFromGin(c, app, method)

	rsp := &Res{}
	err := DISP.Call(ctx, req, rsp)
	if rsp.Code == CodeUnimplemented {
		rsp, err = t.Call(ctx, req)
	}

	BeforeSend(c)

	SendRsp2(c, req, rsp, err)
}

func (t SmarterRouter) LpcHandler() func(c *gin.Context) {
	var h = func(c *gin.Context) {
		t.LpcCall(c, c.Param(TagApp), c.Param(TagMethod))
	}

	return h
}
