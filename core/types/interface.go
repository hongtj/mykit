package types

import (
	"context"
	"reflect"
	"time"
)

type PluginConf []string

type ValuePair [2]interface{}

type ValuePairList []ValuePair

type StaticValue [2]string

type StaticValueList []StaticValue

type Chunk struct {
	Fn    string
	Total int
	Index int
	Data  []byte
}

type Checker func(ctx context.Context) error

type ConnectionId uintptr

type SimpleWriter struct {
	b []byte
}

type I64K interface {
	K() []int64
}

type Sender interface {
	Protocol() string
	ID() ConnectionId
	RemoteAddr() string
	Ping() error
	Close() error
	Send(context.Context, []byte) error
}

type SendFunc func(context.Context, []byte) error

type LogSender interface {
	Send([]byte) (err error)
}

type LogSenderMaker func() (mode int, sender LogSender, err error)

type Msg interface {
	Data() []byte
}

type MsgPod interface {
	Key() string
	Msg(ctx context.Context) Msg
	Interval() TaskInterval
}

type CURSOR interface {
	Get(key string) int64                   //取游标值
	Set(key string, n int64) (int64, error) //设游标值
	Inc(key string, n int64) (int64, error) //游标递增
}

type Selector interface {
	Select(string) int
}

type Speaker interface {
	SetSpeaker(Speech)
	DoSpeech(context.Context) string
}

type Receiver interface {
	Speaker
	Receive(context.Context)
}

type TracedLog interface {
	NewTrace(ctx context.Context)
}

type IgnoreLog interface {
	IgnoreLog()
}

type InitProgress func(interface{})

type DeleteAll interface {
	DeleteAll() error
}

type ErrorFunc func() error

type ErrorCtxFunc func(ctx context.Context) error

type Processor func([]byte) ([]byte, error)

type MsgSource func() Msg

type MsgCheck func([]byte) bool

type MsgDecoder func([]byte) ([]byte, error)

type ByteSource func(string) []byte

type QueryCmdMaker func(table string) string

type Speech func(ctx context.Context) string

type TaskInterval func() time.Duration

type IntervalComputer func(ctx context.Context, t time.Time) time.Duration

type JOB func(ctx context.Context)

var (
	CtxKind        = reflect.TypeOf(Ctx).Kind()
	ErrKind        = reflect.TypeOf(ErrInvalidParam).Kind()
	NormalCtxInput = []reflect.Value{reflect.ValueOf(Ctx)}
)
