package transfer

import (
	"context"
	. "mykit/core/dsp"
	. "mykit/core/internal"
	. "mykit/core/types"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

var (
	p1 = []string{HeaderClient, HeaderUa, HeadFrom, HeadScn, HeadUser, HeadCredential, HeadLanguage}
)

func MetaFromGin(c *gin.Context) map[string]string {
	res := map[string]string{
		TagClient:     GetClientFromGin(c),
		TagUserAgent:  GetUserAgentFromGin(c),
		TagFrom:       c.GetString(TagRemote),
		TagScn:        GetScn(c),
		TagTrace:      TraceIdFromGin(c),
		TagSpan:       SpanIdFromGin(c),
		TagUser:       GetUserAccount(c),
		TagCredential: GetUserCredential(c),
		TagLanguage:   c.GetString(TagLanguage),
	}

	return res
}

func CallCtx(c *gin.Context) context.Context {
	md := MetaFromGin(c)
	return NewMetaContext(c, md)
}

func RpcCtx(ctx context.Context) context.Context {
	md := map[string]string{
		TagClient:    PidStr(),
		TagUserAgent: "smarter rpc",
		TagFrom:      INST(),
		TagUser:      "@#smarter#@",
	}

	trace := GetTrace(ctx)
	if len(trace) > 0 {
		md[TagTrace] = trace
		md[TagSpan] = GetSpan(ctx)
	}

	if md[TagTrace] == "" {
		c := GetSpanContext()
		md[TagTrace] = c.TraceID().String()
		md[TagSpan] = c.SpanID().String()
	}

	md[TagLanguage] = GetLanguage(ctx)

	return NewMetaContext(ctx, md)
}

func NewIncomingContext(raw map[string]string) context.Context {
	md := metadata.MD{}
	for k, v := range raw {
		md.Append(k, v)
	}

	return metadata.NewIncomingContext(Ctx, md)
}

func HandleMd(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		metas, ok2 := MetaFromContext(ctx)
		if !ok2 {
			return ctx
		}

		for _, v := range p1 {
			ctx = context.WithValue(ctx, FirstLower(v), metas[v])
		}

		return ctx
	}

	var traceStr string
	var spanStr string

	for k, v := range md {
		if len(v) > 0 && v[0] != "" {
			if k == TagTrace {
				traceStr = v[0]
			} else if k == TagSpan {
				spanStr = v[0]
			} else {
				ctx = context.WithValue(ctx, k, v[0])
			}
		}
	}

	ctx = TraceFromStr(ctx, traceStr, spanStr)

	ctx = context.WithValue(ctx, TagTime, time.Now())

	return ctx
}
