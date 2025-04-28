package transfer

import (
	"context"
	"sync"
	"time"
)

type Callback struct {
	l    sync.Mutex
	resp map[string]chan InvocationParam
	wait map[string]time.Duration
}

func NewCallback() *Callback {
	res := &Callback{
		resp: make(map[string]chan InvocationParam),
		wait: make(map[string]time.Duration),
	}

	return res
}

func (cc *Callback) Register(trace string, timeout time.Duration) chan InvocationParam {
	cc.l.Lock()
	defer cc.l.Unlock()

	res := make(chan InvocationParam, 1)
	cc.resp[trace] = res
	cc.wait[trace] = timeout

	return res
}

func (cc *Callback) Submit(resp InvocationParam) {
	cc.l.Lock()
	defer cc.l.Unlock()

	ch, ok := cc.resp[resp.Trace]
	if !ok {
		return
	}

	select {
	case ch <- resp:
		delete(cc.resp, resp.Trace)
		delete(cc.wait, resp.Trace)
	default:

	}
}

func (cc *Callback) GetResponse(ctx context.Context, trace, method string) (rsp *InvocationParam, err error) {
	rsp = &InvocationParam{Trace: trace}

	cc.l.Lock()
	respChan, ok := cc.resp[trace]
	timeoutDuration, _ := cc.wait[trace]
	cc.l.Unlock()

	if !ok {
		rsp.BadRequestRes(-1)
		return
	}

	select {
	case resp := <-respChan:
		rsp = &resp
		return

	case <-time.After(timeoutDuration):
		rsp.TimeoutRes(method)
		return

	case <-ctx.Done():
		rsp.TimeoutRes(method)
		return
	}
}
