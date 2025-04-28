package persist

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	. "mykit/core/dsp"
	. "mykit/core/internal"
	. "mykit/core/types"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/go-ini/ini"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-plugins/registry/etcdv3"
)

type EtcdConfig struct {
	Hosts              []string
	Key                string `json:",optional"`
	User               string `json:",optional"`
	Pass               string `json:",optional"`
	CertFile           string `json:",optional"`
	CertKeyFile        string `json:",default=CertFile"`
	CACertFile         string `json:",default=CACertFile"`
	InsecureSkipVerify bool   `json:",optional"`
}

func (t EtcdConfig) MicroRegistry() registry.Registry {
	var addr = registry.Addrs(t.Hosts...)

	var auth = etcdv3.Auth(t.User, t.Pass)

	res := etcdv3.NewRegistry(addr, auth)

	return res
}

func NewEtcdConfigFromIni(cfg *ini.Section) EtcdConfig {
	hosts := cfg.Key("hosts").String()
	user := cfg.Key("user").String()
	pass := cfg.Key("pass").String()

	res := EtcdConfig{
		Hosts: strings.Split(hosts, ","),
		User:  user,
		Pass:  pass,
	}

	return res
}

// EtcdAccount holds the username/password for an etcd cluster.
type EtcdAccount struct {
	User string
	Pass string
}

const (
	EtcdDelimiter      = "/"
	autoSyncInterval   = time.Minute
	coolDownInterval   = time.Second
	dialTimeout        = 3 * time.Second
	dialKeepAliveTime  = 5 * time.Second
	requestTimeout     = 3 * time.Second
	endpointsSeparator = ","
)

var (
	accounts     = map[string]EtcdAccount{}
	tlsConfigs   = map[string]*tls.Config{}
	etcdInitLock sync.RWMutex
)

func DialEtcd(conf EtcdConfig) (*clientv3.Client, error) {
	cfg := clientv3.Config{
		Endpoints:            conf.Hosts,
		AutoSyncInterval:     autoSyncInterval,
		DialTimeout:          dialTimeout,
		DialKeepAliveTime:    dialKeepAliveTime,
		DialKeepAliveTimeout: dialTimeout,
		Username:             conf.User,
		Password:             conf.Pass,
		RejectOldCluster:     true,
	}
	if account, ok := GetEtcdAccount(conf.Hosts); ok {
		cfg.Username = account.User
		cfg.Password = account.Pass
	}
	if tlsCfg, ok := GetTLS(conf.Hosts); ok {
		cfg.TLS = tlsCfg
	}

	return clientv3.New(cfg)
}

func MustGetEtcdClient(conf EtcdConfig) *clientv3.Client {
	res, err := DialEtcd(conf)
	HandleInitErr("DialEtcd", err)

	return res
}

func getClusterKey(endpoints []string) string {
	sort.Strings(endpoints)
	return strings.Join(endpoints, endpointsSeparator)
}

// GetEtcdAccount gets the username/password for the given etcd cluster.
func GetEtcdAccount(endpoints []string) (EtcdAccount, bool) {
	etcdInitLock.RLock()
	defer etcdInitLock.RUnlock()

	account, ok := accounts[getClusterKey(endpoints)]
	return account, ok
}

// GetTLS gets the tls config for the given etcd cluster.
func GetTLS(endpoints []string) (*tls.Config, bool) {
	etcdInitLock.RLock()
	defer etcdInitLock.RUnlock()

	cfg, ok := tlsConfigs[getClusterKey(endpoints)]
	return cfg, ok
}

func ParseEtcdGetResponseToKV(raw *clientv3.GetResponse) []KV {
	var res []KV
	for _, ev := range raw.Kvs {
		item := KV{
			K: BytesToString(ev.Key),
			V: BytesToString(ev.Value),
		}
		res = append(res, item)
	}

	return res
}

func ParseEtcdGetResponseToMap(raw *clientv3.GetResponse) map[string]string {
	res := map[string]string{}

	for _, ev := range raw.Kvs {
		res[BytesToString(ev.Key)] = BytesToString(ev.Value)
	}

	return res
}

type EtcdContext struct {
	*clientv3.Client
}

func NewEtcdContext(conf EtcdConfig) *EtcdContext {
	res := &EtcdContext{
		Client: MustGetEtcdClient(conf),
	}

	var job = func() {
		if res.Client != nil {
			res.Client.Close()
		}
	}

	TeardownJobs = append(TeardownJobs, job)

	return res
}

func (t *EtcdContext) Close() error {
	return t.Client.Close()
}

func (t *EtcdContext) Get(ctx context.Context, key string, obj interface{}) error {
	resp, err := t.Client.Get(ctx, key)
	if err != nil {
		return err
	}

	kvs := ParseEtcdGetResponseToKV(resp)
	if len(kvs) == 0 {
		return nil
	}

	return json.Unmarshal(StringToBytes(kvs[0].V), &obj)
}

func (t *EtcdContext) GetWithPrefix(ctx context.Context, key string) map[string]string {
	resp, err := t.Client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return map[string]string{}
	}

	return ParseEtcdGetResponseToMap(resp)
}

func (t *EtcdContext) Put(ctx context.Context, key string, obj interface{}) error {
	_, err := t.Client.Put(ctx, key, ToJsonStr(obj))

	return err
}

func (t *EtcdContext) TryGet(ctx context.Context, key, try string) (res string, err error) {
	txn := t.Txn(ctx)

	rsp, err := txn.If(MatchNewKey(key)).
		Then(clientv3.OpPut(key, try)).
		Else(clientv3.OpGet(key)).
		Commit()
	if err != nil || !rsp.Succeeded || len(rsp.Responses) == 0 {
		return
	}

	response := rsp.Responses[0].GetResponseRange()
	if response == nil || len(response.Kvs) == 0 {
		res = try
		return
	}

	res = BytesToString(response.Kvs[0].Value)

	return
}

func (t *EtcdContext) RunStm(f func(stm concurrency.STM) error) (rsp *clientv3.TxnResponse, err error) {
	rsp = &clientv3.TxnResponse{}

	return concurrency.NewSTM(t.Client, f)
}

func (t *EtcdContext) Watch(k string, h ...EtcdEventHandler) {
	var handler EtcdEventHandler
	if len(h) > 0 {
		handler = h[0]
	} else {
		handler = DefaultEtcdEventHandler
	}

	c := t.Client.Watch(context.Background(), k, clientv3.WithPrefix())

	for v := range c {
		for _, v2 := range v.Events {
			if v2.Type == EtcdEventPut && v2.Kv != nil {
				handler(v2.Kv)
			}
		}
	}
}

func DefaultEtcdEventHandler(raw *mvccpb.KeyValue) {
	log.Printf("recv etcd event: %v\n", raw)
}

func MatchNewKey(key string) clientv3.Cmp {
	return clientv3.Compare(clientv3.CreateRevision(key), "=", 0)
}
