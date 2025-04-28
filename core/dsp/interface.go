package dsp

import (
	"context"
	. "mykit/core/types"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

type G struct {
	Plat []string
	User int
	Mac  string
}

type CORE struct {
	*ZLogger
}

type NormalObject struct {
}

type DataObject struct {
	Data interface{} `json:"data"`
}

func NewListData() *DataObject {
	res := &DataObject{
		Data: []int{},
	}

	return res
}

type StrData struct {
	Data string `json:"data"`
}

type BaseQuery struct {
	Id   int64  `json:"id"`
	Uuid string `json:"uuid"`
}

type BatchQuery struct {
	Id   []int64  `json:"id"`
	Uuid []string `json:"uuid"`
}

type BatchCodeQuery struct {
	Code []string `json:"code"`
}

type ACCESS struct {
	User string
	Pwd  string
	Host string
	Port int32
}

type AccessDb struct {
	ACCESS
	Db string
}

type RedisACCESS struct {
	Uri []string
	Pwd string
}

type StatusChecker func() int32

type Marker func(int64) float64

type Trigger func(int64, float64) (float64, bool)

type AlertTrigger func(*AlertLine) Trigger

type AlertSystemOption func(*AlertSystem)

type PointDataSeriesSource func(table string) PointDataSeries

type SensorDataSeriesSource func(table string) SensorDataSeries

type SensorDataSeriesFrom func(table string, tick int64) SensorDataSeries

type SensorDataSeriesFrameMap map[time.Time]SensorDataSeries

type DpcParser func(SensorData) float64

type TimeDataSeriesSource func(table string) TimeDataSeries

type TimeDataSeriesMap map[int64]TimeDataSeries

type AlertMark struct {
	Tick  int64
	Value float64
	Level int
	Scale float64
	Delta float64
}

func (t AlertMark) Alert() bool {
	return t.Level != 0
}

func (t AlertMark) TimeData() TimeData {
	return TimeData{Time: t.Tick, Data: t.Value}
}

type AlertMarkList []AlertMark

type AlertMarker interface {
	Mark(tick int64, value float64) AlertMark
}

type InfluxPoint interface {
	ToPoint(name string, tick ...time.Time) *client.Point
}

type InfluxPointSeries []InfluxPoint

type InfluxPointJob chan InfluxPointSeries

func CountInfluxPointRemain(job InfluxPointJob) int {
	TearDownWait()
	return len(job)
}

type StrDataHandle func(string)

type StrDataJob chan string

func CountStrDataRemain(job StrDataJob) int {
	TearDownWait()
	return len(job)
}

func (t StrDataJob) Commit(v string) {
	var job = func(ctx context.Context) {
		t <- v
	}

	Go(context.Background(), job, "commit str data")
}

type IotDataPack map[string]float64

type IotDataPackSeries []IotDataPack

type IotDataPackHandle func(IotDataPack)

type IotDataPackJob chan IotDataPack

func CountIotDataPackRemain(job IotDataPackJob) int {
	TearDownWait()
	return len(job)
}

func (t IotDataPackJob) Commit(v IotDataPack) {
	var job = func(ctx context.Context) {
		t <- v
	}

	Go(context.Background(), job, "commit iot data pack")
}

type SignalStatus interface {
	New(name string, signal SignalStatus) SignalStatus
	Update(bool, int64)
	Signal() (int64, int64, bool)
	Duration() int64
}

type SignalCache interface {
	KVCache
	Load(string) SignalStatus
}

type SensorDataHandle func(...SensorData)

type SensorDataJob chan SensorDataSeries

func CountSensorDataRemain(job SensorDataJob) int {
	TearDownWait()
	return len(job)
}

func (t SensorDataJob) Commit(v ...SensorData) {
	var job = func(ctx context.Context) {
		t <- v
	}

	Go(context.Background(), job, "commit sensor data")
}

type GeoDataConsumer func(GeoDataJob)

type GeoDataJob chan GeoDataSeries

func CountGeoDataJobRemain(job GeoDataJob) int {
	TearDownWait()
	return len(job)
}

func (t GeoDataJob) Commit(v ...GeoData) {
	var job = func(ctx context.Context) {
		t <- v
	}

	Go(context.Background(), job, "commit GeoData")
}

type TenantDbParser func(ctx context.Context) (tenant, db string)

type TenantRedisPrefix func(ctx context.Context) string

type TenantRedisKeyParser func(ctx context.Context, key string) string

type MessagePusher interface {
	Send(ctx context.Context, msg interface{}) error
}

type HashObject interface {
	Key(context.Context) string
	Values() []interface{}
}

type Encoder interface {
	Valid([]byte) bool
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
	Reg(string, interface{}) error
	Load(string, []byte) (interface{}, error)
}

type KVCache interface {
	Put(k string, v []byte)
	Get(k string) []byte
}

type WatchCat int64

type BaseMeta struct {
	Id          string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"` //描述
	Status      int32  `json:"status"`      //状态
	Sort        int64  `json:"sort"`        //排序
}

type ThirdServerControlConf struct {
	CloseSendSms       bool `json:",default=false"` //是否关闭短信发送
	CloseSendWx        bool `json:",default=false"` //是否关闭企业微信发送
	CloseSendBroadcast bool `json:",default=false"` //是否关闭广播
}

type OssFile struct {
	Uuid     string `json:"uuid"`
	Domain   string `json:"domain"`
	Filename string `json:"filename"`
	Mtime    int64  `json:"mtime"`
	Path     string `json:"path"`
	Scenes   string `json:"scenes"`
	Size     int64  `json:"size"`
	URL      string `json:"url"`
}

type Bgd struct {
	Company  int64 `db:"company" json:"company"`   // 公司
	Workshop int64 `db:"workshop" json:"workshop"` // 车间
	Scx      int64 `db:"scx" json:"scx"`           // 生产线
	Team     int64 `db:"team" json:"team"`         // 班组
	Operator int64 `db:"operator" json:"operator"` // 操作员
}
