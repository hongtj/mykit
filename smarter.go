package core

import (
	"mykit/core/smarter"
	"mykit/core/types"
)

func InitSmarter(param ...string) {
	if len(param) >= 4 {
		types.SetStrValue(&Project, param[0])
		types.SetStrValue(&Tenant, param[1])
		types.SetStrValue(&App, param[2])
		types.SetStrValue(&AppVersion, param[3])
	}

	smarter.InitApp(
		Project,
		Tenant,
		App,
		AppVersion,
	)
}

func SETUP(conf smarter.Server) func() {
	InitSmarter()

	f := smarter.SETUP(conf)

	InitRpc("hong")

	return f
}
