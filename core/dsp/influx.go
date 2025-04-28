package dsp

import (
	"encoding/json"
	"math"
	. "mykit/core/types"
	"sort"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	influxModel "github.com/influxdata/influxdb/models"
)

func NewInfluxPoint(name string, tags map[string]string,
	fields map[string]interface{}, t ...time.Time) *client.Point {
	res, err := client.NewPoint(name, tags, fields, t...)
	if err != nil {
		LogS1.Error("NewInfluxPoint",
			LogError(err),
		)
	}

	return res
}

func MergeInfluxResult(raw []client.Result) (res []influxModel.Row) {
	res = []influxModel.Row{}

	if len(raw) == 0 {
		return res
	}

	for _, v := range raw {
		res = append(res, v.Series...)
	}

	return res
}

type PointData struct {
	Time       int64   `json:"time,omitempty"`
	Value      float64 `json:"value,omitempty"`
	PointId    int64   `json:"pointId,omitempty"`
	ResourceId int64   `json:"resourceId,omitempty"`
	PlanId     int64   `json:"planId,omitempty"`
}

func (t PointData) ToPoint(name string, tick ...time.Time) *client.Point {
	return InfluxPointData(name, t.PointId, t.ResourceId, t.Value, tick...)
}

func (t PointData) ValuePair() ValuePair {
	return ValuePair{t.Time, t.Value}
}

func (t PointData) NamedValue(raw string) ValuePair {
	return ValuePair{raw, t.Value}
}

func (t PointData) TimeData() TimeData {
	return TimeData{t.Time, t.Value}
}

func (t PointData) TakeDigitsValue(n int) float64 {
	return TakeDigits(t.Value, n)
}

func InfluxPointData(name string, point, resource int64,
	value float64, tick ...time.Time) *client.Point {
	tags := map[string]string{
		"pointId": ParseInt64ToStr(point),
	}

	fields := map[string]interface{}{
		"value":      value,
		"resourceId": resource,
	}

	res := NewInfluxPoint(name, tags, fields, ParseTime(tick))

	return res
}

func NewPointData(raw []interface{}) PointData {
	res := PointData{}

	if len(raw) < 4 {
		return res
	}

	index := 0
	if raw[index] != nil {
		res.Time = ParseTimeFromResp(raw[index]).Unix()
	}

	index++
	if raw[index] != nil {
		v, ok := raw[index].(string)
		if ok {
			res.PointId, _ = ParseInt64FromStr(v)
		}
	}

	index++
	if raw[index] != nil {
		v, ok := raw[index].(json.Number)
		if ok {
			res.ResourceId, _ = v.Int64()
		}
	}

	index++
	if raw[index] == nil {
		res.Value = math.NaN()
	} else {
		v, ok := raw[index].(json.Number)
		if ok {
			res.Value, _ = v.Float64()
		}
	}

	return res
}

func FirstPointData(raw []client.Result) PointData {
	res := PointData{}

	rows := MergeInfluxResult(raw)
	if len(rows) == 0 {
		return res
	}

	values := rows[0].Values
	if len(values) == 0 {
		return res
	}

	res = NewPointData(values[0])

	return res
}

type PointDataSeries []PointData

func NewPointDataSeries(raw []client.Result) PointDataSeries {
	res := PointDataSeries{}

	rows := MergeInfluxResult(raw)
	if len(rows) == 0 {
		return res
	}

	for _, row := range rows {
		for _, line := range row.Values {
			res = append(res, NewPointData(line))
		}
	}

	return res
}

func (t PointDataSeries) Len() int {
	return len(t)
}

func (t PointDataSeries) Less(i, j int) bool {
	return t[i].Time < t[j].Time
}

func (t PointDataSeries) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t PointDataSeries) AscByTime(raw ...bool) PointDataSeries {
	var sortFunc func(i, j int) bool

	if ParseBool(raw) {
		sortFunc = func(i, j int) bool {
			return t[i].Time < t[j].Time
		}
	} else {
		sortFunc = func(i, j int) bool {
			return t[i].Time > t[j].Time
		}
	}

	sort.Slice(t, func(i, j int) bool {
		return sortFunc(i, j)
	})

	return t
}

func (t PointDataSeries) FilterNaN() PointDataSeries {
	res := PointDataSeries{}

	for _, v := range t {
		if math.IsNaN(v.Value) {
			continue
		}

		res = append(res, v)
	}

	return res
}

func (t PointDataSeries) ValuePair() ValuePairList {
	res := ValuePairList{}

	for _, v := range t {
		res = append(res, v.ValuePair())
	}

	return res
}

func (t PointDataSeries) Map() map[int64]PointDataSeries {
	res := map[int64]PointDataSeries{}

	for _, v := range t {
		_, ok := res[v.PointId]
		if !ok {
			res[v.PointId] = PointDataSeries{}
		}

		res[v.PointId] = append(res[v.PointId], v)
	}

	for k := range res {
		sort.Slice(res[k], func(i, j int) bool {
			return res[k][i].Time < res[k][j].Time
		})
	}

	return res
}

func (t PointDataSeries) TimeData() TimeDataSeries {
	res := TimeDataSeries{}

	for _, v := range t {
		res = append(res, v.TimeData())
	}

	return res
}

type SensorData struct {
	Time    int64   `json:"time"`    //时间
	Point   string  `json:"point"`   //点位
	Kind    int64   `json:"kind"`    //传感器类别 2.雨量 3.水位
	Channel int64   `json:"channel"` //通道
	Branch  int64   `json:"branch"`  //所属分支
	Frame   int64   `json:"frame"`   //帧号
	Value   float64 `json:"value"`   //数值
}

func (t SensorData) ToPoint(name string, tick ...time.Time) *client.Point {
	return InfluxSensorData(name, t.Point, t.Kind, t.Channel, t.Branch, t.Frame, t.Value, tick...)
}

func (t SensorData) TimeData() TimeData {
	return TimeData{t.Time, t.Value}
}

func (t SensorData) FrameData() TimeData {
	return TimeData{t.Frame, t.Value}
}

func (t SensorData) ValuePair() ValuePair {
	return ValuePair{t.Time, t.Value}
}

func (t SensorData) FrameValue() ValuePair {
	return ValuePair{t.Frame, t.Value}
}

func (t SensorData) NamedValue(raw string) ValuePair {
	return ValuePair{raw, t.Value}
}

func InfluxSensorData(name, point string, kind, channel, branch, frame int64,
	value float64, tick ...time.Time) *client.Point {
	tags := map[string]string{
		"point":   point,
		"kind":    ParseInt64ToStr(kind),
		"channel": ParseInt64ToStr(channel),
		"branch":  ParseInt64ToStr(branch),
	}

	fields := map[string]interface{}{
		"frame": frame,
		"value": value,
	}

	res := NewInfluxPoint(name, tags, fields, ParseTime(tick))

	return res
}

func NewSensorData(raw []interface{}) SensorData {
	res := SensorData{}
	if len(raw) < 7 {
		return res
	}

	index := 0
	res.Time = ParseTimeFromResp(raw[index]).Unix()

	index++
	res.Branch = ParseInt64FromTag(raw[index])

	index++
	res.Channel = ParseInt64FromTag(raw[index])

	index++
	res.Frame = ParseInt64FromField(raw[index])

	index++
	res.Kind = ParseInt64FromTag(raw[index])

	index++
	res.Point = ParseStrFromTag(raw[index])

	index++
	res.Value = ParseFloat64FromField(raw[index])

	return res
}

func FirstSensorData(raw []client.Result) SensorData {
	res := SensorData{}

	rows := MergeInfluxResult(raw)
	if len(rows) == 0 {
		return res
	}

	values := rows[0].Values
	if len(values) == 0 {
		return res
	}

	res = NewSensorData(values[0])

	return res
}

type SensorDataSeries []SensorData

func NewSensorDataSeries(raw []client.Result) SensorDataSeries {
	res := SensorDataSeries{}

	rows := MergeInfluxResult(raw)
	if len(rows) == 0 {
		return res
	}

	for _, row := range rows {
		for _, line := range row.Values {
			res = append(res, NewSensorData(line))
		}
	}

	return res
}

func (t SensorDataSeries) InfluxPoint() InfluxPointSeries {
	res := InfluxPointSeries{}

	for _, v := range t {
		res = append(res, v)
	}

	return res
}

func (t SensorDataSeries) FirstFrame() SensorData {
	if len(t) == 0 {
		return SensorData{}
	}

	return t[0]
}

func (t SensorDataSeries) FilterNaN() SensorDataSeries {
	res := SensorDataSeries{}

	for _, v := range t {
		if math.IsNaN(v.Value) {
			continue
		}

		res = append(res, v)
	}

	return res
}

func (t SensorDataSeries) TimeData() TimeDataSeries {
	res := TimeDataSeries{}

	for _, v := range t {
		res = append(res, v.TimeData())
	}

	return res
}

func (t SensorDataSeries) ValuePair() ValuePairList {
	res := ValuePairList{}

	for _, v := range t {
		res = append(res, v.ValuePair())
	}

	return res
}

func (t SensorDataSeries) Delta() TimeDataSeries {
	res := TimeDataSeries{}

	l := len(t)
	if l < 2 {
		return res
	}

	l--
	for i := 0; i < l; i++ {
		item := TimeData{
			Time: t[i+1].Frame,
			Data: t[i+1].Value - t[i].Value,
		}
		res = append(res, item)
	}

	return res
}

func (t SensorDataSeries) FrameData() TimeDataSeries {
	res := TimeDataSeries{}

	for _, v := range t {
		res = append(res, v.FrameData())
	}

	return res
}

func (t SensorDataSeries) FrameDataSet() TimeDataSeries {
	res := TimeDataSeries{}

	m := map[int64]bool{}

	for _, v := range t {
		if m[v.Frame] {
			continue
		}

		res = append(res, v.FrameData())
		m[v.Frame] = true
	}

	sort.Sort(res)

	return res
}

func (t SensorDataSeries) FrameValue() ValuePairList {
	res := ValuePairList{}

	for _, v := range t {
		res = append(res, v.FrameValue())
	}

	return res
}

type TimeData struct {
	Time int64   `json:"time"`
	Data float64 `json:"data"`
}

func (t *TimeData) TakeDigits(n int) {
	t.Data = TakeDigits(t.Data, n)
}

func (t *TimeData) Add(n float64) {
	t.Data += n
}

func (t *TimeData) MinusTo(n float64) {
	t.Data = n - t.Data
}

func (t *TimeData) Mul(n float64) {
	t.Data *= n
}

func (t *TimeData) Div(n float64) {
	if n == 0 {
		return
	}

	t.Data /= n
}

func (t *TimeData) DecimalPlaces(decimal int) {
	t.Data = DecimalPlaces(t.Data, decimal)
}

func (t TimeData) NaNData(n int64) TimeData {
	return TimeData{Time: t.Time + n, Data: math.NaN()}
}

func (t TimeData) SomeNaNBefore(n ...int) TimeDataSeries {
	return t.SomeNaN(-1, n...)
}

func (t TimeData) SomeNaN(delta int64, n ...int) TimeDataSeries {
	res := TimeDataSeries{}
	times := ParseIntParam(n, 1)
	for i := 0; i < times; i++ {
		res = append(res, t.NaNData(int64(i+1)*delta))
	}

	return res
}

func (t TimeData) ValuePair() ValuePair {
	if math.IsNaN(t.Data) {
		return ValuePair{t.Time, ""}
	}

	return ValuePair{t.Time, t.Data}
}

func (t TimeData) NamedValue(raw string) ValuePair {
	return ValuePair{raw, t.Data}
}

func (t TimeData) TakeDigitsValue(n ...int) ValuePair {
	if len(n) == 0 {
		return t.ValuePair()
	}

	if math.IsNaN(t.Data) {
		return ValuePair{t.Time, ""}
	}

	return ValuePair{t.Time, TakeDigits(t.Data, n[0])}
}

func (t TimeData) TakeDigitsNamedValue(raw string, n int) ValuePair {
	return ValuePair{raw, TakeDigits(t.Data, n)}
}

func (t TimeData) Slope(frame TimeData) float64 {
	return TimeDataSlope(t, frame)
}

func (t TimeData) Fitting(frame TimeData, tick int64) float64 {
	slope := t.Slope(frame)
	return t.Data + slope*(float64(tick-t.Time))
}

func NewTimeData(raw []interface{}) TimeData {
	res := TimeData{}
	if len(raw) < 2 {
		return res
	}

	index := 0
	res.Time = ParseTimeFromResp(raw[index]).Unix()

	index++
	res.Data = ParseFloat64FromField(raw[index])

	return res
}

type TimeDataSeries []TimeData

func NewTimeDataSeries(raw []client.Result) TimeDataSeries {
	res := TimeDataSeries{}

	rows := MergeInfluxResult(raw)
	if len(rows) == 0 {
		return res
	}

	for _, row := range rows {
		for _, line := range row.Values {
			res = append(res, NewTimeData(line))
		}
	}

	return res
}

func (t TimeDataSeries) FirstFrame() TimeData {
	if len(t) == 0 {
		return TimeData{}
	}

	return t[0]
}

func (t TimeDataSeries) FilterNaN() TimeDataSeries {
	res := TimeDataSeries{}

	for _, v := range t {
		if math.IsNaN(v.Data) {
			continue
		}

		res = append(res, v)
	}

	return res
}

func (t TimeDataSeries) ValuePair() ValuePairList {
	res := ValuePairList{}

	for _, v := range t {
		res = append(res, v.ValuePair())
	}

	return res
}

func (t TimeDataSeries) TakeDigitsValue(n int) ValuePairList {
	res := ValuePairList{}

	for _, v := range t {
		res = append(res, v.TakeDigitsValue(n))
	}

	return res
}

func (t TimeDataSeries) Sum() float64 {
	var res float64

	for _, v := range t {
		res += v.Data
	}

	return res
}

type GeoData struct {
	Time      int64   `json:"time"`      //时间
	Point     string  `json:"point"`     //点位
	Category  int64   `json:"category"`  //点位类别
	Kind      int64   `json:"kind"`      //传感器类别
	Channel   int64   `json:"channel"`   //通道
	Frame     int64   `json:"frame"`     //帧号
	Longitude float64 `json:"longitude"` //经度
	Latitude  float64 `json:"latitude"`  //纬度
	Height    float64 `json:"height"`    //高度
	Value     float64 `json:"value"`     //数值
}

func (t GeoData) ToPoint(name string, tick ...time.Time) *client.Point {
	return InfluxGeoData(name, t.Point, t.Category, t.Kind, t.Channel, t.Frame,
		t.Longitude, t.Latitude, t.Height, t.Value, tick...)
}

func (t GeoData) GeoPoint() GeoPoint {
	return GeoPoint{Longitude: t.Longitude, Latitude: t.Latitude}
}

func (t GeoData) TimeData() TimeGeoData {
	return TimeGeoData{t.Time, t.Longitude, t.Latitude, t.Height, t.Value}
}

func (t GeoData) FrameData() TimeGeoData {
	return TimeGeoData{t.Frame, t.Longitude, t.Latitude, t.Height, t.Value}
}

func (t GeoData) TakeDigits(n int) GeoData {
	res := GeoData{
		Time:      t.Time,
		Point:     t.Point,
		Kind:      t.Kind,
		Channel:   t.Channel,
		Frame:     t.Frame,
		Longitude: TakeDigits(t.Longitude, n),
		Latitude:  TakeDigits(t.Latitude, n),
		Height:    TakeDigits(t.Height, n),
		Value:     TakeDigits(t.Value, n),
	}

	return res
}

func InfluxGeoData(name, point string, category, kind, channel, frame int64,
	longitude, latitude, height, value float64, tick ...time.Time) *client.Point {

	tags := map[string]string{
		"point":    point,
		"category": ParseInt64ToStr(category),
		"kind":     ParseInt64ToStr(kind),
		"channel":  ParseInt64ToStr(channel),
	}

	fields := map[string]interface{}{
		"frame":     frame,
		"longitude": longitude,
		"latitude":  latitude,
		"height":    height,
		"value":     value,
	}

	res := NewInfluxPoint(name, tags, fields, ParseTime(tick))

	return res
}

func NewGeoData(raw []interface{}) GeoData {
	res := GeoData{}
	if len(raw) < 9 {
		return res
	}

	index := 0
	res.Time = ParseTimeFromResp(raw[index]).Unix()

	index++
	res.Category = ParseInt64FromTag(raw[index])

	index++
	res.Channel = ParseInt64FromTag(raw[index])

	index++
	res.Frame = ParseInt64FromField(raw[index])

	index++
	res.Height = ParseFloat64FromField(raw[index])

	index++
	res.Kind = ParseInt64FromTag(raw[index])

	index++
	res.Latitude = ParseFloat64FromField(raw[index])

	index++
	res.Longitude = ParseFloat64FromField(raw[index])

	index++
	res.Point = ParseStrFromTag(raw[index])

	index++
	res.Value = ParseFloat64FromField(raw[index])

	return res
}

type GeoDataSeries []GeoData

func NewGeoDataSeries(raw []client.Result) GeoDataSeries {
	res := GeoDataSeries{}

	rows := MergeInfluxResult(raw)
	if len(rows) == 0 {
		return res
	}

	for _, row := range rows {
		for _, line := range row.Values {
			res = append(res, NewGeoData(line))
		}
	}

	return res
}

func (t GeoDataSeries) InfluxPoint() InfluxPointSeries {
	res := InfluxPointSeries{}

	for _, v := range t {
		res = append(res, v)
	}

	return res
}

func (t GeoDataSeries) FrameData() TimeGeoDataSeries {
	res := TimeGeoDataSeries{}

	for _, v := range t {
		res = append(res, v.FrameData())
	}

	return res
}

func (t GeoDataSeries) FrameDataSet() TimeGeoDataSeries {
	res := TimeGeoDataSeries{}

	m := map[int64]bool{}

	for _, v := range t {
		if m[v.Frame] {
			continue
		}

		res = append(res, v.FrameData())
		m[v.Frame] = true
	}

	sort.Sort(res)

	return res
}

type TimeGeoData struct {
	Time      int64   `json:"time"`      //时间
	Longitude float64 `json:"longitude"` //经度
	Latitude  float64 `json:"latitude"`  //纬度
	Height    float64 `json:"height"`    //高度
	Value     float64 `json:"value"`     //数值
}

func (t TimeGeoData) TakeDigits(n int) TimeGeoData {
	return TimeGeoData{Time: t.Time, Longitude: t.Longitude, Latitude: t.Latitude, Height: t.Height, Value: TakeDigits(t.Value, n)}
}

func (t TimeGeoData) GeoPoint() GeoPoint {
	return GeoPoint{Longitude: t.Longitude, Latitude: t.Latitude}
}

type TimeGeoDataSeries []TimeGeoData
