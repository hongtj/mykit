package gateway

import (
	core "mykit"
	"mykit/core/smarter"
)

type Config struct {
	core.Config
	smarter.CertPem
}

func (t Config) Setup() func() {
	f := t.Config.Setup()

	core.InitPem(&t.CertPem)

	return f
}

var (
	Conf = Config{}
)

func Init(f string) Config {
	smarter.MustLoad(f, &Conf)

	return Conf
}
