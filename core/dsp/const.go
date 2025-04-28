package dsp

import (
	"net/http"
)

const (
	StatusStopped = -999
	StatusInited  = 1000
)

const (
	prime32   = uint32(16777619)
	DCS       = "dcs"
	HEARTBEAT = "heartbeat"
	METRIC    = "metric"
)

const (
	LogSenderGrpc  = 1
	LogSenderKafka = 2
)

const (
	TagLoggerCost = "LoggerCost"
	TagStation    = "station"
	TagFrame      = "frame"
)

const (
	PolicyRead   = "read"
	PolicyWrite  = "write"
	PolicyAccess = "access"
)

const (
	RspMsgSuccess    = "success"
	RspMsgBadRequest = "BadRequest"
	RspMsgForbidden  = "无操作权限"
	RspCodeSuccess   = http.StatusOK
	RspCodeMsg       = http.StatusPartialContent
	RspCodeForbidden = http.StatusForbidden
)

const (
	CodeUnimplemented    = -12
	CodeInvalidArgument  = -3
	CodeDeadlineExceeded = -4
	CodeFailedOnRequired = -36
	CodeInternal         = -13
)

const (
	TagVersion    = "version"
	TagRequest    = "request"
	TagRemote     = "remote"
	TagClient     = "client"
	TagUserAgent  = "ua"
	TagCredential = "credential"
	TagApp        = "app"
	TagMethod     = "method"
	TagStatus     = "status"
)

const (
	TagScc      = "scc"
	TagScn      = "scn"
	TagSrc      = "src"
	TagDst      = "dst"
	TagUid      = "uid"
	TagUserInst = "userInst"
	TagLanguage = "language"
)
