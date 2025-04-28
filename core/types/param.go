package types

import (
	"context"
	"regexp"
	"strings"
	"time"
)

func DeStrParam(param, v string) string {
	if param == "" {
		return v
	}

	return param
}

func DeByteListParam(param, v []byte) []byte {
	if len(param) == 0 {
		return v
	}

	return param
}

func DeIntParam(param, v int) int {
	if param == 0 {
		return v
	}

	return param
}

func DeInt32Param(param, v int32) int32 {
	if param == 0 {
		return v
	}

	return param
}

func DeInt64Param(param, v int64) int64 {
	if param == 0 {
		return v
	}

	return param
}

func PadPrefix(param *string, prefix string) {
	if strings.HasPrefix(*param, prefix) {
		return
	}

	*param = prefix + *param
}

func PadSuffix(param *string, suffix string) {
	if strings.HasSuffix(*param, suffix) {
		return
	}

	*param += suffix
}

func ParseStrParam(param []string, v string) string {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ParseStrParams(param []string, v []string) []string {
	if len(param) == 0 {
		return v
	}

	return param
}

func ParseUint16Param(param []uint16, v uint16) uint16 {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ParseIntParam(param []int, v int) int {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ParseInt32Param(param []int32, v int32) int32 {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ParseInt64Param(param []int64, v int64) int64 {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ParseUint64Param(param []uint64, v uint64) uint64 {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ParseFloat32Param(param []float32, v float32) float32 {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ParseFloat64Param(param []float64, v float64) float64 {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ParseBoolParam(param []bool, v bool) bool {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ParseBool(raw []bool) bool {
	return ParseBoolParam(raw, true)
}

func ParseErrorParam(param []error, v error) error {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ParseContextParam(param []context.Context) context.Context {
	if len(param) == 0 {
		return context.Background()
	}

	return param[0]
}

func ParseTimeParam(param []time.Time, v time.Time) time.Time {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ParseTime(param []time.Time) time.Time {
	return ParseTimeParam(param, time.Now())
}

func ParseTimeDuration(param []time.Duration, v time.Duration) time.Duration {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func ParseTick(param []int64) int64 {
	return ParseInt64Param(param, time.Now().Unix())
}

func ParseTickMs(param []int64) int64 {
	return ParseInt64Param(param, time.Now().UnixMilli())
}

func ParseTaskIntervalParam(raw []TaskInterval, v TaskInterval) TaskInterval {
	if len(raw) == 0 {
		return v
	}

	return raw[0]
}

func ParseTaskInterval(raw []TaskInterval) TaskInterval {
	return ParseTaskIntervalParam(raw, func() time.Duration { return 0 })
}

func ParseSelectorParam(param []Selector, v Selector) Selector {
	if len(param) == 0 {
		return v
	}

	return param[0]
}

func CheckPhoneNumber(phone string) bool {
	matched, _ := regexp.MatchString(`^1[3456789]\d{9}$`, phone)

	return matched
}

func ParseParamTo(to int64) []int64 {
	if to == 0 {
		return []int64{}
	}

	return []int64{to}
}
