package smarter

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	. "mykit/core/dsp"
	. "mykit/core/internal"
	. "mykit/core/persist"
	. "mykit/core/transfer"
	. "mykit/core/types"
	"time"
)

type HeartBeat struct {
	Project string `json:"Project,omitempty"`
	Tenant  string `json:"Tenant,omitempty"`
	Scn     string `json:"Scn,omitempty"`
	Host    string
	Path    string
	Pid     int
	Start   int64
	Inst    string
	Version string
	Address string
	Tick    int64
	Status  int32
}

func NewHeartBeat() HeartBeat {
	res := HeartBeat{
		Project: PROJECT(),
		Tenant:  TENANT(),
		Scn:     "",
		Host:    HOST(),
		Path:    FilePath(""),
		Pid:     PID(),
		Start:   StartAt(),
		Inst:    INST(),
		Version: Version(),
		Address: SERVER().Address(),
		Tick:    StartAt(),
		Status:  RunningStatus(),
	}

	return res
}

func (t HeartBeat) Payload() string {
	return ToJsonStr(t)
}

func (t HeartBeat) MetaKey() string {
	return MetaKey(t.Inst + EtcdDelimiter + MD5(t.Path))
}

func (t HeartBeat) Param(checker StatusChecker) []byte {
	t.Tick = time.Now().UnixMilli()
	t.Status = checker()

	return MustJsonMarshal(t)
}

func (t HeartBeat) Send(checker StatusChecker, srv string) {
	client := SmarterGrpcClient(srv)

	req := &Req{
		App:    DCS,
		Method: HEARTBEAT,
	}

	go func() {
		defer Recover("send heartbeat")
		var err error

		fmt.Println("start heartbeat to", srv)

		for {
			req.Param = t.Param(checker)
			Debugs(time.Now(), string(req.Param))

			_, err = client.Call(GetTracedContext(), req)
			if err != nil {
				fmt.Println(err)
			}

			time.Sleep(time.Second * 10)
		}
	}()
}

func (t HeartBeat) Heartbeat(srv []string, c ...StatusChecker) {
	addr := ParseStrParam(srv, ENV().Dcs)
	if addr == "" {
		HandleInitErr("dcs server address is null", ErrInvalidParam)
	}

	checker := ParseStatusChecker(c, defaultStatusChecker)

	t.Send(checker, addr)
}

func StartHeartBeat(srv ...string) {
	res := NewHeartBeat()

	res.Heartbeat(srv)
}

func (t Server) Address() string {
	return t.local
}

func (t Server) Public() string {
	return t.public
}

func (t Server) Port() int {
	if t.ip == nil {
		return -1
	}

	return t.ip.Port
}

func (t *Server) putMeta() {
	if !t.StaticMeta {
		return
	}

	cli := EtcdCli()
	defer cli.Close()

	meta := NewHeartBeat()
	cli.Put(Ctx, meta.MetaKey(), meta.Payload())
}

func (t Server) ETCD(raw ...EtcdConfig) EtcdConfig {
	return ParseEtcdConfig(raw, t.Etcd)
}

func (t Server) MYSQL(raw ...MysqlConfig) MysqlConfig {
	return ParseMysqlConfig(raw, t.Mysql)
}

func (t Server) REDIS(raw ...RedisConfig) RedisConfig {
	return ParseRedisConfig(raw, t.Redis)
}

//md5加密
func MD5(data string) string {
	m := md5.Sum([]byte(data))
	return hex.EncodeToString(m[:])
}
