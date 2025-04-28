package dsp

import (
	"encoding/json"
	"fmt"
	"sync"
)

type SimpleSignalStatus struct {
	Name   string
	Status bool
	T0     int64
	T1     int64
}

func NewSignalStatus(name string) SignalStatus {
	res := &SimpleSignalStatus{
		Name: name,
	}

	return res
}

func (t *SimpleSignalStatus) New(name string, signal SignalStatus) SignalStatus {
	t0, t1, status := signal.Signal()

	res := &SimpleSignalStatus{
		Name:   name,
		Status: status,
		T0:     t0,
		T1:     t1,
	}

	return res
}

func (t *SimpleSignalStatus) On(tick int64) {
	t.Status = true
	t.T0 = tick
	t.T1 = tick
}

func (t *SimpleSignalStatus) Off(tick int64) {
	t.Status = false
	t.T1 = tick
}

func (t *SimpleSignalStatus) Open() int64 {
	return t.T0
}

func (t *SimpleSignalStatus) Close() int64 {
	return t.T1
}

func (t *SimpleSignalStatus) Update(status bool, tick int64) {
	if tick < t.T0 {
		return
	}

	if t.Status {
		if !status {
			t.Off(tick)
			Debugs(t.Name, "down @", t.T1)

		} else {
			t.T1 = tick
		}

		return
	}

	if status {
		t.On(tick)
		Debugs(t.Name, "up @", t.T0)
	}
}

func (t *SimpleSignalStatus) Duration() int64 {
	return t.T1 - t.T0
}

func (t *SimpleSignalStatus) Signal() (int64, int64, bool) {
	return t.T0, t.T1, t.Status
}

func (t *SimpleSignalStatus) String() string {
	if t.Status {
		return fmt.Sprintf("%v is running, %v-%v", t.Name, t.T0, t.T1)
	}

	return fmt.Sprintf("%v is down, %v-%v", t.Name, t.T0, t.T1)
}

type SignalStatusMgr struct {
	l      sync.Mutex
	all    []string
	signal map[string]SignalStatus
	cache  SignalCache
	f      func(string) SignalStatus
}

func NewSignalStatusMgr() *SignalStatusMgr {
	res := &SignalStatusMgr{
		signal: map[string]SignalStatus{},
		f:      NewSignalStatus,
	}

	return res
}

func (t *SignalStatusMgr) Init(cache SignalCache) {
	t.cache = cache
	raw := t.cache.Get("all")
	json.Unmarshal(raw, &t.all)

	for _, v := range t.all {
		signal := t.cache.Load(v)

		t.signal[v] = t.f(v).New(v, signal)
	}
}

func (t *SignalStatusMgr) Set(f func(string) SignalStatus) {
	t.l.Lock()
	t.f = f
	t.l.Unlock()
}

func (t *SignalStatusMgr) Load(name string) SignalStatus {
	t.l.Lock()
	defer t.l.Unlock()

	signal, ok := t.signal[name]
	if ok {
		return signal
	}

	t.all = append(t.all, name)

	signal = t.f(name)
	t.signal[name] = signal

	if t.cache != nil {
		t.cache.Put("all", MustJsonMarshal(t.all))
	}

	return signal
}

func (t *SignalStatusMgr) Commit(name string, status bool, tick int64) {
	s := t.Load(name)
	s.Update(status, tick)

	if t.cache != nil {
		t.cache.Put(name, MustJsonMarshal(s))
	}
}
