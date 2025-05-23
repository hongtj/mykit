package redis

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

// ConnPool redis连接池
type ConnPool struct {
	redisPool *redis.Pool
}

// InitRedisPool 初始化
func InitRedisPool(host, pwd string, dbIndex, maxIdle, maxActive, idleTimeout, connTimeout int) *ConnPool {
	rcp := &ConnPool{}
	rcp.redisPool = &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: time.Duration(idleTimeout) * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host,
				redis.DialPassword(pwd),
				redis.DialDatabase(dbIndex),
				redis.DialConnectTimeout(time.Duration(connTimeout)*time.Second),
				redis.DialReadTimeout(time.Duration(connTimeout)*time.Second),
				redis.DialWriteTimeout(time.Duration(connTimeout)*time.Second))

			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}

	return rcp
}

/*############################## Key ##############################*/

// ExistsKey for key
func (rc *ConnPool) ExistsKey(key string) (bool, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.Bool(conn.Do("EXISTS", key))
}

// DelKey for key
func (rc *ConnPool) DelKey(key string) (interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return conn.Do("DEL", key)
}

// ExpireKey for key
func (rc *ConnPool) ExpireKey(key string, seconds int) (interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return conn.Do("EXPIRE", key, seconds)
}

/*############################## String ##############################*/

// Get for string
func (rc *ConnPool) Get(key string) (string, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.String(conn.Do("GET", key))
}

// Set for string
func (rc *ConnPool) Set(key string, value string) (interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return conn.Do("SET", key, value)
}

// SetExpire for string
func (rc *ConnPool) SetExpire(key string, value string, seconds int) (interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return conn.Do("SET", key, value, "EX", seconds)
}

// Incr for string
func (rc *ConnPool) Incr(key string) (int64, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.Int64(conn.Do("INCR", key))
}

// Decr for string
func (rc *ConnPool) Decr(key string) (int64, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.Int64(conn.Do("DECR", key))
}

/*############################## Hash ##############################*/

// Hset for Hash
func (rc *ConnPool) Hset(key string, field string, value string) (interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return conn.Do("HSET", key, field, value)
}

// Hget for Hash
func (rc *ConnPool) Hget(key string, field string) (string, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.String(conn.Do("HGET", key, field))
}

// Hmset for Hash
func (rc *ConnPool) Hmset(key string, fieldValue map[string]string) (interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return conn.Do("HMSET", redis.Args{}.Add(key).AddFlat(fieldValue)...)
}

// Hmget for Hash
func (rc *ConnPool) Hmget(key string, field []string) ([]interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.Values(conn.Do("HMGET", redis.Args{}.Add(key).AddFlat(field)...))
}

// Hlen for Hash
func (rc *ConnPool) Hlen(key string) (int, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.Int(conn.Do("HLEN", key))
}

// Hexists for Hash
func (rc *ConnPool) Hexists(key string, field string) (bool, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.Bool(conn.Do("HEXISTS", key, field))
}

// Hdel for Hash
func (rc *ConnPool) Hdel(key string, field []string) (interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return conn.Do("HDEL", key, field)
}

// Hdel for OneHash
func (rc *ConnPool) HdelOne(key string, field string) (interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return conn.Do("HDEL", key, field)
}

// Hgetall for Hash
func (rc *ConnPool) Hgetall(key string) (map[string]string, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.StringMap(conn.Do("HGETALL", key))
}

/*############################## List ##############################*/

// Rpush for List
func (rc *ConnPool) Rpush(key string, value []string) (interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return conn.Do("RPUSH", redis.Args{}.Add(key).AddFlat(value)...)
}

// Lpop for List
func (rc *ConnPool) Lpop(key string) (string, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.String(conn.Do("LPOP", key))
}

// Llen for List
func (rc *ConnPool) Llen(key string) (int, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.Int(conn.Do("LLEN", key))
}

/*############################## Set ##############################*/

// Sadd for Set
func (rc *ConnPool) Sadd(key string, member []string) (interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return conn.Do("SADD", redis.Args{}.Add(key).AddFlat(member)...)
}

// Smembers for Set
func (rc *ConnPool) Smembers(key string) ([]interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.Values(conn.Do("SMEMBERS", key))
}

// Srem for Set
func (rc *ConnPool) Srem(key string, member []string) (interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return conn.Do("SREM", key, member)
}

// Scard for Set
func (rc *ConnPool) Scard(key string) (int, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.Int(conn.Do("SCARD", key))
}

/*############################## SortedSet ##############################*/

// Zadd for SortedSet
func (rc *ConnPool) Zadd(key string, score int64, member string) (interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return conn.Do("ZADD", key, score, member)
}

// Zrange for SortedSet
func (rc *ConnPool) Zrange(key string, start, stop int) ([]interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.Values(conn.Do("ZRANGE", key, start, stop))
}

// Zrevrange for SortedSet
func (rc *ConnPool) Zrevrange(key string, start, stop int) ([]interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.Values(conn.Do("ZREVRANGE", key, start, stop))
}

// Zcard for SortResponseBodyedSet
func (rc *ConnPool) Zcard(key string) (int, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return redis.Int(conn.Do("ZCARD", key))
}

// Zrem for SortedSet
func (rc *ConnPool) Zrem(key string, member string) (interface{}, error) {
	conn := rc.redisPool.Get()
	defer conn.Close()
	return conn.Do("ZREM", key, member)
}
