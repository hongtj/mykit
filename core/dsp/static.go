package dsp

import (
	. "mykit/core/types"
)

type StaticData struct {
	Name   string
	Period string
	Data   map[int64]float64
}

func NewStaticData(period string, n ...string) StaticData {
	res := StaticData{
		Name:   ParseStrParam(n, ""),
		Period: period,
		Data:   map[int64]float64{},
	}

	return res
}

func (t StaticData) K() []int64 {
	res := []int64{}

	for k := range t.Data {
		res = append(res, k)
	}

	return res
}

func (t *StaticData) TakeDigit(decimal int) {
	for k, v := range t.Data {
		t.Data[k] = TakeDigits(v, decimal)
	}
}

func (t *StaticData) GetVal(k int64) interface{} {
	v, ok := t.Data[k]
	if ok {
		return v
	}

	return "-"
}

func (t *StaticData) Get(k int64) float64 {
	return t.Data[k]
}
