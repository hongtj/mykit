package core

import (
	"mykit/core/dsp"
	"mykit/core/smarter"
	"mykit/core/types"
)

type Config struct {
	smarter.Server
}

func (t Config) Setup() func() {
	dsp.SetDevDebugSwitch(t.OutputDetail)

	types.SetStrValue(&Tenant, t.Zone)
	types.SetStrValue(&App, t.App)
	types.SetStrValue(&AppVersion, t.Version)

	return SETUP(t.Server)
}

func init() {
}

func InitPem(pem *smarter.CertPem) {
	if smarter.DevEnv() {
		return
	}

	pem.LoadPem()
	pem.Verify()

	CertPem = *pem
}
