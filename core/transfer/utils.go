package transfer

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	. "mykit/core/dsp"
	. "mykit/core/types"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/levigross/grequests"
	"github.com/micro/go-micro/v2/metadata"
)

var (
	localAddr  string
	publicAddr string
)

func NewMetaContext(ctx context.Context, md map[string]string) context.Context {
	return metadata.NewContext(ctx, md)
}

func MetaFromContext(ctx context.Context) (metadata.Metadata, bool) {
	return metadata.FromContext(ctx)
}

func ImportMsg(o int, app, method string) string {
	if o == 0 {
		return fmt.Sprintf(importPat, app, method)
	}

	return fmt.Sprintf(overridePat, app, method)
}

func GetIp() (res string) {
	if localAddr != "" {
		return localAddr
	}

	defer func() {
		localAddr = res
	}()

	return GetLocalIp()
}

func GetPublic() (res string) {
	if publicAddr != "" {
		return publicAddr
	}

	defer func() {
		publicAddr = res
	}()

	res, err := GetPublicIP()
	if err != nil {
		return "127.0.0.1"
	}

	return
}

func GetLocalIp() (res string) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return "127.0.0.1"
	}

	udpAddr := conn.LocalAddr().(*net.UDPAddr)
	return strings.Split(udpAddr.String(), ":")[0]
}

func GetPublicIP(url ...string) (string, error) {
	option := grequests.RequestOptions{
		DialTimeout:    time.Second * 1,
		RequestTimeout: time.Second * 3,
	}

	target := ParseStrParam(url, "http://ipinfo.io/ip")
	resp, err := HttpGet("GetPublicIP", target, option)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}

func GetListen(raw string) string {
	if raw == "" {
		return ""
	}

	port := GetPort(raw)
	if port < 0 {
		return ""
	}

	if port == 0 {
		port = 80
	}

	return fmt.Sprintf("%v:%v", GetIp(), port)
}

func ValidateAddress(address *string) (tcpAddr *net.TCPAddr, valid bool) {
	host, port := GetHostAndPort(*address)
	if port < 0 {
		return
	}

	if port == 0 {
		port = 80
	}

	addr := fmt.Sprintf("%v:%d", host, port)

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return
	}

	*address = addr

	return tcpAddr, true
}

func ParseTcp(address *string, must ...bool) (tcpAddr *net.TCPAddr) {
	tcpAddr, ok := ValidateAddress(address)
	if ok {
		return
	}

	msg := fmt.Sprintf("check address [%v]", *address)
	HandleInitErr(msg, ErrInvalidParam, must...)

	return
}

func ParsePrefixedTcp(prefix string, address *string, must ...bool) (tcpAddr *net.TCPAddr) {
	tcpAddr = ParseTcp(address, must...)

	PadPrefix(address, prefix)

	return
}

func ParseHttpAddress(address *string, must ...bool) (tcpAddr *net.TCPAddr) {
	return ParsePrefixedTcp("http://", address, must...)
}

func ParseWsAddress(address *string, must ...bool) (tcpAddr *net.TCPAddr) {
	return ParsePrefixedTcp("ws://", address, must...)
}

func HPUrl(raw *url.URL) string {
	port := raw.Port()
	if port == "" || strings.Contains(raw.Host, port) {
		return fmt.Sprintf("%s%s", raw.Host, raw.Path)
	}

	return fmt.Sprintf("%s:%s%s", raw.Host, port, raw.Path)
}

func BindParam(c *gin.Context, obj interface{}) (err error) {
	err = c.ShouldBindJSON(&obj)
	if err != nil {
		return
	}

	err = ValidateStruct(obj)
	if err != nil {
		NewFinalRsp2(err.Error(), CodeFailedOnRequired).Send(c)
		return
	}

	return
}

func (t FinalRsp) Send(c *gin.Context) {
	BeforeSend(c)

	c.Set(LogFiledCode, int(t.Code))
	c.Set(LogFiledMsg, t.Msg)

	c.JSON(http.StatusOK, t)
}

func (t FinalRsp) ToSend() []byte {
	return MustJsonMarshal(t)
}

func (t FinalRsp2) Send(c *gin.Context) {
	BeforeSend(c)

	c.Set(LogFiledCode, int(t.Code))
	c.Set(LogFiledMsg, t.Msg)

	c.JSON(http.StatusOK, t)
}

func (t FinalRsp2) ToSend() []byte {
	return MustJsonMarshal(t)
}

func SendFinalRsp(c *gin.Context, obj interface{}, err error) {
	if err == nil {
		DumpFinalRsp(RspMsgSuccess, obj, http.StatusOK).Send(c)
		return
	}

	v, ok := err.(*ErrorCode)
	if ok {
		ErrorCodeFinalRsp(obj, v).Send(c)

	} else {
		method := c.Param(TagMethod)
		msg := fmt.Sprintf("call %v failed", method)
		NewFinalRsp(msg, CodeInternal).Send(c)
	}
}

func SendFinalRsp2(c *gin.Context, obj interface{}, err error) {
	if err == nil {
		SuccessFinalRsp2(obj).Send(c)
		return
	}

	v, ok := err.(*ErrorCode)
	if ok {
		ErrorCodeFinalRsp2(obj, v).Send(c)

	} else {
		method := c.Param(TagMethod)
		msg := fmt.Sprintf("call %v failed", method)
		NewFinalRsp2(msg, CodeInternal).Send(c)
	}
}

func GetPort(raw string) int {
	if raw == "" {
		return -1
	}

	tmp := strings.Split(raw, ":")

	if len(tmp) == 1 {
		res, _ := strconv.Atoi(tmp[0])
		return res
	}

	res, _ := strconv.Atoi(tmp[1])
	return res
}

func GetHostAndPort(raw string) (host string, port int) {
	if raw == "" {
		port = -1
		return
	}

	tmp := strings.Split(raw, ":")

	if len(tmp) == 1 {
		port, _ = strconv.Atoi(tmp[0])
		return
	}

	host = tmp[0]
	port, _ = strconv.Atoi(tmp[1])

	return
}

// GetRemoteIP 返回远程客户端IP
func GetRemoteIP(req *http.Request) string {
	remoteAddr := req.RemoteAddr

	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get("X-Forwarded-For"); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

//RsaEncryptWithSha1Base64   rsa加密 to base64
func RsaEncryptWithSha1Base64(originalData string, publicKey []byte) (string, error) {
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return "", errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	pub := pubInterface.(*rsa.PublicKey)
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(originalData))
	return base64.StdEncoding.EncodeToString(encryptedData), err
}

//RsaDecryptWithSha1Base64  rsa解密to string
func RsaDecryptWithSha1Base64(encryptedData string, privateKey []byte) (string, error) {
	encryptedDecodeBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return "", errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}
	originalData, err := rsa.DecryptPKCS1v15(rand.Reader, priv, encryptedDecodeBytes)
	return BytesToString(originalData), err
}

//GenerateTokenString :生成token
func GenerateTokenString(val string, cert []byte) (token string, err error) {
	token, err = RsaEncryptWithSha1Base64(val, cert)
	if err != nil {
		LogS1.Error("GenerateTokenString",
			LogError(err),
		)
		return "", err
	}

	return
}

var (
	defaultPingCount = uint64(4)
	cc               = NewCallback()
)

func SetDefaultPingCount(raw uint64) {
	defaultPingCount = raw
}

type SimpleValidator map[string]bool

func (t SimpleValidator) IsValid(method string) bool {
	return t[method]
}

func NewReverseProxy(target *url.URL, m ...func(r *http.Request)) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		log.Println(req.Method, req.Host, req.RequestURI, req.ContentLength)

		originalDirector(req)

		for _, f := range m {
			f(req)
		}
	}

	return proxy
}

func ReplaceUrl(old string, s ...string) func(req *http.Request) {
	if len(s) == 0 {
		return func(r *http.Request) {}
	}

	to := s[0]
	if old == to {
		return func(r *http.Request) {}
	}

	var res = func(req *http.Request) {
		target := strings.ReplaceAll(req.RequestURI, old, to)
		if strings.Contains(target, "?") {
			target = strings.Split(target, "?")[0]
		}

		req.RequestURI = target
		req.URL.Path = target
	}

	return res
}

func (t RewritePrefix) Match(raw *url.URL) bool {
	return strings.HasPrefix(raw.Path, string(t))
}

func (t RewritePrefix) RewriteGin(c *gin.Context, target *url.URL) {
	RewriteGin(c, string(t), target)
}

func StartProxy(prefix RewritePrefix, addr, to string) {
	if !strings.HasPrefix(to, "http://") {
		to = "http://" + to
	}
	target, err := url.Parse(to)
	if err != nil {
		return
	}
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(gin.Recovery())

	engine.NoRoute(func(c *gin.Context) {
		prefix.RewriteGin(c, target)
	})

	addr = "0.0.0.0:" + addr
	fmt.Printf("start on %v, target -> %v\n", addr, to)

	go engine.Run(addr)
}

func (t RewriteDisp) Start() {
	for k, v := range t.Disp {
		StartProxy(t.Prefix, k, v)
	}
}

var (
	PingPointKind = "ping"
	RttPointKind  = "rtt"
)

func NewStation(ip, station string, name ...string) Station {
	res := Station{
		Name:    ParseStrParam(name, ip),
		Ip:      ip,
		Station: station,
	}

	res.PingPoint = func(n ...int64) string {
		return GenPoint(station, PingPointKind, n...)
	}

	res.RttPoint = func(n ...int64) string {
		return GenPoint(station, RttPointKind, n...)
	}

	return res
}
