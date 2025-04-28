package dsp

import (
	"fmt"
	. "mykit/core/types"
	"reflect"
	"strings"
	"time"
)

var (
	devDebugSwitch bool
)

func SetDevDebugSwitch(raw ...bool) {
	devDebugSwitch = ParseBool(raw)
}

func DevDebugSwitch() bool {
	return devDebugSwitch
}

func DevDebug(raw ...interface{}) {
	if !devDebugSwitch {
		return
	}

	for _, v := range raw {
		fmt.Println(OutputObj(v))
	}
}

func Debugs(raw ...interface{}) {
	if !devDebugSwitch {
		return
	}

	var b strings.Builder

	l := len(raw)
	for i := 0; i < l-1; i++ {
		b.WriteString("%v ")
	}

	b.WriteString("%v")

	DevDebug(fmt.Sprintf(b.String(), raw...))
}

func Debugf(format string, a ...interface{}) {
	if !devDebugSwitch {
		return
	}

	msg := fmt.Sprintf(format, a...)
	DevDebug(msg)
}

func Debuglf(l int, format string, a ...interface{}) {
	if !devDebugSwitch {
		return
	}

	msg := fmt.Sprintf(format, a...)
	if l > 0 {
		n := MinInt(len(msg), l)
		msg = msg[:n]
	}

	DevDebug(msg)
}

func NewDebugBlock(raw ...interface{}) {
	if !devDebugSwitch {
		return
	}

	DevDebug("\n")
	DevDebug(raw...)
}

func OutputObj(raw interface{}) string {
	if str, ok := raw.(string); ok {
		return str
	}

	if bytes, ok := raw.([]byte); ok {
		return BytesToString(bytes)
	}

	if slice, ok := raw.([]interface{}); ok {
		res := []string{}
		for _, element := range slice {
			res = append(res, ToJsonStr(element))
		}
		return strings.Join(res, "\n")
	}

	typeOf := reflect.TypeOf(raw)
	typeOf = DeType(typeOf)

	valueOf := reflect.ValueOf(raw)
	valueOf = DeValue(valueOf)

	if typeOf.Kind() == reflect.Struct {
		return ToJsonStr(raw)
	}

	if typeOf.Kind() == reflect.Slice || typeOf.Kind() == reflect.Array {
		l := valueOf.Len()
		if l == 0 {
			return StrOfNullList
		}

		res := []string{}
		for i := 0; i < l; i++ {
			res = append(res, ToJsonStr(valueOf.Index(i).Interface()))
		}
		return strings.Join(res, "\n")
	}

	if typeOf.Kind() == reflect.Map {
		if valueOf.Len() == 0 {
			return StrOfNullJson
		}

		res := []string{}
		for _, k := range valueOf.MapKeys() {
			res = append(res, fmt.Sprintf("%v: %v",
				k.Interface(), ToJsonStr(valueOf.MapIndex(k).Interface())))
		}
		return strings.Join(res, "\n")
	}

	return ToJsonStr(raw)
}

func PrintInfo(a ...interface{}) {
	if !devDebugSwitch {
		return
	}

	s := fmt.Sprintf("[info] %v: %v", time.Now().Format("2006-01-02 15:04:05.000"), a)
	fmt.Println(s)
}

func PrintError(a ...interface{}) {
	if !devDebugSwitch {
		return
	}

	s := fmt.Sprintf("[error] %v: %v", time.Now().Format("2006-01-02 15:04:05.000"), a)
	fmt.Println(s)
}
