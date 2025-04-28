package transfer

import (
	"fmt"
	"io/ioutil"
	. "mykit/core/dsp"
	. "mykit/core/types"
	"net/http"
	"net/http/httputil"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GinConfig struct {
	Uri        string //0.0.0.0:8888
	Tls        bool   `json:",default=false"`
	Cert       string `json:",default=server.crt"` //HTTPS Cert
	Key        string `json:",default=server.key"` //HTTPS Key
	Host       string `json:",optional"`           //Host
	Gzip       int    `json:",optional"`           //Gzip
	StaticPath string `json:",default=/assets"`    //Static Path
	StaticDir  string `json:",default=./assets"`   //Static Dir
	address    string `json:",optional"`
}

func (t *GinConfig) Address() string {
	if t.address == "" {
		t.CheckUri()
	}

	return t.address
}

func (t *GinConfig) CheckUri() {
	uri := t.Uri
	tmp := strings.Split(uri, ":")

	host := GetIp()
	if t.Host != "" {
		host = t.Host
	}

	if len(tmp) > 1 {
		t.address = fmt.Sprintf("%v:%v", host, tmp[1])
	}

	ParseTcp(&t.address)
}

func (t GinConfig) NewServer(release bool, middleware ...gin.HandlerFunc) *GinServer {
	return NewGinServer(t, release, middleware...)
}

type GinServer struct {
	GinConfig
	*gin.Engine
}

func NewGinServer(conf GinConfig, release bool, middleware ...gin.HandlerFunc) *GinServer {
	if release {
		gin.SetMode(gin.ReleaseMode)
	}

	res := &GinServer{
		GinConfig: conf,
		Engine:    gin.New(),
	}

	res.Address()

	if conf.Gzip != 0 {
		gzipMiddleware := gzip.Gzip(conf.Gzip)
		middleware = append(middleware, gzipMiddleware)
	}

	InitGin(res.Engine, release, middleware...)

	return res
}

func (t *GinServer) NewServer() *http.Server {
	srv := &http.Server{
		Addr:        t.Uri,
		Handler:     t,
		ReadTimeout: time.Second * 10,
	}

	return srv
}

func (t *GinServer) Reg(router ...func(e *gin.Engine)) *GinServer {
	RegRouter(t.Engine, router...)

	return t
}

func (t *GinServer) Static() func(e *gin.Engine) {
	var res = func(e *gin.Engine) {
		e.Static(t.StaticPath, t.StaticDir)
	}

	return res
}

func (t *GinServer) UseStatic() {
	t.Static()(t.Engine)
}

func (t *GinServer) Run(tls ...bool) {
	useHttps := ParseBoolParam(tls, t.Tls)

	if !useHttps {
		fmt.Printf("server run @ http://%v\n", t.Uri)
		t.RunHttp()
		return
	}

	fmt.Printf("server run @ https://%v\n", t.Uri)
	t.RunTLS()
}

func (t *GinServer) RunHttp() {
	defer Recover("run http")

	err := t.NewServer().ListenAndServe()
	HandleInitErr("run gin http", err)
}

func (t *GinServer) RunTLS() {
	defer Recover("run tls error:")

	err := t.NewServer().ListenAndServeTLS(t.Cert, t.Key)
	HandleInitErr("run gin tls", err)
}

type GinProxy struct {
	c *gin.Context
	p *httputil.ReverseProxy
}

func (t *GinProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.p.ServeHTTP(w, r)

	t.c.Set(LogFiledEvent, LogMsgProxy)
	t.c.Set(LogFiledCode, t.c.Writer.Status())
}

func NewGinProxy(c *gin.Context, target *url.URL, m ...func(r *http.Request)) *GinProxy {
	res := &GinProxy{
		c: c,
		p: NewReverseProxy(target, m...),
	}

	return res
}

func RewriteGin(c *gin.Context, prefix string, target *url.URL) {
	proxy := NewGinProxy(c, target, ReplaceUrl(prefix))
	proxy.ServeHTTP(c.Writer, c.Request)
}

func InitGin(e *gin.Engine, release bool, middleware ...gin.HandlerFunc) {
	e.Use(
		CorsMiddleware(),
		GinLogger(),
		TraceMiddle,
	)

	e.Use(middleware...)

	if release {
		e.Use(gin.CustomRecoveryWithWriter(gin.DefaultErrorWriter, handleRecoveryForRelease))
	} else {
		e.Use(gin.CustomRecoveryWithWriter(gin.DefaultErrorWriter, handleRecoveryForDebug))
	}
}

func RegRouter(e *gin.Engine, router ...func(*gin.Engine)) {
	for _, v := range router {
		v(e)
	}
}

func RegGroup(g *gin.RouterGroup, handler ...func(*gin.RouterGroup)) {
	for _, v := range handler {
		v(g)
	}
}

func handleRecoveryForDebug(c *gin.Context, e interface{}) {
	path := c.Request.URL.Path

	err, ok := e.(error)
	if !ok {
		err = fmt.Errorf("%v", e)
	}

	data := MustJsonMarshal(
		NewFinalRsp(err.Error(), http.StatusInternalServerError),
	)

	Logger(c).Failed(
		LogEvent("panic occur @ gin"),
		LogProcessor(path),
		LogError(err),
	)

	c.Data(http.StatusInternalServerError,
		JsonContentType,
		EnsureJsonByte(data),
	)
}

func handleRecoveryForRelease(c *gin.Context, e interface{}) {
	path := c.Request.URL.Path
	err := e.(error)

	data := MustJsonMarshal(NewFinalRsp(path, http.StatusInternalServerError))

	Logger(c).Failed(
		LogEvent("panic occur @ gin"),
		LogProcessor(path),
		LogError(err),
	)

	c.Data(http.StatusInternalServerError,
		JsonContentType,
		EnsureJsonByte(data),
	)
}

func SendRsp(c *gin.Context, req *Req, rsp *Res, err error) {
	code := rsp.GetCode()
	msg := rsp.GetMsg()

	c.Set(LogFiledCode, int(code))
	c.Set(LogFiledMsg, msg)

	if err == nil {
		//StatusMethodNotAllowed
		c.JSON(
			http.StatusOK,
			FinalRsp{
				Code: code,
				Msg:  msg,
				Data: rsp.GetData(),
			},
		)
		return
	}

	app := req.GetApp()
	method := req.GetMethod()
	detail := map[string]string{
		TagApp:    app,
		TagMethod: method,
	}

	ZapFailed(LogS1,
		LogEvent(LogMsgGateway),
		LogProcRpc(),
		LogDetail(detail),
		LogError(err),
	)

	c.JSON(
		http.StatusInternalServerError,
		FinalRsp{
			Code: CodeInternal,
			Msg:  fmt.Sprintf(callFailed, app),
			Data: ByteOfNullJson,
		},
	)
}

func SendRsp2(c *gin.Context, req *Req, rsp *Res, err error) {
	code := rsp.GetCode()
	msg := rsp.GetMsg()

	c.Set(LogFiledCode, int(code))
	c.Set(LogFiledMsg, msg)

	if err == nil {
		//StatusMethodNotAllowed
		c.JSON(
			http.StatusOK,
			FinalRsp2{
				Code: code,
				Msg:  msg,
				Data: rsp.GetData(),
			},
		)
		return
	}

	app := req.GetApp()
	method := req.GetMethod()
	detail := map[string]string{
		TagApp:    app,
		TagMethod: method,
	}

	ZapFailed(LogS1,
		LogEvent(LogMsgGateway),
		LogProcRpc(),
		LogDetail(detail),
		LogError(err),
	)

	c.JSON(
		http.StatusInternalServerError,
		FinalRsp2{
			Code: CodeInternal,
			Msg:  fmt.Sprintf(callFailed, app),
			Data: ByteOfNullJson,
		},
	)
}

func GetBodyFromGin(c *gin.Context) []byte {
	body, _ := ioutil.ReadAll(c.Request.Body)
	return body
}

func GetClientFromGin(c *gin.Context) string {
	return DeStrParam(c.GetHeader(HeaderClient), "??")
}

func GetUserAgentFromGin(c *gin.Context) string {
	return DeStrParam(c.GetHeader("User-Agent"), "??")
}

// 以下为gin-logger

func GinLogger() gin.HandlerFunc {
	return GinLoggerWithConfig(NewGinLoggerConfig())
}

type consoleColorModeValue int

const (
	autoColor consoleColorModeValue = iota
	disableColor
	forceColor
)

var (
	green            = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	white            = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellow           = string([]byte{27, 91, 57, 48, 59, 52, 51, 109})
	red              = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blue             = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magenta          = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyan             = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	reset            = string([]byte{27, 91, 48, 109})
	consoleColorMode = autoColor
)

var (
	noLog = []string{}
)

// LogFormatter gives the signature of the formatter function passed to LoggerWithFormatter
type LogFormatter func(params LogFormatterParams) string

type LogFormatterParams struct {
	Request *http.Request

	// TimeStamp shows the time after the server returns a response.
	TimeStamp time.Time
	// StatusCode is HTTP response code.
	StatusCode int
	// Latency is how much time the server cost to process a certain request.
	Latency time.Duration
	// ClientIP equals Context's ClientIP method.
	ClientIP string
	// Method is the HTTP method given to the request.
	Method string
	// Path is a path the client requests.
	Path string
	Code int
	// ErrorMessage is set if error has occurred in processing the request.
	ErrorMessage string
	// isTerm shows whether does gin's output descriptor refers to a terminal.
	IsTerm bool
	// BodySize is the size of the Response Body
	BodySize int
	// Keys are the keys set on the request's context.
	Keys map[string]interface{}
}

// StatusCodeColor is the ANSI color for appropriately logging http status code to a terminal.
func (p *LogFormatterParams) StatusCodeColor() string {
	code := p.StatusCode

	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

// IsOutputColor indicates whether can colors be outputted to the log.
func (p *LogFormatterParams) IsOutputColor() bool {
	return consoleColorMode == forceColor || (consoleColorMode == autoColor && p.IsTerm)
}

// MethodColor is the ANSI color for appropriately logging http method to a terminal.
func (p *LogFormatterParams) MethodColor() string {
	method := p.Method

	switch method {
	case "GET":
		return blue
	case "POST":
		return cyan
	case "PUT":
		return yellow
	case "DELETE":
		return red
	case "PATCH":
		return green
	case "HEAD":
		return magenta
	case "OPTIONS":
		return white
	default:
		return reset
	}
}

// ResetColor resets all escape attributes.
func (p *LogFormatterParams) ResetColor() string {
	return reset
}

// defaultLogFormatter is the default log format function Logger middleware uses.
var defaultLogFormatter = func(param LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		// Truncate in a golang < 1.8 safe way
		param.Latency = param.Latency - param.Latency%time.Second
	}

	if runtime.GOOS == "windows" {
		return fmt.Sprintf("[GIN] %3d | %13v | %15s | %-7s %s | %d |%s",
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
			param.Code,
			param.ErrorMessage,
		)
	}

	return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %s | %d |%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor,
		param.Method,
		resetColor,
		param.Path,
		param.Code,
		param.ErrorMessage,
	)
}

type GinLoggerConfig struct {
	Formatter LogFormatter

	SkipPaths []string
}

func GinNoLog(url ...string) {
	noLog = append(noLog, url...)
}

func GinLoggerWithConfig(conf GinLoggerConfig) gin.HandlerFunc {
	notLogged := conf.SkipPaths

	var skip map[string]struct{}

	if length := len(notLogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notLogged {
			skip[path] = struct{}{}
		}
	}

	var h = func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path

		// Process request
		c.Next()

		// Log only when path is not being skipped
		_, ok := skip[path]
		if ok {
			return
		}

		//printGinLog(c, path, start)

		cost := time.Now().Sub(start)

		costLogger, ok := c.Get(TagLoggerCost)
		if ok {
			if f, okLogger := costLogger.(GinPluginLogger); okLogger {
				f(c, path, start, cost)
			}
		}

		f := []zap.Field{
			LogEvent(LogMsgGateway),
			LogProcessor(path),
			LogContent(c.GetString(LogFiledMsg)),
			LogDuration(cost),
			LogIntCode(c.GetInt(LogFiledCode)),
			LogUser(GetUserAccount(c)),
			LogFrom(GetRemote(c)),
			LogClient(GetClient(c)),
			LogUserAgent(GetUserAgent(c)),
		}

		msg := c.Param(TagMethod)
		if msg == "" {
			msg = "gin"
		}

		logger := Logger(c).Skip(1)
		logger.Info(msg, f...)
	}

	return h
}

func printGinLog(c *gin.Context, path string, start time.Time) {
	param := LogFormatterParams{
		Request: c.Request,
		IsTerm:  true,
		Keys:    c.Keys,
	}

	// Stop timer
	param.TimeStamp = time.Now()
	param.Latency = param.TimeStamp.Sub(start)
	param.ClientIP = c.ClientIP()
	param.Method = c.Request.Method
	param.StatusCode = c.Writer.Status()
	param.Code = c.GetInt(LogFiledCode)
	param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
	param.BodySize = c.Writer.Size()
	param.Path = path

	fmt.Println(defaultLogFormatter(param))
}

func NewGinLoggerConfig() GinLoggerConfig {
	res := GinLoggerConfig{
		Formatter: nil,
		SkipPaths: noLog,
	}

	return res
}
