package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func ToIntFlag(raw []int) map[int]bool {
	res := map[int]bool{}

	for _, v := range raw {
		res[v] = true
	}

	return res
}

func ToInt32Flag(raw []int32) map[int32]bool {
	res := map[int32]bool{}

	for _, v := range raw {
		res[v] = true
	}

	return res
}

func ToInt64Flag(raw []int64) map[int64]bool {
	res := map[int64]bool{}

	for _, v := range raw {
		res[v] = true
	}

	return res
}

func ToStrListFlag(raw []string) map[string]bool {
	res := map[string]bool{}

	for _, v := range raw {
		res[v] = true
	}

	return res
}

func ToIntCount(raw []int) map[int]int {
	res := map[int]int{}

	for _, v := range raw {
		res[v]++
	}

	return res
}

func ToInt32Count(raw []int32) map[int32]int {
	res := map[int32]int{}

	for _, v := range raw {
		res[v]++
	}

	return res
}

func ToInt64Count(raw []int64) map[int64]int {
	res := map[int64]int{}

	for _, v := range raw {
		res[v]++
	}

	return res
}

func ToStrListCount(raw []string) map[string]int {
	res := map[string]int{}

	for _, v := range raw {
		res[v]++
	}

	return res
}

func ToIntCounter(raw []int) map[int]int {
	res := map[int]int{}

	for _, v := range raw {
		res[v] = 0
	}

	return res
}

func ToInt32Counter(raw []int32) map[int32]int {
	res := map[int32]int{}

	for _, v := range raw {
		res[v] = 0
	}

	return res
}

func ToInt64Counter(raw []int64) map[int64]int {
	res := map[int64]int{}

	for _, v := range raw {
		res[v] = 0
	}

	return res
}

func ToStrListCounter(raw []string) map[string]int {
	res := map[string]int{}

	for _, v := range raw {
		res[v] = 0
	}

	return res
}

//sort map by key
func SortMapByKey2Str(m map[string]interface{}) string {
	// To store the keys in slice in sorted order
	var keys []string
	var s string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// To perform the opertion you want
	for _, k := range keys {
		if m[k] != nil {
			s += k + "=" + fmt.Sprint(m[k]) + "&"
		}
	}
	return strings.TrimSuffix(s, "&")
}

func Struct2Map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	data := map[string]interface{}{}

	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Name] = v.Field(i).Interface()
	}
	return data
}

func ParseToMapStrInterface(r interface{}) map[string]interface{} {
	res := map[string]interface{}{}

	switch raw := r.(type) {
	case map[string]string:
		for k, v := range raw {
			res[k] = v
		}
	case map[string]int32:
		for k, v := range raw {
			res[k] = v
		}
	case map[string]int64:
		for k, v := range raw {
			res[k] = v
		}
	case map[string]float32:
		for k, v := range raw {
			res[k] = v
		}
	case map[string]float64:
		for k, v := range raw {
			res[k] = v
		}
	default:
		msg := fmt.Sprintf("not support: %v", raw)
		panic(msg)
	}

	return res
}

func ParseToMapStrMapStrInterface(raw interface{}) map[string]map[string]interface{} {
	res := map[string]map[string]interface{}{}

	switch obj := raw.(type) {
	case map[string]map[string]string:
		for k, v := range obj {
			res[k] = ParseToMapStrInterface(v)
		}
	case map[string]map[string]int32:
		for k, v := range obj {
			res[k] = ParseToMapStrInterface(v)
		}
	case map[string]map[string]int64:
		for k, v := range obj {
			res[k] = ParseToMapStrInterface(v)
		}
	case map[string]map[string]float32:
		for k, v := range obj {
			res[k] = ParseToMapStrInterface(v)
		}
	case map[string]map[string]float64:
		for k, v := range obj {
			res[k] = ParseToMapStrInterface(v)
		}
	default:
		msg := fmt.Sprintf("not support: %v", obj)
		panic(msg)
	}

	return res
}

func SumMapStrInt32(raw map[string]int32) int32 {
	var sum int32
	for _, v := range raw {
		sum += v
	}

	return sum
}

func SumMapStrFloat64(raw map[string]float64) float64 {
	var sum float64
	for _, v := range raw {
		sum += v
	}

	return sum
}

func DiffMapStrInterface(m1, m2 map[string]interface{}) (res []string) {
	checked := map[string]bool{}

	for k := range m1 {
		checked[k] = true
		if !reflect.DeepEqual(m1[k], m2[k]) {
			res = append(res, k)
		}
	}

	for k := range m2 {
		if !checked[k] {
			res = append(res, k)
		}
	}

	return res
}

func SetMapStringStringFromMapStringInterface(src map[string]interface{}, dst map[string]string) {
	for k, v := range src {
		item, ok := v.(string)
		if ok {
			dst[k] = item
		}
	}
}

type Param map[string]interface{}

func (t Param) Set(k string, v interface{}) {
	t[k] = v
}

func (t Param) Get(k string) interface{} {
	return t[k]
}

func (t Param) GetStr(k string) string {
	raw := t.Get(k)
	v, ok := raw.(string)
	if ok {
		return v
	}

	num, ok := raw.(json.Number)
	if ok {
		return num.String()
	}

	return ""
}

func (t Param) GetBool(k string) bool {
	v, _ := t.Get(k).(bool)

	return v
}

func (t Param) GetNum(k string) json.Number {
	v, _ := t.Get(k).(json.Number)

	return v
}

func (t Param) GetInt(k string) int64 {
	res, _ := t.GetNum(k).Int64()

	return res
}

func (t Param) GetFloat(k string) float64 {
	res, _ := t.GetNum(k).Float64()

	return res
}

type OrderedMap struct {
	filed []string
	data  map[string]interface{}
}

func NewOrderedMap() *OrderedMap {
	res := &OrderedMap{
		data: map[string]interface{}{},
	}

	return res
}

func SomeOrderedMap(n int) []*OrderedMap {
	res := []*OrderedMap{}

	for i := 0; i < n; i++ {
		res = append(res, NewOrderedMap())
	}

	return res
}

func (t *OrderedMap) Set(k string, v interface{}) {
	_, ok := t.data[k]
	if !ok {
		t.filed = append(t.filed, k)
	}

	t.data[k] = v
}

func (t *OrderedMap) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("{")

	first := true

	for _, key := range t.filed {
		if value, ok := t.data[key]; ok {
			if !first {
				b.WriteString(",")
			}
			first = false

			encodedKey, _ := json.Marshal(key)

			encodedValue, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}

			b.Write(encodedKey)
			b.WriteString(":")
			b.Write(encodedValue)
		}
	}

	b.WriteString("}")

	return b.Bytes(), nil
}

type StrListMap map[string][]string

func (t StrListMap) Find(raw ...string) []string {
	res := []string{}

	for _, v := range raw {
		res = append(res, t[v]...)
	}

	res = SetStrList(res)

	return res
}

func (t StrListMap) Set() {
	for k := range t {
		t[k] = SetStrList(t[k])
	}
}

func (t StrListMap) Output() []string {
	res := []string{}

	for k, v := range t {
		line := fmt.Sprintf("%v: %v", k, v)
		res = append(res, line)
	}

	return res
}

type Int32ListMap map[string][]int32

func (t Int32ListMap) Find(raw ...string) []int32 {
	res := []int32{}

	for _, v := range raw {
		res = append(res, t[v]...)
	}

	res = SetInt32List(res)

	return res
}

func (t Int32ListMap) Output() []string {
	res := []string{}

	for k, v := range t {
		line := fmt.Sprintf("%v: %v", k, v)
		res = append(res, line)
	}

	return res
}

type Int64ListMap map[string][]int64

func (t Int64ListMap) Find(raw ...string) []int64 {
	res := []int64{}

	for _, v := range raw {
		res = append(res, t[v]...)
	}

	res = SetInt64List(res)

	return res
}

func (t Int64ListMap) Output() []string {
	res := []string{}

	for k, v := range t {
		line := fmt.Sprintf("%v: %v", k, v)
		res = append(res, line)
	}

	return res
}

type MapStringString map[string]string

func (t MapStringString) FromParam(raw map[string]interface{}) {
	SetMapStringStringFromMapStringInterface(raw, t)
}

func (t MapStringString) Find(raw ...string) []string {
	res := []string{}

	for _, v := range raw {
		res = append(res, t[v])
	}

	res = SetStrList(res)

	return res
}

func (t MapStringString) Merge(raw ...MapStringString) MapStringString {
	for _, v := range raw {
		for k1, v1 := range v {
			t[k1] = v1
		}
	}

	return t
}

type MapStringInt32 map[string]int32

func (t MapStringInt32) Find(raw ...string) []int32 {
	res := []int32{}

	for _, v := range raw {
		res = append(res, t[v])
	}

	res = SetInt32List(res)

	return res
}

func (t MapStringInt32) Merge(raw ...MapStringInt32) MapStringInt32 {
	for _, v := range raw {
		for k1, v1 := range v {
			t[k1] = v1
		}
	}

	return t
}

type MapStringInt64 map[string]int64

func (t MapStringInt64) Find(raw ...string) []int64 {
	res := []int64{}

	for _, v := range raw {
		res = append(res, t[v])
	}

	res = SetInt64List(res)

	return res
}

func (t MapStringInt64) Merge(raw ...MapStringInt64) MapStringInt64 {
	for _, v := range raw {
		for k1, v1 := range v {
			t[k1] = v1
		}
	}

	return t
}

type MapStringStrList map[string][]string

type MapInt32String map[int32]string

func (t MapInt32String) Find(raw ...int32) []string {
	res := []string{}

	for _, v := range raw {
		res = append(res, t[v])
	}

	res = SetStrList(res)

	return res
}

func (t MapInt32String) Merge(raw ...MapInt32String) MapInt32String {
	for _, v := range raw {
		for k1, v1 := range v {
			t[k1] = v1
		}
	}

	return t
}

type MapInt64String map[int64]string

func (t MapInt64String) Find(raw ...int64) []string {
	res := []string{}

	for _, v := range raw {
		res = append(res, t[v])
	}

	res = SetStrList(res)

	return res
}

func (t MapInt64String) Merge(raw ...MapInt64String) MapInt64String {
	for _, v := range raw {
		for k1, v1 := range v {
			t[k1] = v1
		}
	}

	return t
}

type MapInt64StrList map[int64][]string
