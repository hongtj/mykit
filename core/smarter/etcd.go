package smarter

import (
	"encoding/json"
	"fmt"
	. "mykit/core/dsp"
	. "mykit/core/persist"
	. "mykit/core/types"

	"github.com/coreos/etcd/clientv3"
)

const (
	etcdRoot = "smarter"

	etcdRootKey = etcdRoot + EtcdDelimiter + "root"
	etcdSets    = etcdRootKey + EtcdDelimiter + "sets"
	etcdPass    = etcdRootKey + EtcdDelimiter + "pass"

	etcdConfigKey    = etcdRoot + EtcdDelimiter + "config"
	etcdConfigEnv    = etcdConfigKey + EtcdDelimiter + "env"
	etcdConfigAccess = etcdConfigKey + EtcdDelimiter + "access"

	etcdMetaKey = etcdRoot + EtcdDelimiter + "meta"
)

var (
	etcdTenantConfig    = "config"
	etcdTenantConfigEnv = etcdTenantConfig + EtcdDelimiter + "env"

	etcdTenantMeta = "meta"
)

func initTenantEtcd(root string) {
	PadSuffix(&root, EtcdDelimiter)

	AddPrefix(root,
		&etcdTenantConfigEnv,
		&etcdTenantMeta,
	)
}

func GetKVFromEtcd(cli *clientv3.Client, key string, opts ...clientv3.OpOption) []KV {
	ctx := Ctx
	resp, err := cli.Get(ctx, key, opts...)
	if err != nil {
		return []KV{}
	}

	return ParseEtcdGetResponseToKV(resp)
}

func GetKVFromEtcdWithPrefix(cli *clientv3.Client, key string, opts ...clientv3.OpOption) []KV {
	return GetKVFromEtcd(cli, key, clientv3.WithPrefix())
}

func LoadConfFromEtcd(cli *clientv3.Client, key string, conf interface{}, mustLoad bool,
	opts ...clientv3.OpOption) error {
	kvs := GetKVFromEtcd(cli, key, opts...)
	if len(kvs) == 0 {
		if !mustLoad {
			return ErrNotFound
		}

		msg := fmt.Sprintf("LoadConfFromEtcd [%v]", key)
		HandleInitErr(msg, ErrNotFound)
	}

	confStr := kvs[0].V
	err := json.Unmarshal(StringToBytes(confStr), &conf)
	if err != nil {
		msg := fmt.Sprintf("LoadConfFromEtcd [%v], data len %v\n%v\n", key, len(confStr), confStr)
		HandleInitErr(msg, err)
	}

	return nil
}

func LoadConfFromEtcdKeys(cli *clientv3.Client, obj string, conf interface{}, keys ...string) (err error) {
	if len(keys) == 0 {
		return ErrNotFound
	}

	for _, v := range keys {
		err = LoadConfFromEtcd(cli, v, conf, false)
		if err == nil {
			return
		}
	}

	if err != nil {
		msg := fmt.Sprintf("load %v", obj)
		HandleInitErr(msg, err)
	}

	return err
}

func MustLoadConfFromEtcd(cli *clientv3.Client, key string, conf interface{}, opts ...clientv3.OpOption) {
	LoadConfFromEtcd(cli, key, conf, true, opts...)
}

func LoadEnv(conf EtcdConfig) EnvConfig {
	cli := MustGetEtcdClient(conf)
	defer cli.Close()

	res := EnvConfig{}
	LoadConfFromEtcdKeys(cli, "env", &res, etcdTenantConfigEnv, etcdConfigEnv)

	return res
}

func LoadAccess(conf EtcdConfig, key string, opts ...clientv3.OpOption) ACCESS {
	cli := MustGetEtcdClient(conf)
	defer cli.Close()

	kvs := GetKVFromEtcd(cli, AccessKey(key), opts...)
	if len(kvs) == 0 {
		msg := fmt.Sprintf("LoadAccess [%v]", key)
		HandleInitErr(msg, ErrNotFound)
	}

	res := ACCESS{}
	content := DecodeKey(DbKey(), kvs[0].V)
	err := UnmarshalJson(content, &res)
	if err != nil {
		msg := fmt.Sprintf("LoadAccess [%v], data len %v\n%v\n", key, len(content), content)
		HandleInitErr(msg, err)
	}

	return res
}

func LoadRedisACCESS(conf EtcdConfig, key string, opts ...clientv3.OpOption) RedisACCESS {
	cli := MustGetEtcdClient(conf)
	defer cli.Close()

	kvs := GetKVFromEtcd(cli, AccessKey(key), opts...)
	if len(kvs) == 0 {
		msg := fmt.Sprintf("LoadAccess [%v]", key)
		HandleInitErr(msg, ErrNotFound)
	}

	res := RedisACCESS{}
	content := DecodeKey(DbKey(), kvs[0].V)
	err := UnmarshalJson(content, &res)
	if err != nil {
		msg := fmt.Sprintf("LoadAccess [%v], data len %v\n%v\n", key, len(content), content)
		HandleInitErr(msg, err)
	}

	return res
}

func EtcdCli(raw ...EtcdConfig) *clientv3.Client {
	return MustGetEtcdClient(SERVER().ETCD(raw...))
}

func GetEtcdContext(raw ...EtcdConfig) *EtcdContext {
	return NewEtcdContext(SERVER().ETCD(raw...))
}

func AccessKey(raw string) string {
	return etcdConfigAccess + EtcdDelimiter + raw
}

func MetaKey(raw string) string {
	return etcdTenantMeta + EtcdDelimiter + raw
}
