package main

import (
	core "mykit"
	config "mykit/cmd/gateway"
	"mykit/cmd/gateway/router"
	"mykit/core/smarter"
	"mykit/core/transfer"
)

var (
	server  *transfer.GinServer
	release bool
)

func Init(conf config.Config) func() {
	f := conf.Setup()

	release = !smarter.DevEnv()

	dbInit(conf.Server)

	initRpc(conf)

	initApi(conf)

	return f
}

func dbInit(conf smarter.Server) {
	core.InitMysql(conf.Mysql)
	core.InitRedis(conf.Redis)

}

func initRpc(conf config.Config) {
	conf.Add( //注册rpc节点

	)
}

func initApi(conf config.Config) {

	server = conf.Gin.NewServer(release)

	server.Reg(router.Init)
}
