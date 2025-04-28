package types

import (
	"time"
)

func GetZeroTime(t time.Time) time.Time {
	res := time.Date(
		t.Year(),
		t.Month(),
		t.Day(),
		0,
		0,
		0,
		0,
		CstTimeZone,
	)

	return res
}

func GetNextMinuteZero(t time.Time) time.Time {
	return t.Truncate(time.Minute).Add(time.Minute)
}

// 蔡勒（Zeller）公式。即w=y+[y/4]+[c/4]-2c+[26(m+1)/10]+d-1
func ZellerFunction2Week(t time.Time) int {
	var y, m, c int
	if t.Month() >= 3 {
		m = int(t.Month())
		y = t.Year() % 100
		c = t.Year() / 100
	} else {
		m = int(t.Month()) + 12
		y = (t.Year() - 1) % 100
		c = (t.Year() - 1) / 100
	}

	week := y + (y / 4) + (c / 4) - 2*c + ((26 * (m + 1)) / 10) + t.Day() - 1
	if week < 0 {
		week = 7 - (-week)%7
	} else {
		week = week % 7
	}

	if week == 0 {
		week = 7
	}

	return week
}

//获取当天0点的时间戳
func GetZeroTimeStamp(t time.Time) int64 {
	return GetZeroTime(t).Unix()
}

func GetZeroTimeStampMs(t time.Time) int64 {
	return GetZeroTime(t).UnixMilli()
}

func GetFirstDateOfWeek(raw time.Time) time.Time {
	offset := int(time.Monday - raw.Weekday())
	if offset > 0 {
		offset = -6
	}

	weekStartDay := GetZeroTime(raw).AddDate(0, 0, offset)

	return weekStartDay
}

func GetDayRange(t time.Time) (s, e time.Time) {
	s = GetZeroTime(t)
	e = s.AddDate(0, 0, 1)

	return
}

// 获取近一周的开始时间
func GetRecentWeekTimeStamp(t time.Time) int64 {
	oneWeekAgo := t.AddDate(0, 0, -7) // 往前推7天
	return GetZeroTimeStampMs(oneWeekAgo)
}

func GetWeekStartTime(t time.Time) time.Time {
	t = GetZeroTime(t)
	week := ZellerFunction2Week(t)
	n := SecondOf1D * (week - 1)

	delta := time.Duration(n) * time.Second

	return t.Add(delta * -1)
}

//获取给定时间所在周的起始时间和结束时间
func GetWeekRange(t time.Time) (startOfWeek, endOfWeek time.Time) {
	// 周一为一周的第一天，但在 Go 中，星期天（Sunday）为第一天
	// 计算今天是本周的第几天
	weekday := int(t.Weekday()) // 0: Sunday, 1: Monday, ..., 6: Saturday
	// 计算距离上周日的天数
	daysAgo := -1 * weekday
	// 计算上周日
	startOfWeek = t.AddDate(0, 0, daysAgo)
	// 上周日的 00:00
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, CstTimeZone)
	// 本周末的 23:59:59
	endOfWeek = startOfWeek.AddDate(0, 0, 7).Add(-time.Second)

	return startOfWeek, endOfWeek
}

func GetWeekRangeByZeller(t time.Time) (s, e time.Time) {
	s = GetWeekStartTime(t)
	e = s.AddDate(0, 0, 7)

	return
}

// 获取近一月的开始时间
func GetRecentMonthTimeStamp(t time.Time) int64 {
	oneMonthAgo := t.AddDate(0, -1, 0) // 往前推1个月
	return GetZeroTimeStampMs(oneMonthAgo)
}

func GetMonthStartTime(t time.Time) time.Time {
	res := time.Date(
		t.Year(),
		t.Month(),
		1,
		0,
		0,
		0,
		0,
		CstTimeZone,
	)

	return res
}

func GetMonthRange(t time.Time) (s, e time.Time) {
	s = GetMonthStartTime(t)
	e = s.AddDate(0, 1, 0)

	return
}

func GetDaysOfMonth(t time.Time) int {
	year, month, _ := t.Date()

	var nextMonth time.Time
	if month == time.December {
		nextMonth = time.Date(year+1, time.January, 1, 0, 0, 0, 0, t.Location())
	} else {
		nextMonth = time.Date(year, month+1, 1, 0, 0, 0, 0, t.Location())
	}

	lastDay := nextMonth.AddDate(0, 0, -1)

	return lastDay.Day()
}

// 获取近30天的开始时间
func GetRecent30DaysTimeStamp(t time.Time) int64 {
	thirtyDaysAgo := t.AddDate(0, 0, -30) // 往前推30天
	return GetZeroTimeStampMs(thirtyDaysAgo)
}

// 获取近半年的开始时间戳
func GetRecentHalfYearTimeStamp(t time.Time) int64 {
	halfYearAgo := t.AddDate(0, -6, 0) // 往前推6个月
	return GetZeroTimeStampMs(halfYearAgo)
}

type StartEnd struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

//获取当月每周的开始时间和结束时间
func GetWeekStartEndInMonth() (res []StartEnd) {
	// 获取当前时间
	now := time.Now()
	// 获取当前月份
	year, month, _ := now.Date()
	// 构建当前月份的第一天
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, CstTimeZone)
	// 获取下个月的第一天，用于确定本月的结束时间
	nextMonth := firstOfMonth.AddDate(0, 1, 0)
	// 初始化当前月份的开始时间为第一天的开始时间
	startTime := firstOfMonth
	// 循环获取每周的开始时间和结束时间，直到超过本月的结束时间
	for startTime.Before(nextMonth) {
		// 计算本周的结束时间，确保不超过本月的最后一天
		endTime := startTime.AddDate(0, 0, 7)
		if endTime.After(nextMonth) {
			endTime = nextMonth
		}
		wse := StartEnd{
			Start: startTime.UnixMilli(),
			End:   endTime.UnixMilli(),
		}
		res = append(res, wse)
		// 将开始时间移动到下一周的开始时间
		startTime = endTime
	}
	return
}

//获取当月每天的开始时间和结束时间
func GetDayStartEndInMonth() (res []StartEnd) {
	now := time.Now()
	// 获取当前月份
	year, month, _ := now.Date()
	// 构建当前月份的第一天和下个月的第一天
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, CstTimeZone)
	nextMonth := firstOfMonth.AddDate(0, 1, 0)
	// 初始化当前日期为本月第一天
	currentDate := firstOfMonth
	// 循环获取当月每天的开始时间和结束时间，直到超过本月的结束时间
	for currentDate.Before(nextMonth) {
		// 当天的开始时间（零点整）
		startOfDay := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, CstTimeZone)
		// 第二天的开始时间，即当天的结束时间
		endOfDay := startOfDay.AddDate(0, 0, 1)
		wse := StartEnd{
			Start: startOfDay.UnixMilli(),
			End:   endOfDay.UnixMilli(),
		}
		res = append(res, wse)
		// 将当前日期移动到下一天
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return
}

//获取当年每月的开始时间和结束时间
func GetMonthStartEndInYear() (res []StartEnd) {
	// 获取当前时间
	now := time.Now()
	// 获取当前年份
	year := now.Year()
	// 循环获取当年每个月的开始时间和结束时间
	for month := time.January; month <= time.December; month++ {
		// 构建当前月份的第一天和下个月的第一天
		firstOfMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, CstTimeZone)
		if month == time.December {
			firstOfMonth = time.Date(year+1, time.January, 1, 0, 0, 0, 0, CstTimeZone)
		}
		// 当月的开始时间为第一天的零点整
		startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, CstTimeZone)
		wse := StartEnd{
			Start: startOfMonth.UnixMilli(),
			End:   firstOfMonth.UnixMilli(),
		}
		res = append(res, wse)
	}

	return
}

//获取一年的每月0点的时间戳
func GetNowYearTimeStampZero(t time.Time) []int64 {
	res := []int64{}
	for i := 0; i < 12; i++ {
		addTime := time.Date(t.Year(), time.Month(i+1), 1, 0, 0, 0, 0, CstTimeZone)
		timeStamp := addTime.Unix()
		res = append(res, timeStamp)
	}

	return res
}

//获取一年的第1天的0点的时间戳
func GetFirstDayOfYearTimeStamp(t time.Time) int64 {
	return GetYearStartTime(t).Unix()
}

func GetYearStartTime(t time.Time) time.Time {
	res := time.Date(
		t.Year(),
		1,
		1,
		0,
		0,
		0,
		0,
		CstTimeZone,
	)

	return res
}

func GetDayZeroRange(t0, t1 time.Time) (from, to time.Time) {
	from = GetZeroTime(t0)

	to = t1.Add(Duration1D)
	to = GetZeroTime(to)

	return
}

func GetMonthZeroRange(t0, t1 time.Time) (from, to time.Time) {
	from = GetMonthStartTime(t0)

	to = GetMonthStartTime(t1)

	return
}

func GetDayZeroRangeFromTick(t0, t1 int64) (from, to time.Time) {
	from = time.Unix(t0, 0)
	from = GetZeroTime(from)

	to = time.Unix(t1, 0).Add(Duration1D)
	to = GetZeroTime(to)

	return
}

func GetMonthZeroRangeFromTick(t0, t1 int64) (from, to time.Time) {
	from = time.Unix(t0, 0)
	from = GetMonthStartTime(from)

	to = time.Unix(t1, 0)
	to = GetMonthStartTime(to)

	return
}
