package dsp

import (
	"fmt"
	"math"
	. "mykit/core/types"
	"sort"
	"sync"
	"time"
)

func NowBlankValuePair() ValuePairList {
	return SomeBlankTickBefore(time.Now().Unix(), 6)
}

func SomeBlankTickBefore(tick int64, n ...int) ValuePairList {
	return SomeBlackTick(tick, -1, n...)
}

func SomeBlackTick(tick, delta int64, n ...int) ValuePairList {
	res := ValuePairList{}
	times := ParseIntParam(n, 1)
	for i := 0; i < times; i++ {
		res = append(res, BlackValuePair(tick+int64(i+1)*delta))
	}

	return res
}

func BlackValuePair(t interface{}) ValuePair {
	return ValuePair{t, ""}
}

func TakeDigitsTimeValue(tick int64, data float64, n ...int) ValuePair {
	if len(n) == 0 {
		return ValuePair{tick, data}
	}

	return ValuePair{tick, TakeDigits(data, n[0])}
}

type DataOD struct {
	Start  int64
	End    int64
	Value  float64
	Status int32
}

type DataODList []DataOD

func (t DataODList) Len() int {
	return len(t)
}

func (t DataODList) Less(i, j int) bool {
	return t[i].Start < t[j].Start
}

func (t DataODList) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t DataODList) StartTickMap() TickDataODMap {
	res := TickDataODMap{}

	for _, v := range t {
		res[v.Start] = v
	}

	return res
}

func (t DataODList) SeekTick(tick int64) (res int) {
	left := 0
	right := len(t) - 1
	if right < 0 || t[0].Start > tick {
		return -1
	}

	for left <= right {
		mid := left + (right-left)/2
		if t[mid].Start == tick {
			return mid
		} else if t[mid].Start < tick {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return right
}

func (t DataODList) From(start int64) DataODList {
	index := t.SeekTick(start)
	if index > -1 {
		return t[index:]
	}

	return t
}

func (t DataODList) Padding(from, to int64) DataODList {
	res := DataODList{}

	if len(t) == 0 {
		return res
	}

	// 添加开头
	if t[0].Start > from {
		res = append(res, DataOD{from, t[0].Start - 1, 0, 0})
	}

	last := len(t) - 1
	if last == 0 {
		res = append(res, t[0])
	}

	// 添加相邻区间
	for i := 0; i < last; i++ {
		res = append(res, t[i])

		if t[i+1].Start-t[i].End > 2 {
			res = append(res, DataOD{t[i].End + 1, t[i+1].Start - 1, 0, 0})
		}
	}

	if last > 0 {
		res = append(res, t[last])
	}

	// 添加结尾
	if t[last].End < to {
		res = append(res, DataOD{t[last].End + 1, to, 0, 0})
	}

	return res
}

type TickDataODMap map[int64]DataOD

type IndexOD struct {
	Start  int
	End    int
	Status int32
}

type IndexODList []IndexOD

func (t IndexODList) Len() int {
	return len(t)
}

func (t IndexODList) Less(i, j int) bool {
	return t[i].Start < t[j].Start
}

func (t IndexODList) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t IndexODList) Padding(l int) IndexODList {
	res := IndexODList{}

	if len(t) == 0 {
		return res
	}

	// 添加开头
	if t[0].Start > 0 {
		res = append(res, IndexOD{0, t[0].Start, KindFlatCurve})
	}

	last := len(t) - 1

	// 添加相邻区间
	for i := 0; i < last; i++ {
		if t[i].End < t[i+1].Start {
			res = append(res, IndexOD{t[i].End, t[i+1].Start, KindFlatCurve})
		}
	}

	// 添加结尾
	if t[last].End < l {
		res = append(res, IndexOD{t[last].End, l, KindFlatCurve})
	}

	return res
}

type IotData struct {
	Id   string `json:"id"`   //链接点标识
	Tick int64  `json:"tick"` //打包时间 ms
	Data string `json:"data"` //bytes -> base64
}

func AnalysisTick(t0, t1, delta int64) []int64 {
	res := []int64{}

	start := time.Unix(t0, 0).Truncate(time.Minute).Unix()
	for tick := start; tick <= t1; tick += delta {
		res = append(res, tick)
	}

	l := len(res)
	if l == 0 {
		return res
	}

	sp := res[l-1] + delta
	if sp > t1 {
		res = append(res, sp)
	}

	return res
}

type TimeMarker struct {
	Delta float64
	Value float64
}

func TimeDataSlope(v1, v2 TimeData) float64 {
	return ComputeTimeDataSlope(v1.Time, v1.Data, v2.Time, v2.Data)
}

func (t TimeDataSeriesMap) Tick() []int64 {
	res := []int64{}

	for k := range t {
		res = append(res, k)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})

	return res
}

func (t TimeDataSeriesMap) AnalysisTick(delta int64) (tick, marker []int64) {
	l := len(t)
	if l == 0 {
		return
	}

	tick = t.Tick()

	marker = AnalysisTick(tick[0], tick[l-1], delta)

	return
}

func (t TimeDataSeriesMap) Delta(delta int64) (res TimeDataSeries, remain float64) {
	res = TimeDataSeries{}
	remain = 1

	ticks := t.Tick()

	l := len(ticks)
	if l < 2 {
		return
	}

	var tick int64
	var next int64
	var data TimeDataSeries
	ld := 0

	l--
	for i := 0; i < l; i++ {
		tick = ticks[i]
		data = t[tick]
		ld = len(data)

		if ld == 0 {
			continue
		}

		next = ticks[i+1]

		if ld > 1 {
			res = append(res, TimeData{Time: next, Data: data[ld-1].Data - data[0].Data})
			continue
		}

		if next-tick > delta || len(t[next]) == 0 {
			continue
		}

		res = append(res, TimeData{Time: next, Data: t[next][0].Data - data[0].Data})
	}

	tick = ticks[l]
	data = t[tick]
	ld = len(data)
	if ld > 1 {
		length := data[ld-1].Time - data[0].Time
		if length < delta {
			remain = float64(length) / float64(delta)
		}

		res = append(res, TimeData{Time: tick + delta, Data: data[ld-1].Data - data[0].Data})
	}

	return
}

func (t TimeDataSeriesMap) Align(delta int64) TimeDataSeriesMap {
	res := TimeDataSeriesMap{}

	tick, markers := t.AnalysisTick(delta)

	l := len(tick)
	lm := len(markers)
	if l == 0 || lm == 0 {
		return res
	}

	var tickIndex int
	var marker int64
	for _, v := range tick {
		for tickIndex < lm && v >= markers[tickIndex]+delta {
			tickIndex++
		}

		marker = markers[tickIndex]
		res[marker] = append(res[marker], t[v]...)
	}

	return res
}

func (t IotDataPackSeries) Repack(frame int64) IotDataPackSeries {
	res := IotDataPackSeries{}

	m := map[float64]IotDataPack{}

	for _, v := range t {
		station, ok := v[TagStation]
		if !ok {
			continue
		}

		p, ok := m[station]
		if !ok {
			m[station] = IotDataPack{TagStation: station, TagFrame: float64(frame)}
			p = m[station]
		}

		for k2, v2 := range v {
			if k2 == TagFrame || k2 == TagStation {
				continue
			}

			p[k2] = v2
		}

		m[station] = p
	}

	for _, v := range m {
		res = append(res, v)
	}

	return res
}

func (t SensorDataSeries) Len() int {
	return len(t)
}

func (t SensorDataSeries) Less(i, j int) bool {
	return t[i].Frame < t[j].Frame
}

func (t SensorDataSeries) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t SensorDataSeries) Set() SensorDataSeries {
	res := SensorDataSeries{}

	m := map[int64]bool{}

	for _, v := range t {
		if m[v.Frame] {
			continue
		}

		res = append(res, v)
		m[v.Frame] = true
	}

	sort.Sort(res)

	return res
}

func (t SensorDataSeries) MaxTime() int64 {
	if len(t) == 0 {
		return 0
	}

	max := t[0].Time
	for _, v := range t {
		if max < v.Time {
			max = v.Time
		}
	}

	return max
}

func (t SensorDataSeries) MinTime() int64 {
	if len(t) == 0 {
		return 0
	}

	min := t[0].Time
	for _, v := range t {
		if min > v.Time {
			min = v.Time
		}
	}

	return min
}

func (t SensorDataSeries) MaxFrame() int64 {
	if len(t) == 0 {
		return 0
	}

	max := t[0].Frame
	for _, v := range t {
		if max < v.Frame {
			max = v.Frame
		}
	}

	return max
}

func (t SensorDataSeries) MinFrame() int64 {
	if len(t) == 0 {
		return 0
	}

	min := t[0].Frame
	for _, v := range t {
		if min > v.Frame {
			min = v.Frame
		}
	}

	return min
}

func (t SensorDataSeries) DeltaValue() float64 {
	l := len(t)
	if l == 0 {
		return 0
	}

	return t[l-1].Value - t[0].Value
}

func (t SensorDataSeries) Points() []string {
	res := []string{}
	for _, v := range t {
		res = append(res, v.Point)
	}
	res = SetStrList(res)

	return res
}

func (t SensorDataSeries) Map() map[string]SensorDataSeries {
	res := map[string]SensorDataSeries{}

	for _, v := range t {
		_, ok := res[v.Point]
		if !ok {
			res[v.Point] = SensorDataSeries{}
		}

		res[v.Point] = append(res[v.Point], v)
	}

	for k := range res {
		sort.Sort(res[k])
	}

	return res
}

func (t SensorDataSeries) FrameMap() map[string]TimeDataSeries {
	res := map[string]TimeDataSeries{}

	for _, v := range t {
		_, ok := res[v.Point]
		if !ok {
			res[v.Point] = TimeDataSeries{}
		}

		res[v.Point] = append(res[v.Point], v.FrameData())
	}

	for k := range res {
		sort.Sort(res[k])
	}

	return res
}

func (t SensorDataSeries) GroupByBranch() map[int64]SensorDataSeries {
	res := map[int64]SensorDataSeries{}

	for _, v := range t {
		if v.Point == "" {
			continue
		}

		res[v.Branch] = append(res[v.Branch], v)
	}

	return res
}

func (t SensorDataSeries) GroupMsTickByDay() SensorDataSeriesFrameMap {
	res := SensorDataSeriesFrameMap{}

	for _, v := range t {
		tick := time.UnixMilli(v.Frame)
		zero := GetZeroTime(tick)

		res[zero] = append(res[zero], v)
	}

	return res
}

func (t GeoDataSeries) Len() int {
	return len(t)
}

func (t GeoDataSeries) Less(i, j int) bool {
	return t[i].Frame < t[j].Frame
}

func (t GeoDataSeries) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t GeoDataSeries) Set() GeoDataSeries {
	res := GeoDataSeries{}

	m := map[int64]bool{}

	for _, v := range t {
		if m[v.Frame] {
			continue
		}

		res = append(res, v)
		m[v.Frame] = true
	}

	sort.Sort(res)

	return res
}

func (t GeoDataSeries) MaxTime() int64 {
	if len(t) == 0 {
		return 0
	}

	max := t[0].Time
	for _, v := range t {
		if max < v.Time {
			max = v.Time
		}
	}

	return max
}

func (t GeoDataSeries) MinTime() int64 {
	if len(t) == 0 {
		return 0
	}

	min := t[0].Time
	for _, v := range t {
		if min > v.Time {
			min = v.Time
		}
	}

	return min
}

func (t GeoDataSeries) MaxFrame() int64 {
	if len(t) == 0 {
		return 0
	}

	max := t[0].Frame
	for _, v := range t {
		if max < v.Frame {
			max = v.Frame
		}
	}

	return max
}

func (t GeoDataSeries) MinFrame() int64 {
	if len(t) == 0 {
		return 0
	}

	min := t[0].Frame
	for _, v := range t {
		if min > v.Frame {
			min = v.Frame
		}
	}

	return min
}

func (t GeoDataSeries) Points() []string {
	res := []string{}
	for _, v := range t {
		res = append(res, v.Point)
	}
	res = SetStrList(res)

	return res
}

func (t GeoDataSeries) Map() map[string]GeoDataSeries {
	res := map[string]GeoDataSeries{}

	for _, v := range t {
		_, ok := res[v.Point]
		if !ok {
			res[v.Point] = GeoDataSeries{}
		}

		res[v.Point] = append(res[v.Point], v)
	}

	for k := range res {
		sort.Sort(res[k])
	}

	return res
}

func (t GeoDataSeries) FrameMap() map[string]TimeGeoDataSeries {
	res := map[string]TimeGeoDataSeries{}

	for _, v := range t {
		_, ok := res[v.Point]
		if !ok {
			res[v.Point] = TimeGeoDataSeries{}
		}

		res[v.Point] = append(res[v.Point], v.FrameData())
	}

	for k := range res {
		sort.Sort(res[k])
	}

	return res
}

func (t TimeGeoData) MemoryUsage() uintptr {
	return Sizeof(t)
}

func (t TimeGeoDataSeries) Len() int {
	return len(t)
}

func (t TimeGeoDataSeries) Less(i, j int) bool {
	return t[i].Time < t[j].Time
}

func (t TimeGeoDataSeries) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t TimeGeoDataSeries) MemoryUsage() uintptr {
	total := Sizeof(t)
	for _, v := range t {
		total += v.MemoryUsage()
	}

	return total
}

func (t TimeData) MemoryUsage() uintptr {
	return Sizeof(t)
}

func (t TimeDataSeries) Len() int {
	return len(t)
}

func (t TimeDataSeries) Less(i, j int) bool {
	return t[i].Time < t[j].Time
}

func (t TimeDataSeries) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t TimeDataSeries) MemoryUsage() uintptr {
	total := Sizeof(t)
	for _, v := range t {
		total += v.MemoryUsage()
	}

	return total
}

func (t TimeDataSeries) Set() TimeDataSeries {
	res := TimeDataSeries{}

	m := map[int64]bool{}

	for _, v := range t {
		if m[v.Time] {
			continue
		}

		res = append(res, v)
		m[v.Time] = true
	}

	sort.Sort(res)

	return res
}

func (t TimeDataSeries) NormalSet() TimeDataSeries {
	res := TimeDataSeries{}

	m := map[int64]bool{}

	for _, v := range t {
		tick := time.Unix(v.Time, 0).Truncate(time.Minute).Unix()
		if m[tick] {
			continue
		}

		item := TimeData{
			Time: tick,
			Data: v.Data,
		}
		res = append(res, item)

		m[v.Time] = true
	}

	sort.Sort(res)

	return res
}

func (t TimeDataSeries) CumulFilter() TimeDataSeries {
	res := TimeDataSeries{}

	for i, v := range t {
		if i > 0 && v.Data < t[i-1].Data {
			v.Data = t[i-1].Data
		}

		res = append(res, v)
	}

	return res
}

func (t TimeDataSeries) MapSerial() []map[string]interface{} {
	res := []map[string]interface{}{}

	for _, v := range t {
		item := map[string]interface{}{
			"time": v.Time,
			"data": v.Data,
		}

		if math.IsNaN(v.Data) {
			item["data"] = "NaN"
		}

		res = append(res, item)
	}

	return res
}

func (t TimeDataSeries) Zip() (res TimeDataSeries) {
	res = TimeDataSeries{}

	l := len(t)
	if l <= 2 {
		return t
	}

	res = append(res, t[0])

	n := l - 1
	for i := 1; i < n; i++ {
		if t[i].Data == t[i-1].Data && t[i].Data == t[i+1].Data {
			continue
		}

		res = append(res, t[i])
	}

	res = append(res, t[n])

	return
}

func (t TimeDataSeries) Zip2(interval int64) (res TimeDataSeries) {
	res = TimeDataSeries{}

	l := len(t)
	if l <= 2 {
		return t
	}

	res = append(res, t[0])

	n := l - 1
	var cursor int64
	for i := 1; i < n; i++ {
		if t[i].Data == t[i-1].Data && t[i].Data == t[i+1].Data {
			if t[i].Time-cursor > interval {
				cursor = t[i].Time
				res = append(res, t[i])
			}

			continue
		}

		cursor = t[i].Time
		res = append(res, t[i])
	}

	res = append(res, t[n])

	return
}

func (t TimeDataSeries) MaxTime() int64 {
	if len(t) == 0 {
		return 0
	}

	max := t[0].Time
	for _, v := range t {
		if max < v.Time {
			max = v.Time
		}
	}

	return max
}

func (t TimeDataSeries) MinTime() int64 {
	if len(t) == 0 {
		return 0
	}

	min := t[0].Time
	for _, v := range t {
		if min > v.Time {
			min = v.Time
		}
	}

	return min
}

func (t TimeDataSeries) Add(n float64) TimeDataSeries {
	for i := range t {
		t[i].Add(n)
	}

	return t
}

func (t TimeDataSeries) MinusTo(n float64) TimeDataSeries {
	for i := range t {
		t[i].MinusTo(n)
	}

	return t
}

func (t TimeDataSeries) Mul(n float64) TimeDataSeries {
	for i := range t {
		t[i].Mul(n)
	}

	return t
}

func (t TimeDataSeries) Div(n float64) TimeDataSeries {
	for i := range t {
		t[i].Div(n)
	}

	return t
}

func (t TimeDataSeries) DecimalPlaces(n int) TimeDataSeries {
	for i := range t {
		t[i].DecimalPlaces(n)
	}

	return t
}

func (t TimeDataSeries) TakeDigits(n int) TimeDataSeries {
	for i := range t {
		t[i].TakeDigits(n)
	}

	return t
}

func (t TimeDataSeries) SeekTick(tick int64) (res int) {
	left := 0
	right := len(t) - 1
	if right < 0 || t[right].Time < tick {
		return -1
	}

	if tick < t[0].Time {
		return 0
	}

	for left <= right {
		mid := (left + right) / 2
		if t[mid].Time == tick {
			return mid
		} else if t[mid].Time < tick {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return right
}

func (t TimeDataSeries) From(start int64) TimeDataSeries {
	index := t.SeekTick(start)
	if index > -1 {
		return t[index:]
	}

	return TimeDataSeries{}
}

func (t TimeDataSeries) Tick() []int64 {
	res := []int64{}

	for _, v := range t {
		res = append(res, v.Time)
	}

	return res
}

func (t TimeDataSeries) Last() TimeData {
	l := len(t)
	if l == 0 {
		return TimeData{}
	}

	return t[l-1]
}

func (t TimeDataSeries) Max() float64 {
	if len(t) == 0 {
		return 0
	}

	max := t[0].Data
	for _, v := range t {
		if max < v.Data {
			max = v.Data
		}
	}

	return max
}

func (t TimeDataSeries) Min() float64 {
	if len(t) == 0 {
		return 0
	}

	min := t[0].Data
	for _, v := range t {
		if min > v.Data {
			min = v.Data
		}
	}

	return min
}

func (t TimeDataSeries) FindKthLargestTimeData(k int) TimeData {
	l := len(t)
	if k < 1 || k > l {
		return TimeData{}
	}

	pivot := t[0]
	left := 0
	right := len(t) - 1

	for i := 1; i <= right; {
		if t[i].Data < pivot.Data {
			t[left], t[i] = t[i], t[left]
			i++
			left++
		} else if t[i].Data > pivot.Data {
			t[right], t[i] = t[i], t[right]
			right--
		} else {
			i++
		}
	}

	if k <= left {
		return t[:left].FindKthLargestTimeData(k)
	} else if k > right+1 {
		return t[right+1:].FindKthLargestTimeData(k - right - 1)
	} else {
		return pivot
	}
}

func (t TimeDataSeries) Avg() float64 {
	l := len(t)
	if l == 0 {
		return math.NaN()
	}

	var sum float64
	for _, v := range t {
		sum += v.Data
	}

	return sum / float64(l)
}

func (t TimeDataSeries) Var(mean float64) float64 {
	l := len(t)
	if l == 0 {
		return math.NaN()
	}

	sum := 0.0
	for _, v := range t {
		sum += math.Pow(v.Data-mean, 2)
	}

	return sum / float64(l)
}

func (t TimeDataSeries) Static() (max, min, avg float64) {
	l := len(t)
	if l == 0 {
		return math.NaN(), math.NaN(), math.NaN()
	}

	max = t[0].Data
	min = t[0].Data
	var sum float64

	for _, v := range t {
		if max < v.Data {
			max = v.Data
		} else if min > v.Data {
			min = v.Data
		}

		sum += v.Data
	}

	avg = sum / float64(l)

	return
}

func (t TimeDataSeries) Delta() (res TimeDataSeries) {
	res = TimeDataSeries{}

	l := len(t)
	if l < 2 {
		return res
	}

	l--
	for i := 0; i < l; i++ {
		item := TimeData{
			Time: t[i+1].Time,
			Data: t[i+1].Data - t[i].Data,
		}
		res = append(res, item)
	}

	return res
}

func (t TimeDataSeries) Accrue(id ...IndexOD) TimeDataSeries {
	res := TimeDataSeries{}

	l := len(t)
	if l == 0 {
		return res
	}

	idx := 0
	start := 0
	end := 0
	var sum float64

	for _, od := range id {
		if idx > l {
			break
		}

		start = MaxInt(start, od.Start)
		end = MinInt(od.End, l)

		if od.Status == KindAccumulationCurve {
			sum = 0
			for i := start; i < end; i++ {
				sum += t[i].Data
				item := TimeData{
					Time: t[i].Time,
					Data: sum,
				}
				res = append(res, item)
			}

			last := res[len(res)-1]
			last.Time++
			last.Data = math.NaN()

			res = append(res, last)

		} else {

			for i := start; i < end; i++ {
				item := TimeData{
					Time: t[i].Time,
					Data: t[i].Data,
				}
				res = append(res, item)
			}
		}

		idx = end
	}

	if idx < l {
		res = append(res, t[idx:]...)
	}

	return res
}

func (t TimeDataSeries) AnalysisTick(delta int64) []int64 {
	res := []int64{}

	l := len(t)
	if l == 0 {
		return res
	}

	start := t[0].Time
	//todo: valid end
	end := t[l-1].Time

	return AnalysisTick(start, end, delta)
}

func (t TimeDataSeries) Align(delta int64, markers []int64) TimeDataSeriesMap {
	res := TimeDataSeriesMap{}

	l := len(t)
	lm := len(markers)
	if l == 0 || lm == 0 {
		return res
	}

	var tickIndex int
	var marker int64
	for _, v := range t {
		for tickIndex < lm && v.Time >= markers[tickIndex]+delta {
			tickIndex++
		}

		marker = markers[tickIndex]
		res[marker] = append(res[marker], v)
	}

	return res
}

func (t TimeDataSeries) AlignByDelta(delta int64) TimeDataSeriesMap {
	ticks := t.AnalysisTick(delta)
	return t.Align(delta, ticks)
}

func (t TimeDataSeries) Analysis(delta int64) (res TimeDataSeries, remain float64) {
	return t.AlignByDelta(delta).Delta(delta)
}

func (t TimeDataSeries) Normalize(delta, end int64, blankEnd bool, decimal ...int) ValuePairList {
	res := ValuePairList{}

	l := len(t)
	if l == 0 {
		if blankEnd {
			res = append(res, SomeBlankTickBefore(end, 6)...)
		}

		return res
	}

	if l == 1 {
		res = append(res, t[0].TakeDigitsValue(decimal...))
		if blankEnd && t[l-1].Time < end {
			res = append(res, SomeBlankTickBefore(end, 6)...)
		}

		return res
	}

	res = append(res, t[0].TakeDigitsValue(decimal...))

	for i := 1; i < l; i++ {
		if t[i].Time-t[i-1].Time > delta {
			res = append(res, SomeBlankTickBefore(t[i].Time, 6)...)
		}

		res = append(res, t[i].TakeDigitsValue(decimal...))
	}

	if blankEnd && t[l-1].Time < end {
		res = append(res, SomeBlankTickBefore(end, 6)...)
	}

	return res
}

func (t TimeDataSeries) NormalizeNull(delta, end int64, blankEnd bool, decimal ...int) ValuePairList {
	res := ValuePairList{}

	l := len(t)
	if l == 0 {
		if blankEnd {
			res = append(res, SomeBlankTickBefore(end, 6)...)
		}

		return res
	}

	if l == 1 {
		res = append(res, t[0].TakeDigitsValue(decimal...))
		if blankEnd && t[l-1].Time < end {
			res = append(res, SomeBlankTickBefore(end, 6)...)
		}

		return res
	}

	res = append(res, t[0].TakeDigitsValue(decimal...))

	for i := 1; i < l; i++ {
		if t[i].Time-t[i-1].Time > delta {
			res = append(res, SomeBlankTickBefore(t[i].Time, 6)...)
		}

		res = append(res, t[i].TakeDigitsValue(decimal...))
	}

	if blankEnd && t[l-1].Time < end {
		res = append(res, SomeBlankTickBefore(end, 6)...)
	}

	return res
}

func (t TimeDataSeries) Husk(od DataODList) TimeDataSeries {
	res := TimeDataSeries{}

	l := len(t)
	if l == 0 {
		return res
	}

	lo := len(od)
	if lo == 0 {
		return t
	}

	odIndex := 0
	for i := 0; i < l; i++ {
		for odIndex < lo && t[i].Time > od[odIndex].End {
			odIndex++
		}

		t[i].Data -= od[odIndex].Value
		res = append(res, t[i])
	}

	return res
}

func (t TimeDataSeries) Slicing(od ...DataOD) IndexODList {
	res := IndexODList{}

	l := len(t)
	if l == 0 {
		return res
	}

	lo := len(od)
	if lo == 0 {
		return IndexODList{{0, l, 1}}
	}

	//通过二分查找生成稀疏的IndexODList，复杂度O(k * ln(n))
	for _, d := range od {
		i := sort.Search(l, func(i int) bool {
			return t[i].Time >= d.Start
		})
		if i == l || t[i].Time > d.End {
			continue
		}

		j := sort.Search(l-i, func(j int) bool {
			return t[i+j].Time > d.End
		})
		if j == 0 {
			continue
		}

		end := i + j
		if j == -1 {
			end = l
		}
		res = append(res, IndexOD{i, end, d.Status})
	}

	if len(res) == 0 {
		return IndexODList{{0, l, KindAccumulationCurve}}
	}

	//res = append(res, res.Padding(l)...)
	sort.Slice(res, func(i, j int) bool {
		if res[i].Start == res[j].Start {
			return res[i].End < res[j].End
		}

		return res[i].Start < res[j].Start
	})

	return res
}

func (t TimeDataSeries) Reform(od ...DataOD) TimeDataSeries {
	td := t.Set().Delta()
	id := td.Slicing(od...)
	ts := td.Accrue(id...)

	return ts
}

func (t TimeDataSeries) Power(s float64) float64 {
	l := len(t)
	if l == 0 {
		return 0.00
	}

	return (t[l-1].Data - t[0].Data) / (float64(t[l-1].Time-t[0].Time) / s)
}

func (t TimeDataSeries) Repair() TimeDataSeries {
	res := TimeDataSeries{}
	l := len(t)
	if l == 0 {
		return res
	}

	for i := 1; i < l; i++ {
		item := t[i]
		if item.Data == 0 {
			item.Data = t[i-1].Data
		}

		res = append(res, item)
	}

	return res
}

func (t TimeDataSeries) Restore(s ...TimeData) TimeDataSeries {
	ls := len(s)
	if ls == 0 {
		return t
	}

	n := len(t)
	var res TimeDataSeries = make([]TimeData, n)

	current := ls - 1
	n--
	for i := n; i >= 0; i-- {
		v := t[i]
		for current > 0 && s[current].Time > v.Time {
			current--
		}

		v.Data += s[current].Data
		res[i] = v
	}

	return res
}

func (t TimeDataSeries) MinusRestore(s ...TimeData) TimeDataSeries {
	ls := len(s)
	if ls == 0 {
		return t
	}

	n := len(t)
	var res TimeDataSeries = make([]TimeData, n)

	current := ls - 1
	n--
	for i := n; i >= 0; i-- {
		v := t[i]
		for current > 0 && s[current].Time > v.Time {
			current--
		}

		v.Data = s[current].Data - v.Data
		res[i] = v
	}

	return res
}

func (t TimeDataSeries) Split(maxInterval int64) (res []TimeDataSeries) {
	res = []TimeDataSeries{}

	l := len(t) - 1
	if l <= 0 {
		return []TimeDataSeries{t}
	}

	res = append(res, TimeDataSeries{t[0]})
	m := 0

	for i := 1; i < l; i++ {
		if t[i].Time-t[i-1].Time > maxInterval {
			res = append(res, TimeDataSeries{})
			m++
		}

		res[m] = append(res[m], t[i])
	}

	if t[l].Time-t[l-1].Time <= maxInterval {
		res[m] = append(res[m], t[l])
	} else {
		res = append(res, TimeDataSeries{t[l]})
	}

	return res
}

func (t TimeDataSeries) FillGapWithLastValue(std int64) (res TimeDataSeries) {
	l := len(t)
	if l < 2 {
		return TimeDataSeries{}
	}

	res = append(res, t[0])

	maxDelta := t[l-1].Time - t[0].Time
	maxLoop := int(maxDelta/std) + 1

	for i := 1; i < l; i++ {
		delta := t[i].Time - t[i-1].Time
		if delta <= std {
			res = append(res, t[i])
			continue
		}

		for n := 1; n <= maxLoop; n++ {
			it := t[i-1].Time + std*int64(n)
			if it >= t[i].Time {
				res = append(res, t[i])
				break
			}

			item := TimeData{
				Time: it,
				Data: t[i-1].Data,
			}

			res = append(res, item)
		}
	}

	return
}

func (t TimeDataSeries) FillGap(std int64) (res TimeDataSeries) {
	l := len(t)
	if l < 2 {
		return TimeDataSeries{}
	}

	res = append(res, t[0])

	maxDelta := t[l-1].Time - t[0].Time
	maxLoop := int(maxDelta/std) + 1

	for i := 1; i < l; i++ {
		delta := t[i].Time - t[i-1].Time
		if delta <= std {
			res = append(res, t[i])
			continue
		}

		ratio := (t[i].Data - t[i-1].Data) / (float64(delta))

		for n := 1; n <= maxLoop; n++ {
			it := t[i-1].Time + std*int64(n)
			if it >= t[i].Time {
				res = append(res, t[i])
				break
			}

			iv := t[0].Data + ratio*(float64(it-t[0].Time))
			item := TimeData{
				Time: it,
				Data: iv,
			}

			res = append(res, item)
		}
	}

	return
}

func (t TimeDataSeries) FillGapWithNaN(std int64) (res TimeDataSeries) {
	l := len(t)
	if l < 2 {
		return TimeDataSeries{}
	}

	res = append(res, t[0])

	maxDelta := t[l-1].Time - t[0].Time
	maxLoop := int(maxDelta/std) + 1

	for i := 1; i < l; i++ {
		delta := t[i].Time - t[i-1].Time
		if delta <= std {
			res = append(res, t[i])
			continue
		}

		for n := 1; n <= maxLoop; n++ {
			it := t[i-1].Time + std*int64(n)
			if it >= t[i].Time {
				res = append(res, t[i])
				break
			}

			item := TimeData{
				Time: it,
				Data: math.NaN(),
			}

			res = append(res, item)
		}
	}

	return
}

type TickDataSeries struct {
	cap int
	l   sync.RWMutex
	m   TimeDataSeries
	t   int64
	v   float64
}

func NewTickDataSeries(n int) *TickDataSeries {
	res := &TickDataSeries{
		cap: n,
		m:   make(TimeDataSeries, 0, n),
		t:   math.MinInt64,
		v:   math.MinInt64,
	}

	return res
}

func (t *TickDataSeries) Output() []string {
	t.l.RLock()
	defer t.l.RUnlock()

	res := []string{fmt.Sprintf("cap: %v", t.cap)}

	for _, v := range t.m {
		res = append(res, ToJsonStr(v))
	}

	return res
}

func (t *TickDataSeries) Init(raw TimeDataSeries) {
	l := len(raw)
	if l > t.cap {
		raw = raw[l-t.cap:]
	}

	t.m = append(t.m, raw...)
	t.t = t.m.MaxTime()
	t.v = t.m.Max()
}

func (t *TickDataSeries) Last() TimeData {
	t.l.RLock()
	defer t.l.RUnlock()

	l := len(t.m)
	if l == 0 {
		return TimeData{}
	}

	return t.m[l-1]
}

func (t *TickDataSeries) MaxTime() int64 {
	t.l.RLock()
	defer t.l.RUnlock()

	return t.t
}

func (t *TickDataSeries) Max() float64 {
	t.l.RLock()
	defer t.l.RUnlock()

	return t.v
}

func (t *TickDataSeries) AppendTickData(raw TimeData) {
	t.l.Lock()
	defer t.l.Unlock()

	l := len(t.m)
	if l > 1 {
		if t.m[l-1].Time == raw.Time {
			return
		}
	}

	if raw.Time > t.t {
		t.t = raw.Time
	}

	if raw.Data > t.v {
		t.v = raw.Data
	}

	if l >= t.cap {
		t.m = append(t.m[1:], raw)
		return
	}

	t.m = append(t.m, raw)
}

func (t *TickDataSeries) Series() TimeDataSeries {
	t.l.RLock()
	raw := t.m
	t.l.RUnlock()

	res := make(TimeDataSeries, len(raw))
	copy(res, raw)

	return res
}

func (t *TickDataSeries) Set() *TickDataSeries {
	t.l.RLock()
	defer t.l.RUnlock()

	m := map[int64]bool{}
	data := TimeDataSeries{}

	for _, v := range t.m {
		if m[v.Time] {
			continue
		}

		data = append(data, v)
		m[v.Time] = true
	}

	sort.Sort(data)

	res := NewTickDataSeries(t.cap)
	res.Init(data)

	return res
}

func (t *TickDataSeries) MemoryUsage() uintptr {
	t.l.RLock()
	defer t.l.RUnlock()

	total := Sizeof(t)
	total += t.m.MemoryUsage()

	return total
}

type TickGeoDataSeries struct {
	cap int
	l   sync.RWMutex
	m   TimeGeoDataSeries
}

func NewTickGeoDataSeries(n int) *TickGeoDataSeries {
	return &TickGeoDataSeries{cap: n, m: make(TimeGeoDataSeries, 0, n)}
}

func (t *TickGeoDataSeries) Output() []string {
	t.l.RLock()
	defer t.l.RUnlock()

	res := []string{fmt.Sprintf("cap: %v", t.cap)}

	for _, v := range t.m {
		res = append(res, ToJsonStr(v))
	}

	return res
}

func (t *TickGeoDataSeries) Init(raw TimeGeoDataSeries) {
	l := len(raw)
	if l > t.cap {
		raw = raw[l-t.cap:]
	}

	t.m = append(t.m, raw...)
}

func (t *TickGeoDataSeries) Last() TimeGeoData {
	t.l.RLock()
	defer t.l.RUnlock()

	l := len(t.m)
	if l == 0 {
		return TimeGeoData{}
	}

	return t.m[l-1]
}

func (t *TickGeoDataSeries) AppendTickData(raw TimeGeoData) {
	t.l.Lock()
	defer t.l.Unlock()

	l := len(t.m)
	if l > 1 {
		if t.m[l-1].Time == raw.Time {
			return
		}
	}

	if l >= t.cap {
		t.m = append(t.m[1:], raw)
		return
	}

	t.m = append(t.m, raw)
}

func (t *TickGeoDataSeries) Series() TimeGeoDataSeries {
	t.l.RLock()
	raw := t.m
	t.l.RUnlock()

	res := make(TimeGeoDataSeries, len(raw))
	copy(res, raw)

	return res
}

func (t *TickGeoDataSeries) Set() *TickGeoDataSeries {
	t.l.RLock()
	defer t.l.RUnlock()

	m := map[int64]bool{}
	data := TimeGeoDataSeries{}

	for _, v := range t.m {
		if m[v.Time] {
			continue
		}

		data = append(data, v)
		m[v.Time] = true
	}

	sort.Sort(data)

	res := NewTickGeoDataSeries(t.cap)
	res.Init(data)

	return res
}

func (t *TickGeoDataSeries) MemoryUsage() uintptr {
	t.l.RLock()
	defer t.l.RUnlock()

	total := Sizeof(t)
	total += t.m.MemoryUsage()

	return total
}

func FillGap(std int64, raw ...TimeDataSeries) TimeDataSeries {
	res := TimeDataSeries{}

	l := len(raw)
	for i := 0; i < l; i++ {
		res = append(res, raw[i].FillGap(std)...)

		if i < l-1 {
			s0 := TimeData{
				Time: raw[i][len(raw[i])-1].Time + std,
				Data: math.NaN(),
			}
			s1 := TimeData{
				Time: raw[i+1][0].Time - std,
				Data: math.NaN(),
			}
			res = append(res, TimeDataSeries{s0, s1}.FillGapWithNaN(std)...)
		}
	}

	return res
}

func FillGapWithNaN(std int64, raw ...TimeDataSeries) TimeDataSeries {
	res := TimeDataSeries{}

	l := len(raw)
	for i := 0; i < l; i++ {
		res = append(res, raw[i].FillGapWithNaN(std)...)

		if i < l-1 {
			s0 := TimeData{
				Time: raw[i][len(raw[i])-1].Time + std,
				Data: math.NaN(),
			}
			s1 := TimeData{
				Time: raw[i+1][0].Time - std,
				Data: math.NaN(),
			}
			res = append(res, TimeDataSeries{s0, s1}.FillGapWithNaN(std)...)
		}
	}

	return res
}

func SomeStaticValue(n int, suffix string, value string) StaticValueList {
	res := StaticValueList{}

	for i := 0; i < n; i++ {
		item := StaticValue{fmt.Sprintf("%d%s", i+1, suffix), value}
		res = append(res, item)
	}

	return res
}

func SomeStaticZero(n int, suffix string) StaticValueList {
	res := StaticValueList{}

	for i := 0; i < n; i++ {
		item := StaticValue{fmt.Sprintf("%d%s", i+1, suffix), "0"}
		res = append(res, item)
	}

	return res
}
