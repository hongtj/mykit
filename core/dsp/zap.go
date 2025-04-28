package dsp

import (
	"context"
	"fmt"
	. "mykit/core/types"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logInitOnce sync.Once

	zLogger  *zap.Logger
	tracking *zap.Logger
	LogS1    *zap.Logger
)

const (
	defaultMsg         = "zlog"
	logPatAny          = "%v"
	logPatAnyWithSpace = "%v "
)

type ZlogConfig struct {
	Caller   bool
	LogLevel string
	SendMode int
}

func NewZLogConfig(raw LogConfig) ZlogConfig {
	res := ZlogConfig{
		Caller:   raw.Caller,
		LogLevel: raw.LogLevel,
		SendMode: raw.SendMode,
	}

	return res
}

func InitLog(raw LogConfig) {
	logInitOnce.Do(func() {
		initLog(raw)
	})
}

func initLog(raw LogConfig) {
	conf := NewZLogConfig(raw)

	core := newZapCore(conf)

	f := []zap.Field{
		zap.String(TagProject, PROJECT()),
		zap.String(TagTenant, TENANT()),
		zap.String(TagEnv, EnvNAME()),
		zap.String(TagHost, HOST()),
		zap.String(TagInst, INST()),
	}

	if conf.Caller {
		zLogger = zap.New(core, zap.AddCaller())
	} else {
		zLogger = zap.New(core)
	}
	zLogger = zLogger.With(f...)

	LogS1 = Logger(Ctx).Skip(1)

	tracking = zap.New(core).With(f...)
}

func SyncLog() {
	zLogger.Sync()
	LogS1.Sync()
	tracking.Sync()
}

func zapEncoderConfig() zapcore.EncoderConfig {
	res := zapcore.EncoderConfig{
		TimeKey:        LogFiledTs,
		LevelKey:       LogFiledLevel,
		NameKey:        logFiledName,
		CallerKey:      LogFiledCaller,
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     LogFiledMsg,
		StacktraceKey:  logFiledStacktrace,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	return res
}

func parseZapLevel(raw string) zap.AtomicLevel {
	logLevel := DeStrParam(raw, "DEBUG")
	logLevel = strings.ToUpper(logLevel)

	atomicLevel := zap.NewAtomicLevel()
	switch logLevel {
	case "DEBUG":
		atomicLevel.SetLevel(zapcore.DebugLevel)
	case "INFO":
		atomicLevel.SetLevel(zapcore.InfoLevel)
	case "WARN":
		atomicLevel.SetLevel(zapcore.WarnLevel)
	case "ERROR":
		atomicLevel.SetLevel(zapcore.ErrorLevel)
	case "DPANIC":
		atomicLevel.SetLevel(zapcore.DPanicLevel)
	case "PANIC":
		atomicLevel.SetLevel(zapcore.PanicLevel)
	case "FATAL":
		atomicLevel.SetLevel(zapcore.FatalLevel)
	}

	return atomicLevel
}

type defaultLogSender struct {
}

func (t defaultLogSender) Send(raw []byte) (err error) {
	fmt.Print(BytesToString(raw))

	return
}

var (
	senderM = map[int]LogSender{}
)

func RegLogSender(m ...LogSenderMaker) {
	for _, v := range m {
		mode, sender, err := v()
		if err != nil {
			continue
		}

		senderM[mode] = sender
	}
}

type SmarterLog struct {
	LogSender
}

func NewSmarterLog(conf ZlogConfig) SmarterLog {
	res := SmarterLog{}

	sender, ok := senderM[conf.SendMode]
	if ok {
		res.LogSender = sender
		return res
	}

	res.LogSender = defaultLogSender{}

	return res
}

func (t SmarterLog) Write(raw []byte) (n int, err error) {
	if len(raw) == 0 {
		return
	}

	err = t.Send(raw)
	if err != nil {
		//fmt.Printf("mode %v send err: %v", t.SendMode, err)
	}

	return len(raw), nil
}

func newZapCore(conf ZlogConfig) zapcore.Core {
	zapCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapEncoderConfig()),
		zapcore.AddSync(NewSmarterLog(conf)),
		parseZapLevel(conf.LogLevel),
	)

	return zapCore
}

type ZLogger struct {
	*zap.Logger
	prefixed  bool
	msg       string
	procDebug bool
}

func LogS(n int) *zap.Logger {
	return loggerSkip(LogS1, n)
}

func loggerSkip(l *zap.Logger, n int) *zap.Logger {
	res := l.WithOptions(zap.AddCallerSkip(n))

	return res
}

func newLogger(ctx context.Context, logger *zap.Logger, msg string) *ZLogger {
	res := &ZLogger{}
	res.Logger = loggerSkip(logger, 1)
	res.msg = msg
	res.WithTrace(ctx)
	res.procDebug = true

	return res
}

func NewZLogger(ctx context.Context, msg string) *ZLogger {
	return newLogger(ctx, zLogger, msg)
}

func NewZLoggerWithFields(ctx context.Context, msg string, fields ...zap.Field) *ZLogger {
	res := &ZLogger{}
	res.Logger = loggerSkip(zLogger, 1)
	res.msg = msg
	res.WithTrace(ctx)
	res.procDebug = true

	res.WithFields(fields...)

	return res
}

func Logger(ctx context.Context) *ZLogger {
	return NewZLogger(ctx, defaultMsg)
}

func (t *ZLogger) Skip(n int) *zap.Logger {
	return loggerSkip(t.Logger, n)
}

func (t *ZLogger) Msg(s string) *ZLogger {
	t.msg = s

	return t
}

func (t *ZLogger) ProcDebug(b bool) *ZLogger {
	t.procDebug = b

	return t
}

func (t *ZLogger) WithFields(fields ...zap.Field) {
	t.Logger = t.Logger.With(fields...)
}

func (t *ZLogger) WithTrace(ctx context.Context) {
	trace := GetStringFromContext(ctx, TagTrace)
	span := GetStringFromContext(ctx, TagSpan)
	if len(trace) > 0 {
		t.WithFields(LogTrace(trace), LogSpan(span))
	}
}

func (t *ZLogger) NewTrace(ctx context.Context) {
	_ = t.Logger.Sync()

	t.Logger = loggerSkip(zLogger, 1)
	t.WithTrace(ctx)
}

func (t *ZLogger) Output(msg string, err error, fields ...zap.Field) {
	if err != nil {
		fields = append(fields, LogError(err))
		t.Error(msg, fields...)
	} else {
		t.Info(msg, fields...)
	}
}

func (t *ZLogger) Slow(msg string, fields ...zap.Field) {
	fields = append(fields, LogPerformanceSlow())
	t.Logger.Warn(msg, fields...)
}

func (t *ZLogger) Debugf(template string, args ...interface{}) {
	t.Logger.Debug(t.msg, LogContentf(template, args...))
}

func (t *ZLogger) Debugs(args ...interface{}) {
	var b strings.Builder

	l := len(args)
	for i := 0; i < l-1; i++ {
		b.WriteString(logPatAnyWithSpace)
	}

	b.WriteString(logPatAny)

	t.Logger.Debug(t.msg, LogContentf(b.String(), args...))
}

func (t *ZLogger) Infof(template string, args ...interface{}) {
	t.Logger.Info(t.msg, LogContentf(template, args...))
}

func (t *ZLogger) Infos(args ...interface{}) {
	var b strings.Builder

	l := len(args)
	for i := 0; i < l-1; i++ {
		b.WriteString(logPatAnyWithSpace)
	}

	b.WriteString(logPatAny)

	t.Logger.Info(t.msg, LogContentf(b.String(), args))
}

func (t *ZLogger) Warnf(template string, args ...interface{}) {
	t.Logger.Warn(t.msg, LogContentf(template, args...))
}

func (t *ZLogger) Slowf(template string, args ...interface{}) {
	t.Logger.Warn(t.msg,
		LogPerformanceSlow(),
		LogContentf(template, args...),
	)
}

func (t *ZLogger) Failed(fields ...zap.Field) {
	t.Error(LogMsgFailed, fields...)
}

func (t *ZLogger) Errorf(template string, args ...interface{}) {
	t.Logger.Error(t.msg, LogContentf(template, args...))
}

func (t *ZLogger) Errorx(s string, err error) {
	if err == nil {
		return
	}

	t.Logger.Error(t.msg, LogContent(s), LogError(err))
}

func (t *ZLogger) Errors(err error) {
	if err == nil {
		return
	}

	t.Logger.Error(t.msg, LogError(err))
}

func (t *ZLogger) info(k int, msg string, fields ...zap.Field) {
	t.Skip(k).Info(msg, fields...)
}

func (t *ZLogger) debug(k int, msg string, fields ...zap.Field) {
	t.Skip(k).Debug(msg, fields...)
}

func (t *ZLogger) Start(msg ...string) {
	t.info(1, ParseStrParam(msg, t.msg), LogProcStart())
}

func (t *ZLogger) End(cost time.Duration, msg ...string) {
	t.info(1, ParseStrParam(msg, t.msg), LogProcEnd(), zap.Duration(logFiledCost, cost))
}

func (t *ZLogger) Proc(proc string, cost time.Duration, n ...uint64) {
	f := []zap.Field{
		LogProcessor(proc),
		LogDuration(cost),
	}

	if len(n) > 0 {
		f = append(f, LogBatch(n[0]))
	}

	if t.procDebug {
		t.debug(1, t.msg, f...)
	} else {
		t.info(1, t.msg, f...)
	}
}

func (t *ZLogger) BatchStart(n uint64) {
	t.info(1, t.msg, LogProcStart(), LogBatch(n))
}

func (t *ZLogger) BatchEnd(cost time.Duration, n uint64) {
	t.info(1, t.msg, LogProcEnd(), LogDuration(cost), LogBatch(n))
}

func Tracking(msg string, fields ...zap.Field) {
	tracking.Info(msg, fields...)
}

func CallerLogger(caller string) *zap.Logger {
	return tracking.With(LogCaller(caller))
}

func ZapSlow(l *zap.Logger, msg string, fields ...zap.Field) {
	fields = append(fields, LogPerformanceSlow())
	l.Warn(msg, fields...)
}

func ZapFailed(l *zap.Logger, fields ...zap.Field) {
	l.Error(LogMsgFailed, fields...)
}
