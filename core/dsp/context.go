package dsp

import (
	"context"
	. "mykit/core/types"
	"time"

	oteltrace "go.opentelemetry.io/otel/trace"
)

// GetValue returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func GetValueFromContext(c context.Context, key string) (value interface{}, exists bool) {
	value = c.Value(key)
	return value, value != nil
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func MustGetValueFromContext(c context.Context, key string) interface{} {
	if value, exists := GetValueFromContext(c, key); exists {
		return value
	}

	msg := "Key \"" + key + "\" does not exist"
	panic(msg)
}

// GetString returns the value associated with the key as a string.
func GetStringFromContext(c context.Context, key string) (s string) {
	if val, ok := GetValueFromContext(c, key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool returns the value associated with the key as a boolean.
func GetBoolFromContext(c context.Context, key string) (b bool) {
	if val, ok := GetValueFromContext(c, key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt returns the value associated with the key as an integer.
func GetIntFromContext(c context.Context, key string) (i int) {
	if val, ok := GetValueFromContext(c, key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func GetInt64FromContext(c context.Context, key string) (i64 int64) {
	if val, ok := GetValueFromContext(c, key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetUint returns the value associated with the key as an unsigned integer.
func GetUintFromContext(c context.Context, key string) (ui uint) {
	if val, ok := GetValueFromContext(c, key); ok && val != nil {
		ui, _ = val.(uint)
	}
	return
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func GetUint64FromContext(c context.Context, key string) (ui64 uint64) {
	if val, ok := GetValueFromContext(c, key); ok && val != nil {
		ui64, _ = val.(uint64)
	}
	return
}

// GetFloat64 returns the value associated with the key as a float64.
func GetFloat64FromContext(c context.Context, key string) (f64 float64) {
	if val, ok := GetValueFromContext(c, key); ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}

// GetTime returns the value associated with the key as time.
func GetTimeFromContext(c context.Context, key string) (t time.Time) {
	if val, ok := GetValueFromContext(c, key); ok && val != nil {
		t, _ = val.(time.Time)
	}
	return
}

// GetDuration returns the value associated with the key as a duration.
func GetDurationFromContext(c context.Context, key string) (d time.Duration) {
	if val, ok := GetValueFromContext(c, key); ok && val != nil {
		d, _ = val.(time.Duration)
	}
	return
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func GetStringSliceFromContext(c context.Context, key string) (ss []string) {
	if val, ok := GetValueFromContext(c, key); ok && val != nil {
		ss, _ = val.([]string)
	}
	return
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func GetStringMapFromContext(c context.Context, key string) (sm map[string]interface{}) {
	if val, ok := GetValueFromContext(c, key); ok && val != nil {
		sm, _ = val.(map[string]interface{})
	}
	return
}

// GetStringMapString returns the value associated with the key as a map of strings.
func GetStringMapStringFromContext(c context.Context, key string) (sms map[string]string) {
	if val, ok := GetValueFromContext(c, key); ok && val != nil {
		sms, _ = val.(map[string]string)
	}
	return
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func GetStringMapStringSliceFromContext(c context.Context, key string) (smss map[string][]string) {
	if val, ok := GetValueFromContext(c, key); ok && val != nil {
		smss, _ = val.(map[string][]string)
	}
	return
}

func NewContext(c context.Context) context.Context {
	ctx := context.Background()
	span := oteltrace.SpanFromContext(c)
	res := oteltrace.ContextWithSpan(ctx, span)

	return res
}

func GetClient(ctx context.Context) string {
	return GetStringFromContext(ctx, TagClient)
}

func GetUserAgent(ctx context.Context) string {
	return GetStringFromContext(ctx, TagUserAgent)
}

func GetFrom(ctx context.Context) string {
	return GetStringFromContext(ctx, TagFrom)
}

func GetRemote(ctx context.Context) string {
	return GetStringFromContext(ctx, TagRemote)
}

func GetScc(ctx context.Context) string {
	return GetStringFromContext(ctx, TagScc)
}

func GetScn(ctx context.Context) string {
	return GetStringFromContext(ctx, TagScn)
}

func GetUser(ctx context.Context) string {
	return GetStringFromContext(ctx, TagUser)
}

func GetSession(ctx context.Context) string {
	return GetStringFromContext(ctx, TagSession)
}

func GetCredential(ctx context.Context) string {
	return GetStringFromContext(ctx, TagCredential)
}

func GetLanguage(ctx context.Context) string {
	return GetStringFromContext(ctx, TagLanguage)
}

func GetMethod(ctx context.Context) string {
	return GetStringFromContext(ctx, TagMethod)
}

func GetTenant(ctx context.Context) string {
	return GetStringFromContext(ctx, TagTenant)
}

func GetTrace(ctx context.Context) string {
	return GetStringFromContext(ctx, TagTrace)
}

func GetSpan(ctx context.Context) string {
	return GetStringFromContext(ctx, TagSpan)
}

func GetTime(ctx context.Context) time.Time {
	return GetTimeFromContext(ctx, TagTime)
}

func GetMethodParam(ctx context.Context, param string) string {
	return DeStrParam(GetMethod(ctx), param)
}

func SpanIdFromContext(ctx context.Context) string {
	spanCtx := oteltrace.SpanContextFromContext(ctx)
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}

	return ""
}

func TraceIdFromContext(ctx context.Context) string {
	spanCtx := oteltrace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}

	return ""
}

func TracedContextFromStr(ctx context.Context, traceStr string, spanStr ...string) context.Context {
	spanContext, err := MakeSpan(traceStr, ParseStrParam(spanStr, DefaultSpanStr))
	if err != nil {
		return ctx
	}

	return oteltrace.ContextWithSpanContext(ctx, spanContext)
}

func MakeSpan(traceStr, spanStr string) (span oteltrace.SpanContext, err error) {
	traceId, err := oteltrace.TraceIDFromHex(traceStr)
	if err != nil {
		return
	}

	spanId, err := oteltrace.SpanIDFromHex(spanStr)
	if err != nil {
		return
	}

	spanConf := oteltrace.SpanContextConfig{
		TraceID: traceId,
		SpanID:  spanId,
	}

	span = oteltrace.NewSpanContext(spanConf)

	return
}

func SetCtxTenant(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, TagTenant, tenant)
}
