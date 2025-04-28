package smarter

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	. "mykit/core/dsp"
	. "mykit/core/persist"
	. "mykit/core/transfer"
	. "mykit/core/types"
)

type EnvConfig struct {
	Release int32  `json:"release"` //发布模式
	Dcs     string `json:"dcs"`     //dcs地址
}

type CertPem struct {
	PrivatePem     string          `json:",default=private.pem"`
	PublicPem      string          `json:",default=public.pem"`
	CertPrivatePem []byte          `json:",optional"`
	CertPublicPem  []byte          `json:",optional"`
	publicKey      *rsa.PublicKey  `json:",optional"`
	privateKey     *rsa.PrivateKey `json:",optional"`
}

func (t *CertPem) LoadPem(must ...bool) {
	t.CertPrivatePem = LoadFile(t.PrivatePem, ParseBool(must))
	t.CertPublicPem = LoadFile(t.PublicPem, ParseBool(must))
}

func (t *CertPem) TryLoad() {
	t.LoadPem(false)
}

func (t *CertPem) Verify() {
	t.VerifyPublicPem()
	t.VerifyPrivatePem()
}

func (t *CertPem) VerifyPublicPem() {
	if len(t.CertPublicPem) == 0 {
		HandleInitErr("CertPublicPem is empty", ErrInvalidParam)
	}

	block, _ := pem.Decode(t.CertPublicPem)
	if block == nil {
		HandleInitErr("Decode public key", ErrInvalidBodyParam)
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	HandleInitErr("Parse public key", err)

	t.publicKey = key.(*rsa.PublicKey)
}

func (t *CertPem) VerifyPrivatePem() {
	if len(t.CertPrivatePem) == 0 {
		HandleInitErr("CertPrivatePem is empty", ErrInvalidParam)
	}

	block, _ := pem.Decode(t.CertPrivatePem)
	if block == nil {
		HandleInitErr("Decode private key", ErrInvalidBodyParam)
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	HandleInitErr("Parse private key", err)

	t.privateKey = key
}

func (t *CertPem) Encrypt(raw string) (res []byte) {
	m, err := rsa.EncryptPKCS1v15(rand.Reader, t.publicKey, StringToBytes(raw))
	if err != nil {
		return
	}

	res = make([]byte, base64.StdEncoding.EncodedLen(len(m)))

	base64.StdEncoding.Encode(res, m)

	return
}

func (t *CertPem) Decrypt(raw string) (res []byte, err error) {
	in, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return
	}

	return rsa.DecryptPKCS1v15(rand.Reader, t.privateKey, in)
}

type TenantConfig struct {
	Project     string `json:",optional"`
	Tenant      string `json:",optional"`
	EtcdPrefix  string `json:",optional"`
	DbPrefix    string `json:",optional"`
	RedisPrefix string `json:",optional"`
	RpcPrefix   string `json:",optional"`
}

func (t *TenantConfig) Init() {
	t.Project = PROJECT()
	t.Tenant = TENANT()

	t.EtcdPrefix = t.Project + EtcdDelimiter + t.Tenant + EtcdDelimiter
	initTenantEtcd(t.EtcdPrefix)

	t.DbPrefix = t.Project + dbDelimiter
	t.RedisPrefix = t.Project + RedisKeyDelimiter + t.Tenant + RedisKeyDelimiter
	t.RpcPrefix = t.Project + rpcDelimiter + t.Tenant + rpcDelimiter
}

type RpcConfig struct {
	Uri        string
	Endpoint   string     `json:",optional"`
	MaxMsgSize int        `json:",optional"`
	Etcd       EtcdConfig `json:",optional" help:"Etcd配置"`
}

func (t RpcConfig) RunRpc(raw ...SmarterHandler) {
	etcd := t.Etcd
	if len(etcd.Hosts) == 0 {
		etcd = SERVER().Etcd
	}

	var handler SmarterHandler = DISP
	if len(raw) > 0 {
		handler = raw[0]
	}

	server := NewSmarterServer(
		t.Endpoint,
		t.Uri,
		etcd,
		handler,
	)

	err := server.Run()
	HandleInitErr("run rpc", err)
}
