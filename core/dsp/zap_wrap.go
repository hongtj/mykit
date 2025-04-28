package dsp

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

const (
	LogFiledLevel  = "level"
	LogFiledTs     = "ts"
	LogFiledCaller = "caller"
	LogFiledMsg    = "msg"

	TagProject = "project"
	TagTenant  = "tenant"
	TagEnv     = "env"
	TagHost    = "host"
	TagInst    = "inst"
	TagTrace   = "trace"
	TagSpan    = "span"
	TagTime    = "@time"
	TagSession = "session"

	LogFiledEvent     = "event"
	logFiledProcessor = "proc"
	logFiledContent   = "content"
	logFiledCost      = "cost"
	LogFiledCode      = "code"

	TagUser        = "user"
	TagFrom        = "from"
	logFiledBinary = "binary"
	logFiledError  = "error"
	logFiledSpecif = "speci"
	logFiledExtra  = "extra"
)

const (
	LogMsgSetup   = "<SETUP>"
	LogMsgCron    = "<CRON>"
	LogMsgFailed  = "<FAILED>"
	LogMsgRecover = "<RECOVER>"
)

// db
const (
	LogMsgMongo    = "<MONGO>"
	LogMsgSql      = "<SQL>"
	LogMsgInflux   = "<INFLUX>"
	LogMsgRedis    = "<REDIS>"
	LogMsgTracking = "<TRACKING>"
)

//link
const (
	LogMsgPost      = "<POST>"
	LogMsgGrpc      = "<GRPC>"
	LogMsgScc       = "<SCC>"
	LogMsgProxy     = "<PROXY>"
	LogMsgGateway   = "<GATEWAY>"
	LogMsgTcp       = "<TCP>"
	LogMsgUdp       = "<UDP>"
	LogMsgWebSocket = "<WS>"
	LogMsgSession   = "<SESS>"
)

var (
	logDbMap = map[string]bool{
		LogMsgMongo:  true,
		LogMsgSql:    true,
		LogMsgInflux: true,
		LogMsgRedis:  true,
	}

	logLinkMap = map[string]bool{
		LogMsgPost:      true,
		LogMsgGrpc:      true,
		LogMsgScc:       true,
		LogMsgProxy:     true,
		LogMsgGateway:   true,
		LogMsgTcp:       true,
		LogMsgUdp:       true,
		LogMsgWebSocket: true,
		LogMsgSession:   true,
	}
)

const (
	logFiledName       = "logger"
	logFiledStacktrace = "stacktrace"
	logContentBatch    = "batch"
	logContentRpc      = "rpc"
	logContentSlow     = "slow"
	logContentRecv     = "recv"
	logContentSend     = "send"
)

type LogConfig struct {
	Caller   bool   `json:",default=true" help:"是否输出caller"`
	LogLevel string `json:",default=INFO" help:"日志级别"`
	SendMode int    `json:",default=0" help:"发送方式"`
}

func (t *LogConfig) SetDebug() {
	t.LogLevel = "DEBUG"
}

func (t *LogConfig) SetInfo() {
	t.LogLevel = "INFO"
}

func (t *LogConfig) SetWarn() {
	t.LogLevel = "WARN"
}

func (t *LogConfig) SetError() {
	t.LogLevel = "ERROR"
}

func IsLogDb(raw string) bool {
	return logDbMap[raw]
}

func IsLogLink(raw string) bool {
	return logLinkMap[raw]
}

func IsTracking(raw string) bool {
	return raw == LogMsgTracking
}

func LogCaller(s string) zap.Field {
	return zap.String(LogFiledCaller, s)
}

func LogTrace(s string) zap.Field {
	return zap.String(TagTrace, s)
}

func LogSpan(s string) zap.Field {
	return zap.String(TagSpan, s)
}

func LogPerformanceSlow() zap.Field {
	return LogSpec(logContentSlow)
}

func LogClient(s string) zap.Field {
	return LogSpec(s)
}

func LogUserAgent(s string) zap.Field {
	return LogExtra(s)
}

func LogSpec(s string) zap.Field {
	return zap.String(logFiledSpecif, s)
}

func LogExtra(s string) zap.Field {
	return zap.String(logFiledExtra, s)
}

func LogFrom(s string) zap.Field {
	if s == "" {
		return zap.Skip()
	}

	return zap.String(TagFrom, s)
}

func LogProcessor(s string) zap.Field {
	return zap.String(logFiledProcessor, s)
}

func LogDetail(v interface{}) zap.Field {
	return zap.Any(logFiledContent, v)
}

func LogUser(s string) zap.Field {
	return zap.String(TagUser, s)
}

func LogEvent(s string) zap.Field {
	return zap.String(LogFiledEvent, s)
}

func LogContent(s string) zap.Field {
	return zap.String(logFiledContent, s)
}

func LogBinary(b []byte) zap.Field {
	return zap.Binary(logFiledBinary, b)
}

func LogError(v error) zap.Field {
	if v == nil {
		return zap.Skip()
	}

	return zap.String(logFiledError, v.Error())
}

func LogEventGrpc() zap.Field {
	return LogEvent(LogMsgGrpc)
}

func LogEventScc() zap.Field {
	return LogEvent(LogMsgScc)
}

func LogEventSql() zap.Field {
	return LogEvent(LogMsgSql)
}

func LogEventInflux() zap.Field {
	return LogEvent(LogMsgInflux)
}

func LogProcRpc() zap.Field {
	return LogProcessor(logContentRpc)
}

func LogProcRecv() zap.Field {
	return LogProcessor(logContentRecv)
}

func LogProcSend() zap.Field {
	return LogProcessor(logContentSend)
}

func LogProcStart() zap.Field {
	return LogProcessor("start")
}

func LogProcEnd() zap.Field {
	return LogProcessor("end")
}

func LogContentf(template string, fmtArgs ...interface{}) zap.Field {
	return zap.String(logFiledContent, getMessage(template, fmtArgs))
}

func LogDuration(d time.Duration) zap.Field {
	return zap.Duration(logFiledCost, d)
}

func LogCode(n int32) zap.Field {
	return zap.Int32(LogFiledCode, n)
}

func LogIntCode(n int) zap.Field {
	return zap.Int(LogFiledCode, n)
}

func LogStep(n uint64) zap.Field {
	msg := fmt.Sprintf("step %d", n)
	return LogProcessor(msg)
}

func LogBatch(n uint64) zap.Field {
	detail := map[string]interface{}{
		logContentBatch: n,
	}
	return LogDetail(detail)
}

func getMsg(template string, fmtArgs ...interface{}) string {
	return getMessage(template, fmtArgs)
}

// getMessage format with Sprint, Sprintf, or neither.
func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}

	return fmt.Sprint(fmtArgs...)
}
