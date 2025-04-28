package types

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

func ParseIntListFromStr(s string) []int {
	tmp := strings.Split(s, ",")
	return ParseIntListFromStrList(tmp)
}

func ParseIntListFromStrList(s []string) []int {
	res := []int{}
	for _, v := range s {
		n, err := strconv.Atoi(v)
		if err != nil {
			continue
		}

		res = append(res, n)
	}

	return res
}

func ParseInt64FromStr(s string) (i int64, err error) {
	return strconv.ParseInt(s, 10, 64)
}

func ParseInt64ToStr(num int64) string {
	return strconv.FormatInt(num, 10)
}

func ParseIntToStr(num int) string {
	return strconv.Itoa(num)
}

func ParseIntFromStr(s string) (i int, err error) {
	return strconv.Atoi(s)
}

func ParseInt64ListFromStrList(s []string) []int64 {
	res := []int64{}
	for _, v := range s {
		n, err := ParseInt64FromStr(v)
		if err != nil {
			continue
		}

		res = append(res, n)
	}

	return res
}

func ParseToInt64(raw interface{}) int64 {
	i, ok := raw.(int)
	if ok {
		return int64(i)
	}

	return raw.(int64)
}

func ParseFloat64FromStr(s string) (i float64, err error) {
	return strconv.ParseFloat(s, 64)
}

func ParseFloat64ListFromStr(s string) []float64 {
	tmp := strings.Split(s, ",")
	return ParseFloat64FromStrList(tmp)
}

func ParseFloat64FromStrList(s []string) []float64 {
	res := []float64{}
	for _, v := range s {
		n, err := ParseFloat64FromStr(v)
		if err != nil {
			continue
		}

		res = append(res, n)
	}

	return res
}

func ParseFloat64ToStr(num float64, decimal ...int) string {
	if len(decimal) == 0 {
		return strconv.FormatFloat(num, 'f', -1, 64)
	}

	num = TakeDigits(num, decimal[0])
	return strconv.FormatFloat(num, 'f', -1, 64)
}

func ParseToInterfaceSlice(raw ...interface{}) []interface{} {
	res := []interface{}{}

	for _, v := range raw {
		res = append(res, v)
	}

	return res
}

func ParseStrFromTag(raw interface{}) string {
	if raw == nil {
		return ""
	}

	v, ok := raw.(string)
	if !ok {
		return ""
	}

	return v
}

func ParseInt64FromTag(raw interface{}) int64 {
	if raw == nil {
		return 0
	}

	v, ok := raw.(string)
	if !ok {
		return 0
	}

	res, _ := ParseInt64FromStr(v)

	return res
}

func ParseInt64FromField(raw interface{}) int64 {
	if raw == nil {
		return 0
	}

	v, ok := raw.(json.Number)
	if !ok {
		return 0
	}

	res, _ := v.Int64()

	return res
}

func ParseFloat64FromField(raw interface{}) float64 {
	if raw == nil {
		return math.NaN()
	}

	v, ok := raw.(json.Number)
	if !ok {
		return math.NaN()
	}

	res, _ := v.Float64()

	return res
}

func ParseBoolToStr(raw bool) string {
	if raw {
		return StrOfTrue
	}

	return StrOfFalse
}

func ParseBoolFromStr(raw string) bool {
	return raw == StrOfTrue
}

type SomeId []string

func (t SomeId) Id() []int64 {
	res := []int64{}

	for _, v := range t {
		item, err := ParseInt64FromStr(v)
		if err == nil {
			res = append(res, item)
		}
	}

	return res
}

func ToSomeId(id ...int64) []string {
	res := []string{}

	for _, v := range id {
		item := ParseInt64ToStr(v)
		res = append(res, item)
	}

	return res
}
