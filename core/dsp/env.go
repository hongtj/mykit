package dsp

import (
	. "mykit/core/internal"
	. "mykit/core/types"
	"time"
)

func PROJECT() string {
	return Project()
}

func TENANT() string {
	return Tenant()
}

func EnvNAME() string {
	return EnvName()
}

func SPACE() string {
	return Space()
}

func HOST() string {
	return Host()
}

func PID() int {
	return Pid()
}

func INST() string {
	return Inst()
}

func VERSION() string {
	return Version()
}

func MODULE() int32 {
	return Module()
}

var RunningStatus = Status

func INITED() bool {
	return RunningStatus() >= StatusInited
}

func STOPPED() bool {
	return RunningStatus() <= StatusStopped
}

func CHECK(d ...time.Duration) {
	if len(d) > 0 {
		time.Sleep(d[0])
	}

	if !STOPPED() {
		return
	}

	time.Sleep(time.Hour * 999999)
}

func HOLD() {
	for {
		if !STOPPED() {
			return
		}

		time.Sleep(time.Millisecond * 200)
	}
}

func WaitForConsumer(d ...time.Duration) {
	interval := ParseTimeDuration(d, time.Millisecond*200)
	time.Sleep(interval)
}

var (
	GlobalUseMS = true
)
