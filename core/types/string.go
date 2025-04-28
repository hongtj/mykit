package types

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// 将字符串数组转化为逗号分割的字符串形式  ["str1","str2","str3"] >>> "str1,str2,str3"
func StrListToString(strList []string) (str string) {
	if len(strList) > 0 {
		for k, v := range strList {
			if k == 0 {
				str = v
			} else {
				str = str + "," + v
			}
		}
		return
	}
	return ""
}

//检查字符串是否是MD5加密字符串
func CheckIsMd5(str string) bool {
	if len(str) != 32 {
		return false
	}

	for _, runeValue := range str {
		if (runeValue >= 48 && runeValue <= 57) || //0-9
			(runeValue >= 97 && runeValue <= 102) { //a-f
			continue
		}

		return false
	}

	return true
}

//获取字符串文字长度（单字节字符计1长度，其他计2长度）
func GetStrWordLength(str string) int {
	result := 0
	if str == "" {
		return result
	}

	for _, runeValue := range str {
		if len(string(runeValue)) == 1 {
			result = result + 1
		} else {
			result = result + 2
		}
	}

	return result
}

func FirstLower(s string) string {
	if s == "" {
		return ""
	}

	return strings.ToLower(s[:1]) + s[1:]
}

func ExistNullStr(raw ...string) bool {
	for _, v := range raw {
		if v == "" {
			return true
		}
	}

	return false
}

func CheckNullStr(raw ...string) {
	for i, v := range raw {
		if v == "" {
			msg := fmt.Sprintf("Variable at index %d is null", i+1)
			panic(msg)
		}
	}
}

func CheckNullStrParam(raw string, param ...string) {
	for i, v := range param {
		if v == "" {
			msg := fmt.Sprintf("%v | null str found, %v", raw, i)
			panic(msg)
		}
	}
}

func AddPrefix(prefix string, raw ...*string) {
	for _, v := range raw {
		if v == nil {
			continue
		}

		*v = prefix + *v
	}
}

type StrBuilder struct {
	*strings.Builder
}

func NewStrBuilder() *StrBuilder {
	var b strings.Builder

	res := &StrBuilder{
		Builder: &b,
	}

	return res
}

func (t *StrBuilder) WriteInt64(n int64) {
	t.Builder.WriteString(ParseInt64ToStr(n))
}

func (t *StrBuilder) WriteInt(n int) {
	t.Builder.WriteString(strconv.Itoa(n))
}

func BatchSetSomeStr(p []*string, raw ...string) {
	l := MinInt(len(p), len(raw))

	for i := 0; i < l; i++ {
		*p[i] = raw[i]
	}
}

func (t PluginConf) Apply(p []*string) {
	BatchSetSomeStr(p, t...)
}
