package transfer

import (
	. "mykit/core/persist"
	. "mykit/core/types"
	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/client"
	cgrpc "github.com/micro/go-micro/v2/client/grpc"
	sgrpc "github.com/micro/go-micro/v2/server/grpc"
)

var (
	MaxMsgSize = 1024 * 1024 * 32
	rpcPrefix  = "srv.lixx."
)

func SetRpcPrefix(raw string) {
	rpcPrefix = raw
}

func InitRpcPrefix(prefix string, rpc ...*string) {
	SetRpcPrefix(prefix)

	AddPrefix(prefix, rpc...)
}

func RpcEndpoint(raw string) string {
	return rpcPrefix + raw
}

func RegistryServer(app, addr string, etcd EtcdConfig) micro.Service {
	service := micro.NewService(
		micro.RegisterInterval(time.Second*30),
		micro.Name(app),
		micro.Address(addr),
		micro.Registry(etcd.MicroRegistry()),
	)

	service.Init()

	service.Server().Init(
		sgrpc.MaxMsgSize(MaxMsgSize),
	)

	return service
}

func NewSmarterServer(app, addr string, etcd EtcdConfig, hdlr SmarterHandler) micro.Service {
	service := RegistryServer(app, addr, etcd)

	err := RegisterSmarterHandler(service.Server(), hdlr)
	HandleInitErr("NewSmarterServer", err)

	return service
}

func RegistryClient(app string, etcd EtcdConfig) micro.Service {
	service := micro.NewService(
		micro.RegisterInterval(time.Second*30),
		micro.Name(app),
		micro.Registry(etcd.MicroRegistry()),
	)

	service.Init()

	return service
}

func NewSmarterClient(app string, etcd EtcdConfig) SmarterService {
	service := RegistryClient(app, etcd)

	err := service.Client().Init(
		client.DialTimeout(60*time.Second),
		client.RequestTimeout(60*time.Second),
		cgrpc.MaxSendMsgSize(MaxMsgSize),
		cgrpc.MaxRecvMsgSize(MaxMsgSize),
	)
	HandleInitErr("Micro Client init", err)

	return NewSmarterService(app, service.Client())
}
