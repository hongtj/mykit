package smarter

import (
	"context"
	"fmt"
	. "mykit/core/dsp"
	. "mykit/core/persist"
	. "mykit/core/types"
	"strings"

	"github.com/go-redis/redis/v8"
)

type RedisConfig struct {
	Access         string
	Db             int    `json:",default=0"`
	Debug          bool   `json:",optional"`
	Model          string `json:",optional"`
	PoolSize       int    `json:",default=0"`
	DialTimeoutMs  uint64 `json:",default=0"`
	ReadTimeoutMs  uint64 `json:",default=0"`
	WriteTimeoutMs uint64 `json:",default=0"`
	IdleTimeoutMs  uint64 `json:",default=0"`
}

func (t RedisConfig) NewRedisConfig(etcd EtcdConfig) *redis.Options {
	access := LoadRedisACCESS(etcd, t.Access)

	option := &redis.Options{
		Addr:         access.Uri[0],
		Password:     access.Pwd,
		DB:           t.Db,
		PoolSize:     t.PoolSize,
		DialTimeout:  MsTimeout(t.DialTimeoutMs),
		ReadTimeout:  MsTimeout(t.ReadTimeoutMs),
		WriteTimeout: MsTimeout(t.WriteTimeoutMs),
		IdleTimeout:  MsTimeout(t.IdleTimeoutMs),
	}

	return option
}

func (t RedisConfig) OpenRedis(raw ...EtcdConfig) redis.Cmdable {
	etcd := SERVER().ETCD(raw...)
	return OpenRedis(etcd, t, t.Db)
}

var (
	redisCli = map[int]redis.Cmdable{}
)

func GetRedis(db int) redis.Cmdable {
	return redisCli[db]
}

func InitRedis(conf RedisConfig, db ...int) {
	if conf.Access == "" {
		return
	}

	db = SetIntList(db)
	if len(db) == 0 {
		db = append(db, conf.Db)
	}

	for _, v := range db {
		OpenRedis(SERVER().ETCD(), conf, v)
	}
}

func OpenRedis(etcd EtcdConfig, conf RedisConfig, db ...int) redis.Cmdable {
	conf.Db = ParseIntParam(db, 0)
	cli, ok := redisCli[conf.Db]
	if ok {
		return cli
	}

	cli = openRedis(etcd, conf)
	_, err := cli.Ping(Ctx).Result()
	HandleInitErr("redis ping", err)

	redisCli[conf.Db] = cli

	return cli
}

func openRedis(etcd EtcdConfig, conf RedisConfig) redis.Cmdable {
	mode := strings.ToLower(conf.Model)
	mode = strings.TrimSpace(mode)
	if mode == "client" || mode == "" {
		option := conf.NewRedisConfig(etcd)
		return redis.NewClient(option)
	}

	msg := fmt.Sprintf("unsupported redis mode [%v]", mode)
	HandleInitErr(msg, ErrInvalidParam)

	return nil
}

func UtRedis(conf RedisConfig) redis.Cmdable {
	return OpenRedis(SERVER().ETCD(), conf, 12)
}

func BatchHset(ctx context.Context, pipe redis.Pipeliner, raw ...HashObject) error {
	for _, v := range raw {
		pipe.HSet(ctx, v.Key(ctx), v.Values()...)
	}

	_, err := pipe.Exec(ctx)

	return err
}

func RedisKeyBuilder(ctx context.Context) *StrBuilder {
	b := NewStrBuilder()

	tenant := GetTenant(ctx)
	if tenant != "" {
		b.WriteString(tenant)
	} else {
		b.WriteString(redisPrefix)
	}

	return b
}
