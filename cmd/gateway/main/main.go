package main

import (
	core "mykit"
	config "mykit/cmd/gateway"
	"mykit/core/smarter"
)

func main() {
	conf := config.Init(core.ConfigFile)
	defer Init(conf)()

	server.Run()

	smarter.TEARDOWN()
}
