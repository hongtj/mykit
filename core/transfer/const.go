package transfer

import (
	"time"
)

const (
	RpcIot   = "iot"
	RpcUtils = "utils"
)

const (
	logImport      = "import"
	logOverride    = "override"
	importPat      = "app [%v] import [%v]"
	overridePat    = "app [%v] override [%v]"
	unknownMethod  = "unknown method [%v]"
	invalidMethod  = "invalid method [%v], code %v"
	invalidLpcMeta = "reg invalid lpc meta to [%v]"
	dupedLpcMeta   = "[%v] has imported to [%v]"
	appNotImpl     = "app [%v] not implemented"
	methodNotImpl  = "method [%v] not implemented"
	decodeErr      = "decode err, payload size %v"
	callFailed     = "call [%v] failed"
	callTimeout    = "call [%v] timeout"
)

const (
	defaultSpanStr        = "8888888888888888"
	rpcQuerySlowThreshold = time.Second * 3
)

const (
	ContentLength   = "Content-Length"
	ContentLanguage = "Content-Language"
	ContentType     = "Content-Type"
	HtmlContentType = "text/html; charset=utf-8"
	JsonContentType = "application/json; charset=utf-8"
	ApplicationJson = "application/json"
	ApplicationForm = "application/x-www-form-urlencoded"
)

const (
	ProtocolWeb  = "web"
	ProtocolApp  = "app"
	ProtocolScc  = "scc"
	ProtocolNode = "node"
)

const (
	HeadSign       = "Sign"
	HeadXToken     = "X-Token"
	HeadAToken     = "A-Token"
	HeaderClient   = "Client"
	HeaderUa       = "Ua"
	HeadFrom       = "From"
	HeadScc        = "Scc"
	HeadScn        = "Scn"
	HeadSession    = "Session"
	HeadSecret     = "Secret"
	HeadTrace      = "Trace"
	HeadUser       = "User"
	HeadCredential = "Credential"
	HeadNonce      = "Nonce"
	HeadLanguage   = "Language"
	HeadTenant     = "Tenant"
	HeadTs         = "Ts"
	HeadRole       = "Role"
	HeadMenu       = "Menu"
	HeadWxApp      = "wxapp"
)
