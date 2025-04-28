package persist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	. "mykit/core/dsp"
	. "mykit/core/types"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
)

func IgnoreRedisErr(err error) bool {
	return err == redis.Nil
}

func RealRedisErr(err error) bool {
	if err == nil {
		return false
	}

	return !IgnoreRedisErr(err)
}

func MustRecvObj(desc string, f func(duration time.Duration) (res []string, err error),
	duration time.Duration, obj interface{}) (err error) {
	recv, err := f(duration)
	if err != nil {
		ZapFailed(LogS1,
			LogEvent("MustRecvObj"+"::"+FnName(f)),
			LogProcessor("recv"),
			LogContent(desc),
			LogError(err),
		)
		return err
	}

	var data []byte
	if len(recv) < 2 {
		err = errors.New("recv failed")
		return
	}

	data = StringToBytes(recv[1])
	if len(data) == 0 {
		err = errors.New("recv empty data")
		return
	}

	Debugf("data: %s", data)

	err = json.Unmarshal(data, &obj)

	if err != nil {
		ZapFailed(LogS1,
			LogEvent("MustRecvObj"+"::"+FnName(f)),
			LogProcessor("unmarshal"),
			LogContent(desc),
			LogError(err),
		)
	}

	return
}

func ReceiveByteFromRedis(cli redis.Cmdable, ctx context.Context, key string,
	l ...int) (res []byte, err error) {
	cmd := cli.BLPop(ctx, 0, key)
	recv, err := cmd.Result()
	if err != nil {
		return
	}

	if len(recv) < 2 {
		err = errors.New("recv failed")
		return
	}

	Debuglf(ParseIntParam(l, 0), "key: %s\n%s", key, recv[1])

	res = StringToBytes(recv[1])

	if len(res) == 0 {
		err = errors.New("recv empty data")
		return
	}

	return
}

func ReceiveObjFromRedis(cli redis.Cmdable, ctx context.Context, key string,
	obj interface{}, l ...int) (err error) {
	cmd := cli.BLPop(ctx, 0, key)
	recv, err := cmd.Result()
	if err != nil {
		return
	}

	if len(recv) < 2 {
		err = errors.New("recv failed")
		return
	}

	data := StringToBytes(recv[1])
	if len(data) == 0 {
		err = errors.New("recv empty data")
		return
	}

	Debuglf(ParseIntParam(l, 0), "key: %s\n%s", key, data)

	err = json.Unmarshal(data, &obj)

	return
}

func UpdateInt64WhenBigger(cli redis.Cmdable, ctx context.Context, key string, v int64) error {
	cmd := cli.Get(ctx, key)
	data, err := cmd.Int64()
	if err != nil {
		if err == redis.Nil {
			goto Set
		}

		return err
	}

	if v <= data {
		return nil
	}

Set:
	err = cli.Set(ctx, key, v, 0).Err()

	return err
}

func ScanKeys(cli redis.Cmdable, ctx context.Context, pattern string, c ...int64) (res []string) {
	res = []string{}

	var err error
	var cursor uint64
	var tmp []string
	count := ParseInt64Param(c, 300)
	loop := 0

	for {
		tmp, cursor, err = cli.Scan(ctx, cursor, pattern, count).Result()
		if err != nil {
			break
		}

		res = append(res, tmp...)
		loop++
		if cursor == 0 || loop > 1000 {
			break
		}
	}

	return
}

type RedisContext struct {
	ctx         context.Context
	client      redis.Cmdable
	isCluster   bool
	transaction bool
	step        *int
	*ZLogger
}

func NewRedisContext(ctx context.Context, db redis.Cmdable) *RedisContext {
	res := &RedisContext{}
	res.Init(ctx, db)

	return res
}

func (t *RedisContext) Init(ctx context.Context, client redis.Cmdable) {
	var step int

	t.ctx = ctx
	t.client = client
	_, t.isCluster = t.client.(*redis.ClusterClient)
	t.transaction = false
	t.step = &step

	//t.Logger = logx.WithContext(ctx)
	t.ZLogger = NewZLogger(ctx, LogMsgRedis)
}

func (t *RedisContext) Ctx() context.Context {
	return t.ctx
}

func (t *RedisContext) Db() redis.Cmdable {
	*t.step++
	return t.client
}

func (t *RedisContext) PipGet(key ...string) []*redis.StringCmd {
	return nil
}

func (t *RedisContext) PipGetString(key ...string) []string {
	res := []string{}

	d := t.PipGet(key...)

	item := ""
	var err error
	for _, v := range d {
		item, err = v.Result()
		if err != nil {
			res = append(res, "")
		} else {
			res = append(res, item)
		}
	}

	return res
}

func (t *RedisContext) FlushDB() error {
	return t.client.FlushDB(t.ctx).Err()
}

func (t *RedisContext) Mock() redismock.ClientMock {
	mockDb, mock := redismock.NewClientMock()
	t.client = mockDb
	return mock
}

func (t *RedisContext) Echo(message interface{}) *redis.StringCmd {
	cmd := t.Db().Echo(t.ctx, message)
	return cmd
}

func (t *RedisContext) Ping() *redis.StatusCmd {
	cmd := t.Db().Ping(t.ctx)
	return cmd
}

func (t *RedisContext) Keys(pattern string) *redis.StringSliceCmd {
	cmd := t.Db().Keys(t.ctx, pattern)
	return cmd
}

func (t *RedisContext) Del(keys ...string) *redis.IntCmd {
	cmd := t.Db().Del(t.ctx, keys...)
	return cmd
}

func (t *RedisContext) Exists(keys ...string) *redis.IntCmd {
	cmd := t.Db().Exists(t.ctx, keys...)
	return cmd
}

func (t *RedisContext) Save(key string, obj interface{}) error {
	return t.Set(key, MustJsonMarshal(obj), -1).Err()
}

func (t *RedisContext) Load(key string, obj interface{}) error {
	b, err := t.Get(key).Bytes()
	if err != nil || len(b) == 0 {
		return fmt.Errorf("load %v failed", key)
	}

	return json.Unmarshal(b, &obj)
}

func (t *RedisContext) UpdateInt64WhenBigger(key string, v int64) error {
	cmd := t.Get(key)
	o, err := cmd.Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			goto Set
		}

		detail := map[string]interface{}{
			"key":   key,
			"value": v,
		}

		t.Error(LogMsgRedis,
			LogEvent("UpdateInt64WhenBigger"),
			LogProcessor("get"),
			LogDetail(detail),
			LogError(err),
		)
		return err
	}

	if v <= o {
		return nil
	}

Set:
	err = t.Set(key, v, -1).Err()
	if err != nil {
		detail := map[string]interface{}{
			"key":   key,
			"value": v,
		}

		t.Error(LogMsgRedis,
			LogEvent("UpdateInt64WhenBigger"),
			LogProcessor("set"),
			LogDetail(detail),
			LogError(err),
		)
	}

	return err
}

func (t *RedisContext) Expire(key string, expiration time.Duration) *redis.BoolCmd {
	cmd := t.Db().Expire(t.ctx, key, expiration)
	return cmd
}

func (t *RedisContext) TTL(key string) *redis.DurationCmd {
	cmd := t.Db().TTL(t.ctx, key)
	return cmd
}

func (t *RedisContext) Get(key string) *redis.StringCmd {
	cmd := t.Db().Get(t.ctx, key)
	return cmd
}

func (t *RedisContext) Incr(key string) *redis.IntCmd {
	cmd := t.Db().Incr(t.ctx, key)
	return cmd
}

func (t *RedisContext) IncrBy(key string, value int64) *redis.IntCmd {
	cmd := t.Db().IncrBy(t.ctx, key, value)
	return cmd
}

func (t *RedisContext) IncrByFloat(key string, value float64) *redis.FloatCmd {
	cmd := t.Db().IncrByFloat(t.ctx, key, value)
	return cmd
}

func (t *RedisContext) MGet(keys ...string) *redis.SliceCmd {
	cmd := t.Db().MGet(t.ctx, keys...)
	return cmd
}

func (t *RedisContext) MSet(values ...interface{}) *redis.StatusCmd {
	cmd := t.Db().MSet(t.ctx, values...)
	return cmd
}

func (t *RedisContext) MSetNX(values ...interface{}) *redis.BoolCmd {
	cmd := t.Db().MSetNX(t.ctx, values...)
	return cmd
}

func (t *RedisContext) Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := t.Db().Set(t.ctx, key, value, expiration)
	return cmd
}

func (t *RedisContext) SCard(key string) *redis.IntCmd {
	cmd := t.Db().SCard(t.ctx, key)
	return cmd
}

func (t *RedisContext) Scan(cursor uint64, match string, count int64) *redis.ScanCmd {
	cmd := t.Db().Scan(t.ctx, cursor, match, count)
	return cmd
}

func (t *RedisContext) HDel(key string, fields ...string) *redis.IntCmd {
	cmd := t.Db().HDel(t.ctx, key, fields...)
	return cmd
}

func (t *RedisContext) HExists(key, field string) *redis.BoolCmd {
	cmd := t.Db().HExists(t.ctx, key, field)
	return cmd
}

func (t *RedisContext) HGet(key, field string) *redis.StringCmd {
	cmd := t.Db().HGet(t.ctx, key, field)
	return cmd
}

func (t *RedisContext) HGetAll(key string) *redis.StringStringMapCmd {
	cmd := t.Db().HGetAll(t.ctx, key)
	return cmd
}

func (t *RedisContext) HIncrBy(key, field string, incr int64) *redis.IntCmd {
	cmd := t.Db().HIncrBy(t.ctx, key, field, incr)
	return cmd
}

func (t *RedisContext) HIncrByFloat(key, field string, incr float64) *redis.FloatCmd {
	cmd := t.Db().HIncrByFloat(t.ctx, key, field, incr)
	return cmd
}

func (t *RedisContext) HKeys(key string) *redis.StringSliceCmd {
	cmd := t.Db().HKeys(t.ctx, key)
	return cmd
}

func (t *RedisContext) HLen(key string) *redis.IntCmd {
	cmd := t.Db().HLen(t.ctx, key)
	return cmd
}

func (t *RedisContext) HMGet(key string, fields ...string) *redis.SliceCmd {
	cmd := t.Db().HMGet(t.ctx, key, fields...)
	return cmd
}

func (t *RedisContext) HSet(key string, values ...interface{}) *redis.IntCmd {
	cmd := t.Db().HSet(t.ctx, key, values...)
	return cmd
}

func (t *RedisContext) HMSet(key string, values ...interface{}) *redis.BoolCmd {
	cmd := t.Db().HMSet(t.ctx, key, values...)
	return cmd
}

func (t *RedisContext) BRPop(timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	cmd := t.Db().BRPop(t.ctx, timeout, keys...)
	return cmd
}

func (t *RedisContext) LPush(key string, values ...interface{}) *redis.IntCmd {
	cmd := t.Db().LPush(t.ctx, key, values...)
	return cmd
}

func (t *RedisContext) RPop(key string) *redis.StringCmd {
	cmd := t.Db().RPop(t.ctx, key)
	return cmd
}

func (t *RedisContext) LIndex(key string, index int64) *redis.StringCmd {
	cmd := t.Db().LIndex(t.ctx, key, index)
	return cmd
}

func (t *RedisContext) LLen(key string) *redis.IntCmd {
	cmd := t.Db().LLen(t.ctx, key)
	return cmd
}

func (t *RedisContext) LPop(key string) *redis.StringCmd {
	cmd := t.Db().LPop(t.ctx, key)
	return cmd
}

func (t *RedisContext) LRem(key string, count int64, value interface{}) *redis.IntCmd {
	cmd := t.Db().LRem(t.ctx, key, count, value)
	return cmd
}

func (t *RedisContext) RPush(key string, values ...interface{}) *redis.IntCmd {
	cmd := t.Db().RPush(t.ctx, key, values...)
	return cmd
}

func (t *RedisContext) ZRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd {
	return t.ZRangeArgsWithScores(redis.ZRangeArgs{
		Key:   key,
		Start: start,
		Stop:  stop,
	})
}

func (t *RedisContext) ZRangeArgsWithScores(z redis.ZRangeArgs) *redis.ZSliceCmd {
	cmd := t.Db().ZRangeArgsWithScores(t.ctx, z)
	return cmd
}

func (t *RedisContext) SAdd(key string, members ...interface{}) *redis.IntCmd {
	cmd := t.Db().SAdd(t.ctx, key, members...)
	return cmd
}

func (t *RedisContext) SRem(key string, members ...interface{}) *redis.IntCmd {
	cmd := t.Db().SRem(t.ctx, key, members...)
	return cmd
}

func (t *RedisContext) SIsMember(key string, member interface{}) *redis.BoolCmd {
	cmd := t.Db().SIsMember(t.ctx, key, member)
	return cmd
}

func (t *RedisContext) SMIsMember(key string, members ...interface{}) *redis.BoolSliceCmd {
	cmd := t.Db().SMIsMember(t.ctx, key, members...)
	return cmd
}

func (t *RedisContext) SMembers(key string) *redis.StringSliceCmd {
	cmd := t.Db().SMembers(t.ctx, key)
	return cmd
}

func (t *RedisContext) XAdd(a *redis.XAddArgs) *redis.StringCmd {
	cmd := t.Db().XAdd(t.ctx, a)
	return cmd
}

func (t *RedisContext) XPut(stream string, item map[string]interface{}) *redis.StringCmd {
	return t.XAdd(&redis.XAddArgs{Stream: stream, Values: item})
}

func (t *RedisContext) XGet(stream string, n ...int64) *redis.XMessageSliceCmd {
	count := ParseInt64Param(n, 1)
	return t.XRangeN(stream, "-", "+", count)
}

func (t *RedisContext) XRange(stream, start, stop string) *redis.XMessageSliceCmd {
	cmd := t.Db().XRange(t.ctx, stream, start, stop)
	return cmd
}

func (t *RedisContext) XRangeN(stream, start, stop string, count int64) *redis.XMessageSliceCmd {
	cmd := t.Db().XRangeN(t.ctx, stream, start, stop, count)
	return cmd
}

func (t *RedisContext) XRevRange(stream, start, stop string) *redis.XMessageSliceCmd {
	cmd := t.Db().XRevRange(t.ctx, stream, start, stop)
	return cmd
}

func (t *RedisContext) XRevRangeN(stream, start, stop string, count int64) *redis.XMessageSliceCmd {
	cmd := t.Db().XRevRangeN(t.ctx, stream, start, stop, count)
	return cmd
}

func (t *RedisContext) XRead(stream string, group string, n ...int64) *redis.XStreamSliceCmd {
	count := ParseInt64Param(n, 1)
	return t.XReadGroup(&redis.XReadGroupArgs{Group: group, Consumer: "Consumer", Streams: []string{stream, ">"}, Count: count, NoAck: true})
}

func (t *RedisContext) XReadGroup(a *redis.XReadGroupArgs) *redis.XStreamSliceCmd {
	cmd := t.Db().XReadGroup(t.ctx, a)
	return cmd
}

func (t *RedisContext) XAck(stream, group string, ids ...string) *redis.IntCmd {
	cmd := t.Db().XAck(t.ctx, stream, group, ids...)
	return cmd
}

func (t *RedisContext) XPending(stream, group string) *redis.XPendingCmd {
	cmd := t.Db().XPending(t.ctx, stream, group)
	return cmd
}

func (t *RedisContext) ZAdd(key string, members ...*redis.Z) *redis.IntCmd {
	cmd := t.Db().ZAdd(t.ctx, key, members...)
	return cmd
}

func (t *RedisContext) ZCard(key string) *redis.IntCmd {
	cmd := t.Db().ZCard(t.ctx, key)
	return cmd
}

func (t *RedisContext) ZIncrBy(key string, increment float64, member string) *redis.FloatCmd {
	cmd := t.Db().ZIncrBy(t.ctx, key, increment, member)
	return cmd
}

func (t *RedisContext) ZRange(key string, start, stop int64) *redis.StringSliceCmd {
	cmd := t.Db().ZRange(t.ctx, key, start, stop)
	return cmd
}

func (t *RedisContext) ZRangeByScore(key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	cmd := t.Db().ZRangeByScore(t.ctx, key, opt)
	return cmd
}

func (t *RedisContext) ZRank(key, member string) *redis.IntCmd {
	cmd := t.Db().ZRank(t.ctx, key, member)
	return cmd
}

func (t *RedisContext) ZRem(key string, members ...interface{}) *redis.IntCmd {
	cmd := t.Db().ZRem(t.ctx, key, members...)
	return cmd
}

func (t *RedisContext) ZRemRangeByRank(key string, start, stop int64) *redis.IntCmd {
	cmd := t.Db().ZRemRangeByRank(t.ctx, key, start, stop)
	return cmd
}

func (t *RedisContext) ZRemRangeByScore(key, min, max string) *redis.IntCmd {
	cmd := t.Db().ZRemRangeByScore(t.ctx, key, min, max)
	return cmd
}

func (t *RedisContext) ZRevRange(key string, start, stop int64) *redis.StringSliceCmd {
	cmd := t.Db().ZRevRange(t.ctx, key, start, stop)
	return cmd
}

func (t *RedisContext) ZRevRangeByScoreWithScores(key string, opt *redis.ZRangeBy) *redis.ZSliceCmd {
	cmd := t.Db().ZRevRangeByScoreWithScores(t.ctx, key, opt)
	return cmd
}

func (t *RedisContext) ZRevRank(key, member string) *redis.IntCmd {
	cmd := t.Db().ZRevRank(t.ctx, key, member)
	return cmd
}

func (t *RedisContext) ZScore(key, member string) *redis.FloatCmd {
	cmd := t.Db().ZScore(t.ctx, key, member)

	return cmd
}

// 输入min、max，从key(zset)获取元素列表
func (t *RedisContext) ZRangeByScoreRes(key string, min, max int64) []string {
	q := &redis.ZRangeBy{
		Min: ParseInt64ToStr(min),
		Max: ParseInt64ToStr(max),
	}

	cmd := t.ZRangeByScore(key, q)
	res, err := cmd.Result()
	if err != nil {
		t.Error(LogMsgRedis,
			LogEvent("ZRangeByScoreRes"),
			LogProcessor(cmd.String()),
			LogError(err),
		)

		return []string{}
	}

	return res
}

// 输入起始时间戳、结束时间戳，从key获取元素列表
func (t *RedisContext) ZRangeByTime(key string, start, end time.Time) []string {
	min := start.Unix()
	max := end.Unix()

	return t.ZRangeByScoreRes(key, min, max)
}

// 输入时间戳、过期时间，从key获取元素列表
func (t *RedisContext) ZRangeByTimeWithExpire(key string, end time.Time, expire time.Duration) []string {
	start := end.Add(-1 * expire)
	return t.ZRangeByTime(key, start, end)
}

// 输入start, stop，从key(zset)获取排行榜前n个元素列表
func (t *RedisContext) ZRevRangeRes(key string, start, stop int64) []string {
	cmd := t.ZRevRange(key, start, stop)
	res, err := cmd.Result()
	if err != nil {
		t.Error(LogMsgRedis,
			LogEvent("ZRevRangeRes"),
			LogProcessor(cmd.String()),
			LogError(err),
		)

		return []string{}
	}

	return res
}

// 从key(zset)获取排行榜前n个元素列表
func (t *RedisContext) ZRevRangeByN(key string, n int64) []string {
	return t.ZRevRangeRes(key, 0, n-1)
}

type RedisStream struct {
	Group    string
	Stream   []string
	Timeout  time.Duration
	Consumer string
	*RedisContext
}

func NewRedisStream(ctx context.Context, stream, group, consumer string, db redis.Cmdable) *RedisStream {
	return NewRedisStreamWithContext(stream, group, consumer, NewRedisContext(ctx, db))
}

func NewRedisStreamWithContext(stream, group, consumer string, r *RedisContext) *RedisStream {
	res := &RedisStream{
		Group:        group,
		Stream:       []string{stream, ">"},
		Timeout:      0,
		Consumer:     consumer,
		RedisContext: r,
	}

	return res
}

func NewRedisStreamWithTimeout(ctx context.Context, stream, group, consumer string, d time.Duration, db redis.Cmdable) *RedisStream {
	res := NewRedisStream(ctx, stream, group, consumer, db)
	res.SetTimeout(d)

	return res
}

func (t *RedisStream) SetTimeout(d time.Duration) {
	t.Timeout = d
}

func (t *RedisStream) Read(n ...int64) *redis.XStreamSliceCmd {
	count := ParseInt64Param(n, 1)
	a := &redis.XReadGroupArgs{
		Group:    t.Group,
		Consumer: t.Consumer,
		Streams:  t.Stream,
		Count:    count,
	}
	return t.XReadGroup(a)
}

func (t *RedisStream) Pending() *redis.XPendingCmd {
	return t.XPending(t.Stream[0], t.Group)
}

func (t *RedisStream) Ack(ids ...string) *redis.IntCmd {
	return t.XAck(t.Stream[0], t.Group, ids...)
}
