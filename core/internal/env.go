package internal

import (
	"context"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	_project = ""
	_tenant  = ""
	_env     = ""
	_inst    = ""
	_version = ""
	_space   = ""
	_start   int64
)

var (
	deploy  string
	host    string
	pid     int
	pidStr  string
	address string
	module  int32
	env     int32
	status  int32
)

var (
	GlobalWG      = new(sync.WaitGroup)
	GlobalContext context.Context
	GlobalCancel  context.CancelFunc
	TeardownJobs  = []func(){}
)

func BatchSet(raw ...string) {
	l := len(raw)

	toSet := []*string{
		&_project,
		&_tenant,
		&_inst,
		&_version,
	}

	n := l
	lt := len(toSet)
	if lt < n {
		n = lt
	}

	for i := 0; i < n; i++ {
		*toSet[i] = raw[i]
	}

	_space = _project + "." + _tenant

	pid = os.Getpid()
	pidStr = strconv.Itoa(pid)

	_start = time.Now().UnixMilli()
}

func SetProject(raw string) {
	_project = raw
}

func Project() string {
	return _project
}

func SetTenant(raw string) {
	_tenant = raw
}

func Tenant() string {
	return _tenant
}

func SetEnvName(raw string) {
	_env = raw
}

func EnvName() string {
	return _env
}

func Space() string {
	return _space
}

func StartAt() int64 {
	return _start
}

func SetInst(raw string) {
	_inst = raw
}

func Inst() string {
	return _inst
}

func SetVersion(raw string) {
	_version = raw
}

func Version() string {
	return _version
}

func SetDeploy(raw string) {
	deploy = raw
}

func Host() string {
	return host
}

func Pid() int {
	return pid
}

func PidStr() string {
	return pidStr
}

func SetModule(n int32) {
	module = n
}

func Module() int32 {
	return module
}

func SetEnv(n int32) {
	env = n
}

func Env() int32 {
	return env
}

func SetStatus(n int32) {
	atomic.StoreInt32(&status, n)
}

func Status() int32 {
	return atomic.LoadInt32(&status)
}

func SetAddress(raw string) {
	address = raw
}

func Address() string {
	return address
}

func FilePath(f string) string {
	if deploy == "" {
		initServerPath()
	}

	if strings.HasPrefix(f, deploy) {
		return f
	}

	return path.Join(deploy, f)
}
