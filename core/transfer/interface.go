package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"mykit/core/dsp"
	. "mykit/core/types"
	"net"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	MethodConfirmNode = "confirmNode"
	MethodNodeStatus  = "nodeStatus"
)

type TransNode struct {
	Node       string `json:",optional"`
	NodeAddr   string `json:",optional"`
	NodeAccess string `json:",optional"`
}

func (t *TransNode) CheckNode() {
	if ExistNullStr(t.Node, t.NodeAddr) {
		msg := "must set node"
		HandleInitErr(msg, ErrInvalidParam)
	}

	tcp := ParseTcp(&t.NodeAddr)

	if t.NodeAccess == "" {
		t.NodeAccess = fmt.Sprintf("%v:%d", GetPublic(), tcp.Port)
	}
}

type Disp func(Caller, context.Context, *Req) (res *Res, err error)

type Caller func(ctx context.Context, req *Req) (res *Res, err error)

type GinPluginLogger func(c *gin.Context, path string, start time.Time, cost time.Duration)

type Handler interface {
	New(ctx context.Context) Handler
}

type BaseHandler struct {
	*dsp.CORE
}

type GinHandlerMgr map[string]map[string]func(ctx *gin.Context)

func (t GinHandlerMgr) Get(app, method string) func(ctx *gin.Context) {
	m, ok := t[app]
	if !ok {
		return nil
	}

	f := m[method]

	return f
}

type MethodValidator interface {
	IsValid(method string) bool
}

type Worker func(context.Context, []byte) ([]byte, error)

type UdpProcessor func(addr *net.UDPAddr, raw []byte) ([]byte, error)

type UdpSender func(conn *net.UDPConn, addr *net.UDPAddr, resp []byte)

type UdpOnError func(conn *net.UDPConn, addr *net.UDPAddr, resp []byte, err error)

type Station struct {
	Name      string
	Ip        string
	Station   string
	PingPoint func(n ...int64) string
	RttPoint  func(n ...int64) string
}

type PingJob struct {
	Ctx    context.Context
	Point  string
	Target string
	Res    chan PingResponse
}

type RewritePrefix string

type RewriteDisp struct {
	Prefix RewritePrefix
	Disp   map[string]string
}

type JsonCallRes struct {
	Code int32           `json:"code"`
	Data json.RawMessage `json:"data"`
}
