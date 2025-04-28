package transfer

import (
	"context"
	. "mykit/core/dsp"
	. "mykit/core/internal"
	. "mykit/core/types"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var (
	initTraceOnce sync.Once
)

func InitTrace() {
	initTraceOnce.Do(func() {
		GlobalContext, GlobalCancel = context.WithCancel(context.Background())

		tp := traceSdk.NewTracerProvider(
			traceSdk.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNamespaceKey.String(Space()),
				semconv.HostNameKey.String(HOST()),
				semconv.ProcessPIDKey.Int(PID()),
				semconv.ServiceNameKey.String(INST()),
				semconv.ServiceVersionKey.String(Version()),
			)),
		)

		otel.SetTracerProvider(tp)
	})
}

func TraceFromStr(ctx context.Context, traceStr string, spanStr string) context.Context {
	if traceStr == "" || spanStr == "" {
		return ctx
	}

	spanContext, err := MakeSpan(traceStr, spanStr)
	if err != nil {
		return ctx
	}

	tracer := otel.GetTracerProvider().Tracer(Host())
	_, span := tracer.Start(
		oteltrace.ContextWithSpanContext(ctx, spanContext),
		INST(),
	)
	defer span.End()

	sc := span.SpanContext()
	ctx = context.WithValue(ctx, TagTrace, sc.TraceID().String())
	ctx = context.WithValue(ctx, TagSpan, sc.SpanID().String())

	return ctx
}

func GetTracedContext(c ...context.Context) context.Context {
	ctx := ParseContextParam(c)

	sc := GetSpanContext(ctx)
	ctx = context.WithValue(ctx, TagTrace, sc.TraceID().String())
	ctx = context.WithValue(ctx, TagSpan, sc.SpanID().String())

	return oteltrace.ContextWithSpanContext(ctx, sc)
}

func GetSpanContext(c ...context.Context) oteltrace.SpanContext {
	ctx := ParseContextParam(c)

	tracer := otel.GetTracerProvider().Tracer(Host())
	_, span := tracer.Start(
		ctx,
		INST(),
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
	)
	defer span.End()

	sc := span.SpanContext()

	return sc
}

func GetLongTracedContext() context.Context {
	tracer := otel.GetTracerProvider().Tracer(Host())
	ctx := context.Background()
	_, span := tracer.Start(
		ctx,
		INST(),
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
	)
	defer span.End()

	sc := span.SpanContext()
	traceId := LongTrace(sc.TraceID().String(), sc.SpanID().String())
	ctx = context.WithValue(ctx, TagTrace, traceId)

	return ctx
}

func GetLongTrace() string {
	tracer := otel.GetTracerProvider().Tracer(Host())
	ctx := context.Background()
	_, span := tracer.Start(
		ctx,
		INST(),
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
	)
	defer span.End()

	sc := span.SpanContext()
	traceId := LongTrace(sc.TraceID().String(), sc.SpanID().String())

	return traceId
}

func LongTrace(traceStr string, spanStr ...string) string {
	return "00-" + traceStr + "-" + ParseStrParam(spanStr, defaultSpanStr) + "-00"
}

func FromLongTrace(raw string) (trace, span string, err error) {
	tmp := strings.Split(raw, "-")
	if len(tmp) < 4 {
		err = ErrInvalidTraceParam
		return
	}

	trace = tmp[1]
	_, err = oteltrace.TraceIDFromHex(trace)
	if err != nil {
		return
	}

	span = tmp[2]
	_, err = oteltrace.SpanIDFromHex(span)

	return
}

func ParseContextTraced(param []context.Context) context.Context {
	if len(param) == 0 {
		return GetTracedContext()
	}

	return param[0]
}
