package dsp

import (
	"fmt"
	. "mykit/core/types"
	"sort"
	"time"
)

var (
	globalAlertOutput bool
)

func SetAlertOutput(raw ...bool) {
	globalAlertOutput = ParseBool(raw)
}

// 警戒线
type AlertLine struct {
	Name    string  // 警戒线名称
	Level   int     // 警戒线级别
	Marker  Marker  // 警戒线阈值
	Trigger Trigger // 是否触发
}

type AlertLineOption struct {
	Name   string      // 警戒线名称
	Level  int         // 警戒线级别
	Marker interface{} // 警戒线阈值
}

func MarkWrap(raw float64) Marker {
	var res Marker = func(int64) float64 {
		return raw
	}

	return res
}

func ParseMarker(raw interface{}) Marker {
	v, ok := raw.(float64)
	if ok {
		return MarkWrap(v)
	}

	n, ok := raw.(int)
	if ok {
		return MarkWrap(float64(n))
	}

	return raw.(Marker)
}

func TimeMarkerSlope(v1, v2 TimeMarker) float64 {
	return ComputeSlope(v1.Delta, v1.Value, v2.Delta, v2.Value)
}

func NewMarkerFromTimeMarker(start int64, raw ...TimeMarker) Marker {
	l := len(raw) - 1
	if l == 0 {
		return MarkWrap(raw[0].Value)
	}

	sort.Slice(raw, func(i, j int) bool {
		return raw[i].Delta < raw[j].Delta
	})

	var res Marker = func(tick int64) float64 {
		delta := float64(tick - start)
		if delta < raw[0].Delta {
			return raw[0].Value
		}

		// 根据delta选择合适的两个TimeMarker，线性拟合对应的Value
		for i := 0; i < l; i++ {
			if delta < raw[i+1].Delta {
				slope := TimeMarkerSlope(raw[i], raw[i+1])
				return raw[i].Value + slope*(delta-raw[i].Delta)
			}
		}

		// 如果delta>=最大的Delta，则用最大的TimeMarker的Value
		return raw[l].Value
	}

	return res
}

func UpperTrigger(raw *AlertLine) Trigger {
	var res Trigger = func(tick int64, value float64) (float64, bool) {
		mark := raw.Marker(tick)
		delta := value - mark
		return delta, delta >= 0
	}

	return res
}

func LowerTrigger(raw *AlertLine) Trigger {
	var res Trigger = func(tick int64, value float64) (float64, bool) {
		mark := raw.Marker(tick)
		delta := value - mark
		return delta, delta <= 0
	}

	return res
}

// 报警系统
type AlertSystem struct {
	Name      string
	AlertLine []*AlertLine // 警戒线列表
	running   bool
	start     time.Time
	msg       string
	align     bool
	output    bool
}

func NewAlertSystem(name string, opt ...AlertSystemOption) *AlertSystem {
	res := &AlertSystem{
		Name:      name,
		AlertLine: []*AlertLine{},
		output:    globalAlertOutput,
	}

	return res.Apply(opt...)
}

func AlertSystemOutput(raw ...bool) AlertSystemOption {
	var res AlertSystemOption = func(system *AlertSystem) {
		system.output = ParseBool(raw)
	}

	return res
}

func AlertSystemAlign(raw ...bool) AlertSystemOption {
	var res AlertSystemOption = func(system *AlertSystem) {
		system.align = ParseBool(raw)
	}

	return res
}

func (t *AlertSystem) Apply(opt ...AlertSystemOption) *AlertSystem {
	for _, v := range opt {
		v(t)
	}

	return t
}

func (t *AlertSystem) AddAlertLine(alert ...*AlertLine) *AlertSystem {
	for _, v := range alert {
		t.AlertLine = append(t.AlertLine, v)
	}

	return t
}

func (t *AlertSystem) AddAlert(trigger AlertTrigger, alert ...AlertLineOption) *AlertSystem {
	if t.running {
		return t
	}

	for _, v := range alert {
		item := &AlertLine{
			Name:   v.Name,
			Level:  v.Level,
			Marker: ParseMarker(v.Marker),
		}

		item.Trigger = trigger(item)

		t.AlertLine = append(t.AlertLine, item)
	}

	return t
}

func (t *AlertSystem) AddUpper(alert ...AlertLineOption) *AlertSystem {
	return t.AddAlert(UpperTrigger, alert...)
}

func (t *AlertSystem) AddLower(alert ...AlertLineOption) *AlertSystem {
	return t.AddAlert(LowerTrigger, alert...)
}

func (t *AlertSystem) SetMarker(alert ...AlertLineOption) *AlertSystem {
	m := map[string]Marker{}

	for _, v := range alert {
		m[v.Name] = ParseMarker(v.Marker)
	}

	for i := range t.AlertLine {
		line := t.AlertLine[i]
		marker, ok := m[line.Name]
		if ok {
			line.Marker = marker
		}
	}

	return t
}

func (t *AlertSystem) Run() {
	if t.running {
		return
	}

	sort.Slice(t.AlertLine, func(i, j int) bool {
		return t.AlertLine[i].Level < t.AlertLine[j].Level
	})

	if t.output {
		fmt.Println(t.Name)
		for _, v := range t.AlertLine {
			fmt.Printf("level %v, %v警戒线, mark %v\n",
				v.Level, v.Name, v.Marker(time.Now().Unix()))
		}

		fmt.Println("\n")
	}

	t.running = true
}

func (t *AlertSystem) Trigger(tick int64, value float64) (alert int, delta float64) {
	t.Run()

	var name string
	d1 := []float64{}
	n := 0

	for i, line := range t.AlertLine {
		lineDelta, ok := line.Trigger(tick, value)
		d1 = append(d1, lineDelta)

		if !ok {
			continue
		}

		if line.Level > alert {
			alert = line.Level
			name = line.Name
			n = i
		}
	}

	delta = d1[n]

	if t.output {
		if t.start.IsZero() {
			t.start = time.Unix(tick, 0)
		}

		eventTime := time.Unix(tick, 0)
		if alert == 0 {
			t.msg = fmt.Sprintf("%v, value [%v] 未触发报警 | delta %v",
				eventTime.Sub(t.start), value, delta)
		} else {
			t.msg = fmt.Sprintf("%v, value [%v] 触发：level %v, %v警戒线 | delta %v",
				eventTime.Sub(t.start), value, alert, name, delta)
		}

		fmt.Println(t.msg)
	}

	return
}
