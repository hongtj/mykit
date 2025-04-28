package transfer

import (
	"context"
	"fmt"
	. "mykit/core/dsp"
	. "mykit/core/types"
	"net/http"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
)

var (
	rpcRetryTimes int32 = 3
)

func SetRpcRetryTimes(t int32) {
	atomic.StoreInt32(&rpcRetryTimes, t)
}

func GetRpcRetryTimes() int32 {
	return atomic.LoadInt32(&rpcRetryTimes)
}

func SmarterGrpcClient(address string) SmarterClient {
	conn, err := grpc.Dial(address,
		grpc.WithInsecure(),
	)
	HandleInitErr("SmarterGrpcClient", err)

	return SmarterGrpc(conn)
}

func SmarterCall(client SmarterClient, ctx context.Context, req *Req) (rsp *Res, err error) {
	c := RpcCtx(ctx)

	rsp, err = client.Call(c, req)
	if err != nil {
		msg := fmt.Sprintf("app [%v] call faild", req.App)
		rsp = &Res{Code: http.StatusInternalServerError, Msg: msg, Data: ByteOfNullJson}
		return
	}

	return
}

func RpcCall(client SmarterClient,
	latency time.Duration,
	req *Req, res ...interface{}) (err error) {

	return RpcCallWithTimeoutAndInterval(client, int(GetRpcRetryTimes()), latency, Interval100ms, req, res...)
}

func RpcCallWithTimeout(client SmarterClient,
	retry int,
	latency time.Duration,
	req *Req, res ...interface{}) (err error) {

	return RpcCallWithTimeoutAndInterval(client, retry, latency, Interval100ms, req, res...)
}

func RpcCallWithTimeoutAndInterval(client SmarterClient,
	retry int,
	latency time.Duration, interval TaskInterval,
	req *Req, res ...interface{}) (err error) {

	ctx, cancel := context.WithTimeout(Ctx, latency)
	defer cancel()

	var rsp *Res
	for i := 0; i < retry; i++ {
		select {
		case <-ctx.Done():
			return ErrorCodeTimeout
		default:
		}

		if rsp, err = SmarterCall(client, ctx, req); err == nil {
			if len(res) > 0 {
				err = UnmarshalJson(rsp.GetData(), &res[0])
			}
			return
		}

		time.Sleep(interval())
	}

	return err
}

type GrpcLogSender struct {
	address string
}

func (t GrpcLogSender) Send(raw []byte) (err error) {
	//todo:
	return
}

func NewGrpcLogSender(address string) (res LogSender, err error) {
	res = GrpcLogSender{
		address: address,
	}

	return
}
