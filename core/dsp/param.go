package dsp

import (
	. "mykit/core/types"
	"time"
)

func ParseSelector(param []Selector, w int) Selector {
	return ParseSelectorParam(param, NewFnvSelector(w))
}

func ParseStatusChecker(param []StatusChecker, v StatusChecker) StatusChecker {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func (t AlertMarkList) Len() int {
	return len(t)
}

func (t AlertMarkList) Less(i, j int) bool {
	return t[i].Level < t[j].Level
}

func (t AlertMarkList) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func GetTickParam(raw ...int64) int64 {
	if len(raw) == 0 {
		if GlobalUseMS {
			return time.Now().UnixMilli()
		} else {
			return time.Now().Unix()
		}

	} else {

		if GlobalUseMS {
			return GetTick(raw[0])
		} else {
			return GetTickMs(raw[0])
		}
	}
}

func GetTimeParam(raw ...time.Time) int64 {
	if len(raw) == 0 {
		if GlobalUseMS {
			return time.Now().UnixMilli()
		} else {
			return time.Now().Unix()
		}

	} else {

		if GlobalUseMS {
			return raw[0].UnixMilli()
		} else {
			return raw[0].Unix()
		}
	}
}

func EnsureTickParam(raw ...int64) int64 {
	if len(raw) == 0 {
		return time.Now().Unix()
	} else {
		return GetTick(raw[0])
	}
}

func EnsureTickMsParam(raw ...int64) int64 {
	if len(raw) == 0 {
		return time.Now().UnixMilli()
	} else {
		return GetTickMs(raw[0])
	}
}

func GetTick(raw int64) int64 {
	if raw > MinusMsTick {
		return raw / 1000
	} else {
		return raw
	}
}

func GetTickMs(raw int64) int64 {
	if raw < MinusMsTick {
		return raw * 1000
	} else {
		return raw
	}
}
