package types

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

// Equal

func IsEqualStrList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func IsEqualBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// List交集

func IntersectIntList(a, b []int) []int {
	res := []int{}
	if len(a) == 0 || len(b) == 0 {
		return res
	}

	m := ToIntCounter(a)
	for _, e := range b {
		_, ok := m[e]
		if ok {
			m[e]++
		}

		if m[e] == 1 {
			res = append(res, e)
		}
	}

	return res
}

func IntersectInt32List(a, b []int32) []int32 {
	res := []int32{}
	if len(a) == 0 || len(b) == 0 {
		return res
	}

	m := ToInt32Counter(a)
	for _, e := range b {
		_, ok := m[e]
		if ok {
			m[e]++
		}

		if m[e] == 1 {
			res = append(res, e)
		}
	}

	return res
}

func IntersectInt64List(a, b []int64) []int64 {
	res := []int64{}
	if len(a) == 0 || len(b) == 0 {
		return res
	}

	m := ToInt64Counter(a)
	for _, e := range b {
		_, ok := m[e]
		if ok {
			m[e]++
		}

		if m[e] == 1 {
			res = append(res, e)
		}
	}

	return res
}

func IntersectStrList(a, b []string) []string {
	res := []string{}
	if len(a) == 0 || len(b) == 0 {
		return res
	}

	m := ToStrListCounter(a)

	for _, e := range b {
		_, ok := m[e]
		if ok {
			m[e]++
		}

		if m[e] == 1 {
			res = append(res, e)
		}
	}

	return res
}

func SliceIntersection(a, b []string) []string {
	res := []string{}
	if len(a) == 0 || len(b) == 0 {
		return res
	}

	m := ToStrListFlag(b)
	for _, v := range a {
		if m[v] {
			res = append(res, v)
		}
	}

	return res
}

// List去重

func SetIntList(raw []int) []int {
	res := []int{}
	if len(raw) == 0 {
		return res
	}

	m := ToIntFlag(raw)
	for k := range m {
		res = append(res, k)
	}

	return res
}

func SetInt32List(raw []int32) []int32 {
	res := []int32{}
	if len(raw) == 0 {
		return res
	}

	m := ToInt32Flag(raw)
	for k := range m {
		res = append(res, k)
	}

	return res
}

func SetInt64List(raw []int64) []int64 {
	res := []int64{}
	if len(raw) == 0 {
		return res
	}

	m := ToInt64Flag(raw)
	for k := range m {
		res = append(res, k)
	}

	return res
}

func OpSetInt64List(raw []int64) []int64 {
	res := []int64{}
	if len(raw) == 0 {
		return res
	}

	m := map[int64]bool{}
	rawOrder := map[int64]int{}

	for i, v := range raw {
		if m[v] {
			continue
		}

		m[v] = true
		rawOrder[v] = i
	}

	for k := range m {
		res = append(res, k)
	}

	sort.Slice(res, func(i, j int) bool {
		return rawOrder[res[i]] < rawOrder[res[j]]
	})

	return res
}

func SetStrList(raw []string) []string {
	res := []string{}
	if len(raw) == 0 {
		return res
	}

	m := ToStrListFlag(raw)
	for k := range m {
		res = append(res, k)
	}

	return res
}

func OpSetStrList(raw []string) []string {
	res := []string{}
	if len(raw) == 0 {
		return res
	}

	m := map[string]bool{}
	rawOrder := map[string]int{}

	for i, v := range raw {
		if m[v] {
			continue
		}

		m[v] = true
		rawOrder[v] = i
	}

	for k := range m {
		res = append(res, k)
	}

	sort.Slice(res, func(i, j int) bool {
		return rawOrder[res[i]] < rawOrder[res[j]]
	})

	return res
}

func SetDurationList(raw []time.Duration) []time.Duration {
	res := []time.Duration{}
	if len(raw) == 0 {
		return res
	}

	m := map[time.Duration]bool{}

	for _, v := range raw {
		m[v] = true
	}

	for k := range m {
		res = append(res, k)
	}

	return res
}

func SortedSetInt32List(raw []int32) []int32 {
	res := SetInt32List(raw)

	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})

	return res
}

func SortedSetInt64List(raw []int64) []int64 {
	res := SetInt64List(raw)

	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})

	return res
}

// List1 - List2

func IntListSub(a, b []int) []int {
	if len(b) == 0 {
		return a
	}

	res := []int{}
	if len(a) == 0 {
		return res
	}

	m := ToIntFlag(b)
	for _, v := range a {
		_, ok := m[v]
		if !ok {
			res = append(res, v)
		}
	}

	return res
}

func Int32ListSub(a, b []int32) []int32 {
	if len(b) == 0 {
		return a
	}

	res := []int32{}
	if len(a) == 0 {
		return res
	}

	m := ToInt32Flag(b)
	for _, v := range a {
		_, ok := m[v]
		if !ok {
			res = append(res, v)
		}
	}

	return res
}

func Int64ListSub(a, b []int64) []int64 {
	if len(b) == 0 {
		return a
	}

	res := []int64{}
	if len(a) == 0 {
		return res
	}

	m := ToInt64Flag(b)
	for _, v := range a {
		_, ok := m[v]
		if !ok {
			res = append(res, v)
		}
	}

	return res
}

func StrListSub(a, b []string) []string {
	if len(b) == 0 {
		return a
	}

	res := []string{}
	if len(a) == 0 {
		return res
	}

	m := ToStrListFlag(b)
	for _, v := range a {
		_, ok := m[v]
		if !ok {
			res = append(res, v)
		}
	}

	return res
}

//数组差集：返回数组a中不存在于数组b的元素

func StrListDiff(a, b []string) []string {
	if len(b) == 0 {
		return a
	}

	m := ToStrListFlag(b)

	result := make([]string, 0, len(a))
	for _, va := range a {
		if m[va] == true {
			continue
		}
		result = append(result, va)
	}

	return result
}

//List Match

func MatchIntList(a, b []int) int {
	res := 0

	m := ToIntFlag(b)
	for _, v := range a {
		if m[v] {
			res++
		}
	}

	return res
}

func MatchStrList(a, b []string) int {
	res := 0

	m := ToStrListFlag(a)
	for _, v := range b {
		if m[v] {
			res++
		}
	}

	return res
}

//In List

func InIntList(raw []int, x int) bool {
	for _, v := range raw {
		if v == x {
			return true
		}
	}

	return false
}

func InInt32List(raw []int32, x int32) bool {
	for _, v := range raw {
		if v == x {
			return true
		}
	}

	return false
}

func InInt64List(raw []int64, x int64) bool {
	for _, v := range raw {
		if v == x {
			return true
		}
	}

	return false
}

func InStrList(raw []string, x string) bool {
	for _, v := range raw {
		if v == x {
			return true
		}
	}

	return false
}

//union

func SliceUnion(a, b []string) []string {
	resList := make([]string, 0, len(a)+len(b))
	tmpMap := make(map[string]bool, len(b))

	for _, v := range b {
		if _, ok := tmpMap[v]; !ok {
			tmpMap[v] = true
			resList = append(resList, v)
		}
	}

	for _, v := range a {
		if _, ok := tmpMap[v]; !ok {
			tmpMap[v] = true
			resList = append(resList, v)
		}
	}

	return resList
}

//join

func JoinIntList(m string, raw ...int) string {
	if len(raw) == 0 {
		return ""
	}

	var b strings.Builder
	l := len(raw) - 1
	for i := 0; i < l; i++ {
		b.WriteString(strconv.Itoa(raw[i]))
		b.WriteString(m)
	}
	b.WriteString(strconv.Itoa(raw[l]))

	res := b.String()

	return res
}

func JoinAndSortInt(m string, raw ...int) string {
	sort.Slice(raw, func(i, j int) bool {
		return raw[i] < raw[j]
	})

	return JoinIntList(m, raw...)
}

func JoinInt32List(m string, raw ...int32) string {
	if len(raw) == 0 {
		return ""
	}

	var b strings.Builder
	l := len(raw) - 1
	for i := 0; i < l; i++ {
		b.WriteString(ParseInt64ToStr(int64(raw[i])))
		b.WriteString(m)
	}
	b.WriteString(ParseInt64ToStr(int64(raw[l])))

	res := b.String()

	return res
}

func JoinAndSortInt32(m string, raw ...int32) string {
	sort.Slice(raw, func(i, j int) bool {
		return raw[i] < raw[j]
	})

	return JoinInt32List(m, raw...)
}

func JoinInt64List(m string, raw ...int64) string {
	if len(raw) == 0 {
		return ""
	}

	var b strings.Builder
	l := len(raw) - 1
	for i := 0; i < l; i++ {
		b.WriteString(ParseInt64ToStr(raw[i]))
		b.WriteString(m)
	}
	b.WriteString(ParseInt64ToStr(raw[l]))

	res := b.String()

	return res
}

func JoinFlot64List(m string, raw ...float64) string {
	if len(raw) == 0 {
		return ""
	}

	var b strings.Builder
	l := len(raw) - 1
	for i := 0; i < l; i++ {
		b.WriteString(ParseFloat64ToStr(raw[i]))
		b.WriteString(m)
	}
	b.WriteString(ParseFloat64ToStr(raw[l]))

	res := b.String()

	return res
}

func JoinAndSortInt64(m string, raw ...int64) string {
	sort.Slice(raw, func(i, j int) bool {
		return raw[i] < raw[j]
	})

	return JoinInt64List(m, raw...)
}

func JoinStrList(m string, raw ...string) string {
	if len(raw) == 0 {
		return ""
	}

	var b strings.Builder
	l := len(raw) - 1
	for i := 0; i < l; i++ {
		b.WriteString(raw[i])
		b.WriteString(m)
	}
	b.WriteString(raw[l])

	res := b.String()

	return res
}

func JoinKeys(raw ...string) string {
	return JoinStrList(KeyDelimiter, raw...)
}

func JoinTags(raw ...string) string {
	return JoinStrList(TagDelimiter, raw...)
}

func JoinI64Tags(raw ...int64) string {
	return JoinInt64List(NumDelimiter, raw...)
}

func JoinF64Tags(raw ...float64) string {
	return JoinFlot64List(NumDelimiter, raw...)
}

//split

func SplitToIntSlice(raw string, d ...string) []int {
	if len(raw) == 0 {
		return []int{}
	}

	m := ParseStrParam(d, NumDelimiter)
	tmp := strings.Split(raw, m)

	res := []int{}
	for _, v := range tmp {
		id, e := strconv.Atoi(v)
		if e != nil {
			continue
		}

		res = append(res, id)
	}

	return res
}

func SplitToInt32Slice(raw string, d ...string) []int32 {
	if len(raw) == 0 {
		return []int32{}
	}

	m := ParseStrParam(d, NumDelimiter)
	tmp := strings.Split(raw, m)

	res := []int32{}
	for _, v := range tmp {
		id, e := strconv.Atoi(v)
		if e != nil {
			continue
		}

		res = append(res, int32(id))
	}

	return res
}

func SplitToInt64Slice(raw string, d ...string) []int64 {
	if len(raw) == 0 {
		return []int64{}
	}

	m := ParseStrParam(d, NumDelimiter)
	tmp := strings.Split(raw, m)

	res := []int64{}
	for _, v := range tmp {
		id, e := ParseInt64FromStr(v)
		if e != nil {
			continue
		}

		res = append(res, id)
	}

	return res
}

func SplitToFloat64Slice(raw string, d ...string) []float64 {
	if len(raw) == 0 {
		return []float64{}
	}

	m := ParseStrParam(d, NumDelimiter)
	tmp := strings.Split(raw, m)

	res := []float64{}
	for _, v := range tmp {
		id, e := ParseFloat64FromStr(v)
		if e != nil {
			continue
		}

		res = append(res, id)
	}

	return res
}

func SplitToStrSlice(raw string, d ...string) []string {
	if len(raw) == 0 {
		return []string{}
	}

	m := ParseStrParam(d, NumDelimiter)
	tmp := strings.Split(raw, m)

	res := []string{}
	for _, v := range tmp {
		res = append(res, v)
	}

	return res
}

func SplitKeys(raw string) []string {
	return SplitToStrSlice(raw, KeyDelimiter)
}

func SplitTags(raw string) []string {
	return SplitToStrSlice(raw, TagDelimiter)
}

func SplitI64Tags(raw string) []int64 {
	return SplitToInt64Slice(raw, NumDelimiter)
}

func SplitF64Tags(raw string) []float64 {
	return SplitToFloat64Slice(raw, NumDelimiter)
}

//find index

func FindStrIndex(m string, raw ...string) int {
	for i, v := range raw {
		if m == v {
			return i
		}
	}

	return -1
}

//filter

func FilterNaNStr(raw []string) []string {
	res := []string{}

	for _, v := range raw {
		if v == "" {
			continue
		}

		res = append(res, v)
	}

	return res
}

//reverse

func ReverseInt32List(s []int32) {
	l := len(s)
	for i := 0; i < l/2; i++ {
		s[i], s[l-1-i] = s[l-1-i], s[i]
	}
}

func ReverseInt64List(s []int64) {
	l := len(s)
	for i := 0; i < l/2; i++ {
		s[i], s[l-1-i] = s[l-1-i], s[i]
	}
}

func ReverseStr64List(s []string) {
	l := len(s)
	for i := 0; i < l/2; i++ {
		s[i], s[l-1-i] = s[l-1-i], s[i]
	}
}

// remove

func SliceRemoveInt64(slice []int64, i int) []int64 {
	if i < 0 || i >= len(slice) {
		return slice
	}

	return append(slice[:i], slice[i+1:]...)
}

func SortInt64(x []int64) { sort.Sort(Int64Slice(x)) }

type Int64Slice []int64

func (x Int64Slice) Len() int           { return len(x) }
func (x Int64Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Int64Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func MergeI64k(raw ...I64K) []int64 {
	res := []int64{}

	for _, v := range raw {
		res = append(res, v.K()...)
	}

	res = SetInt64List(res)

	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})

	return res
}
