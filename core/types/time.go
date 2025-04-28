package types

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	RFC3339Tail                   = ".000Z"
	ISO8601                       = "2006-01-02T15:04:05.000Z0700"
	ISO8601Z8Tail                 = "T00:00:00.000Z"
	TimeLayout                    = "2006-01-02T15:04:05Z"
	SimpleTimeLayoutYyyyMmDdMmSs  = "2006-01-02 15:04:05"
	SimpleTimeLayout              = "2006-01-02"
	CompactTimeLayoutYyyyMmDdMmSs = "200601021504"
	CompactTimeLayoutYyyyMmDdMm   = "2006010215"
	CompactTimeLayout             = "20060102"
)

const (
	Duration1D = time.Hour * 24
	Duration1W = Duration1D * 7
)

const (
	SecondOf1M = 60
	SecondOf1H = SecondOf1M * 60
	SecondOf1D = SecondOf1H * 24
	SecondOf1W = SecondOf1D * 7
)

const (
	MsOf1S = 1000
	MsOf1M = MsOf1S * 60
	MsOf1H = MsOf1M * 60
	MsOf1D = MsOf1H * 24
	MsOf1W = MsOf1D * 7
)

func FormatTimeWithLayout(t time.Time, layout string) string {
	return time.Unix(t.Unix(), 0).Format(layout)
}

func ISO8601TimeStr(t time.Time) string {
	return FormatTimeWithLayout(t, ISO8601)
}

func SimpleTimeStr(t time.Time) string {
	return FormatTimeWithLayout(t, SimpleTimeLayout)
}

func ParseTimeStrWithLayout(layout, raw string) (t time.Time, err error) {
	t, err = time.ParseInLocation(layout, raw, CstTimeZone)

	return
}

func ParseTimeStr(raw string) (t time.Time, err error) {
	return ParseTimeStrWithLayout(TimeLayout, raw)
}

func ParseCompactTimeStr(raw string) (t time.Time, err error) {
	return time.ParseInLocation(CompactTimeLayout, raw, CstTimeZone)
}

func MustParseTimeStr(raw string) time.Time {
	t, _ := ParseTimeStr(raw)

	return t
}

func ParseInfluxTime(raw string) time.Time {
	ts, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}
	}

	return ts.In(CstTimeZone)
}

func ToUtcTime(raw time.Time) time.Time {
	return raw.In(CstTimeZone).UTC()
}

func StrToTime(raw string) time.Time {
	timeLayout := "2006-1-2 15:04:05" //模板
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(timeLayout, raw, loc)
	return theTime
}

func FormatToUtc(raw time.Time) string {
	return ToUtcTime(raw).Format(time.RFC3339)
}

func IsSameMonthByTick(t1, t2 int64) bool {
	dateTime1 := time.Unix(t1, 0)
	dateTime2 := time.Unix(t2, 0)

	return IsSameMonthByTime(dateTime1, dateTime2)
}

func IsSameMonthByTime(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.Month() == t2.Month()
}

func IsSameDay(t1, t2 int64) bool {
	tt1 := time.Unix(t1, 0)
	tt2 := time.Unix(t2, 0)

	return tt1.Year() == tt2.Year() && tt1.Month() == tt2.Month() && tt1.Day() == tt2.Day()
}

func ParseRFC3339TimeStr(raw string) (int64, error) {
	if !strings.HasSuffix(raw, RFC3339Tail) {
		raw += RFC3339Tail
	}

	uTime, err := time.ParseInLocation(time.RFC3339, raw, CstTimeZone)
	if err != nil {
		return 0, err
	}

	return uTime.Unix(), nil
}

func RandomMinute(min, max int) time.Duration {
	return RandomSecond(min*60, max*60)
}

func RandomSecond(min, max int) time.Duration {
	return RandomMillisecond(min*1000, max*1000)
}

func RandomMillisecond(min, max int) time.Duration {
	return RandomMicrosecond(min*1000, max*1000)
}

func RandomMicrosecond(min, max int) time.Duration {
	n := RandomInt(min, max)
	return Microsecond(n)
}

func Microsecond(t int) time.Duration {
	return time.Microsecond * time.Duration(t)
}

func IntervalByHour(n int) TaskInterval {
	return IntervalByMinute(n * 60)
}

func IntervalByMinute(n int) TaskInterval {
	return IntervalBySecond(n * 60)
}

func IntervalBySecond(n int) TaskInterval {
	return IntervalByMillisecond(n * 1000)
}

func IntervalByMillisecond(n int) TaskInterval {
	return IntervalByMicrosecond(n * 1000)
}

func IntervalByMicrosecond(n int) TaskInterval {
	if n <= 0 {
		panic("argument <= 0")
	}

	return NewInterval(time.Microsecond * time.Duration(n))
}

func NewInterval(d time.Duration) TaskInterval {
	var res TaskInterval = func() time.Duration {
		return d
	}

	return res
}

func SumDuration(raw ...time.Duration) time.Duration {
	var res time.Duration
	for _, v := range raw {
		res += v
	}

	return res
}

func MinDuration(raw ...time.Duration) time.Duration {
	if len(raw) == 0 {
		return 0
	}

	min := raw[0]
	for _, i := range raw {
		if min > i {
			min = i
		}
	}

	return min
}

func MaxDuration(raw ...time.Duration) time.Duration {
	if len(raw) == 0 {
		return 0
	}

	max := raw[0]
	for _, i := range raw {
		if max < i {
			max = i
		}
	}

	return max
}

func ComputedWait(ctx context.Context, f IntervalComputer, t ...time.Time) {
	interval := f(ctx, ParseTime(t))
	time.Sleep(interval)
}

func Count(year int, month int) (days int) {
	if month != 2 {
		if month == 4 || month == 6 || month == 9 || month == 11 {
			return 30
		}

		return 31
	}

	if (year%4 == 0 && year%100 != 0) || (year%400) == 0 {
		return 29
	}

	return 28
}

func GetMonthChar(t int64) string {
	month := time.Unix(t, 0).Month()
	return strconv.Itoa(int(month)) + "月"
}

var MonthName = map[int]string{
	1:  "一月",
	2:  "二月",
	3:  "三月",
	4:  "四月",
	5:  "五月",
	6:  "六月",
	7:  "七月",
	8:  "八月",
	9:  "九月",
	10: "十月",
	11: "十一月",
	12: "十二月",
}

func GetMonthName(raw int) string {
	return MonthName[raw]
}

func MonthsBetween(start, end time.Time) int {
	if start.After(end) {
		return -MonthsBetween(end, start)
	}

	startYear, startMonth, _ := start.Date()
	endYear, endMonth, _ := end.Date()

	months := (endYear-startYear)*12 + int(endMonth-startMonth) + 1

	return months
}

func MsTimeout(raw uint64) time.Duration {
	return time.Millisecond * time.Duration(raw)
}

func SecondTimeout(raw uint64) time.Duration {
	return time.Second * time.Duration(raw)
}

func MinuteTimeout(raw uint64) time.Duration {
	return time.Minute * time.Duration(raw)
}

func DefaultFromToMs(from, to int64) (f, t int64) {
	now := time.Now()
	if from == 0 {
		from = GetZeroTimeStampMs(now)
	}

	if to == 0 {
		to = now.UnixMilli()
	}

	return from, to
}

var (
	timerStart time.Time
	timerCount int64
)

func StartTimer() {
	timerStart = time.Now()
	setTimerCount(0)
}

func setTimerCount(raw int64) {
	atomic.StoreInt64(&timerCount, raw)
}

func getTimerCount() int64 {
	return atomic.LoadInt64(&timerCount)
}

func CostSince() string {
	atomic.AddInt64(&timerCount, 1)
	return fmt.Sprintf("step %v, cost %v", getTimerCount(), time.Now().Sub(timerStart))
}

func ShowCost() {
	fmt.Println(CostSince())
}

func TimeStrToTimestamp(timeStr string) (int64, error) {
	// 解析时间字符串
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return 0, err
	}
	// 获取当前日期
	now := time.Now()
	// 将解析的时间与当前日期组合
	combinedTime := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
	// 转为时间戳(秒)
	timestamp := combinedTime.Unix()

	return timestamp, nil
}

func FillMissingDays(tick []int64) []int64 {
	sort.Slice(tick, func(i, j int) bool {
		return tick[i] < tick[j]
	})

	var result []int64
	if len(tick) == 0 {
		return result
	}

	result = append(result, tick[0])

	for i := 1; i < len(tick); i++ {
		current := tick[i]
		previous := result[len(result)-1]

		diff := (current - previous) / SecondOf1D

		if diff > 1 {
			for d := 1; d < int(diff); d++ {
				result = append(result, previous+int64(d)*(24*60*60))
			}
		}

		result = append(result, current)
	}

	return result
}

func FillMissingMonths(tick []int64) []int64 {
	sort.Slice(tick, func(i, j int) bool {
		return tick[i] < tick[j]
	})

	var result []int64
	l := len(tick)
	if l == 0 {
		return result
	}

	if l == 1 {
		return []int64{tick[0]}
	}

	curr := time.Unix(tick[0], 0)
	total := MonthsBetween(curr, time.Unix(tick[l-1], 0))
	for i := 0; i < total; i++ {
		curr = curr.AddDate(0, i, 0)
		result = append(result, GetMonthStartTime(curr).Unix())
	}

	return result
}
