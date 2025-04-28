package transfer

import (
	"context"
	. "mykit/core/dsp"
	. "mykit/core/types"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
	"go.opentelemetry.io/otel/trace"
)

var defaultCorsAllowMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodDelete,
	http.MethodOptions,
}

var corsAllowMethods = strings.Join(defaultCorsAllowMethods, HeaderDelimiter)

var defaultCorsAllowHeaders = []string{
	"Origin",
	ContentLength,
	ContentType,
	"Access-Control-Allow-Origin",
	"Access-Control-Allow-Headers",
	"X-Requested-With",
	"Accept",
	HeadSign,
	HeadXToken,
	HeadAToken,
	HeadLanguage,
	HeaderClient,
	HeadSession,
	HeadSecret,
	HeadNonce,
	HeadWxApp,
}

var corsAllowHeaders = strings.Join(defaultCorsAllowHeaders, HeaderDelimiter)

var defaultCorsExposeHeaders = []string{
	ContentLength,
	ContentLanguage,
	ContentType,
	"Access-Control-Allow-Origin",
	"Access-Control-Allow-Headers",
	"Cache-Control",
	"Expires",
	"Last-Modified",
	"Pragma",
	"FooBar",
	HeadSecret,
	HeadTrace,
	HeadNonce,
	HeadRole,
	HeadMenu,
}

var corsExposeHeaders = strings.Join(defaultCorsExposeHeaders, HeaderDelimiter)

func SetCorsAllowHeaders(raw ...string) {
	corsAllowHeaders = strings.Join(raw, ", ")
}

//CorsMiddleware : 处理跨域
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", corsAllowMethods)
		c.Header("Access-Control-Allow-Headers", corsAllowHeaders)
		c.Header("Access-Control-Expose-Headers", corsExposeHeaders)
		c.Header("Access-Control-Allow-Credentials", "false")
		c.Header(ContentType, JsonContentType)

		request := c.Request.Method
		if request == http.MethodGet || request == http.MethodPost {
			c.Next()
			return
		}

		if request == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.JSON(
			http.StatusOK,
			Res{Code: http.StatusMethodNotAllowed, Msg: "Permission Deny 0000", Data: ByteOfNullJson},
		)
		c.Abort()
	}
}

func CheckIp(validator MethodValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		doCheck(c, validator, ip)
	}
}

func OnlyLocal() gin.HandlerFunc {
	return CheckIp(SimpleValidator{"127.0.0.1": true})
}

func CheckMethod(validator MethodValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Param(TagMethod)
		doCheck(c, validator, method)
	}
}

func doCheck(c *gin.Context, validator MethodValidator, k string) {
	if !validator.IsValid(k) {
		c.String(http.StatusUnauthorized, "Unauthorized")
		Logger(c).Warnf("Unauthorized in %v，from %v!", c.FullPath(), c.Request.RemoteAddr)
		c.Abort()
		return
	}

	c.Next()
}

func SimpleToken(raw string) gin.HandlerFunc {
	credential := raw

	var res = func(c *gin.Context) {
		token := c.GetHeader(HeadXToken)
		if token != credential {
			c.JSON(
				http.StatusUnauthorized,
				NewFinalRsp("permission deny 0001", http.StatusUnauthorized),
			)
			c.Abort()
			return
		}

		c.Next()
	}

	return res
}

func SetFailedCode(c *gin.Context, code uint) {
	c.Set(LogFiledCode, int(code)*-1)
}

func GetUserAccount(c *gin.Context) string {
	return c.GetString(TagUser)
}

func GetUserCredential(c *gin.Context) string {
	return c.GetString(TagCredential)
}

func TraceIdFromGin(c *gin.Context) string {
	return c.GetString(TagTrace)
}

func SpanIdFromGin(c *gin.Context) string {
	return c.GetString(TagSpan)
}

func GinContext(c *gin.Context) context.Context {
	ctx := context.WithValue(Ctx, TagUser, c.GetString(TagUser))

	traceStr := c.GetString(TagTrace)
	if len(traceStr) > 0 {
		ctx = TracedContextFromStr(ctx, traceStr)
	}

	return ctx
}

func TraceMiddle(c *gin.Context) {
	propagator := otel.GetTextMapPropagator()
	tracer := otel.GetTracerProvider().Tracer(INST())
	ctx := propagator.Extract(c, propagation.HeaderCarrier(c.Request.Header))

	_, span := tracer.Start(
		ctx,
		INST(),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.HTTPServerAttributesFromHTTPRequest(
				INST(),
				c.Request.URL.Path,
				c.Request,
			)...,
		),
	)
	defer span.End()

	sc := span.SpanContext()
	traceId := sc.TraceID().String()
	c.Header(HeadTrace, traceId)

	c.Set(TagClient, GetClientFromGin(c))
	c.Set(TagUserAgent, GetUserAgentFromGin(c))
	c.Set(TagTrace, traceId)
	c.Set(TagSpan, sc.SpanID().String())
	c.Set(TagRemote, GetRemoteIP(c.Request))
	c.Set(TagScn, c.GetHeader(HeadScn))
	c.Set(TagLanguage, c.GetHeader(HeadLanguage))

	c.Next()
}

func GinHeartbeat(c *gin.Context) {
	res := map[string]int64{
		"tick": time.Now().UnixMilli(),
	}

	SendFinalRsp2(c, res, nil)
}

func SetHtmlHeader(h http.Header) {
	h.Set(ContentType, HtmlContentType)
}

func SetFileHeader(h http.Header, filename string) {
	h.Set(ContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	h.Set("Content-Disposition", "attachment; filename="+filename)
	h.Set("File-Name", filename)
}
