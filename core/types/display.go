package types

import (
	"database/sql/driver"
	"fmt"
)

type NamedValue struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

func NewNamedValue(v interface{}, name ...string) NamedValue {
	res := NamedValue{
		Name:  ParseStrParam(name, ""),
		Value: v,
	}

	return res
}

func (t NamedValue) Str() string {
	v, _ := t.Value.(string)
	return v
}

func (t NamedValue) Int32() int32 {
	v, ok := t.Value.(int32)
	if ok {
		return v
	}

	f, ok := t.Value.(float64)
	if ok {
		return int32(f)
	}

	return 0
}

func (t NamedValue) Int64() int64 {
	v, ok := t.Value.(int64)
	if ok {
		return v
	}

	f, ok := t.Value.(float64)
	if ok {
		return int64(f)
	}

	return 0
}

func (t NamedValue) Float64() float64 {
	v, _ := t.Value.(float64)
	return v
}

type NamedValueList []NamedValue

type NamedStr struct {
	K string `json:"name"`
	V string `json:"value"`
}

func (t *NamedStr) Scan(src interface{}) error {
	if src == nil {
		t.V = ""
		return nil
	}

	switch v := src.(type) {
	case []byte:
		t.V = string(v)
	case string:
		t.V = v
	default:
		return fmt.Errorf("cannot convert %T to NamedStr", src)
	}

	return nil
}

func (t NamedStr) Value() (driver.Value, error) {
	return t.V, nil
}

func NewNamedStr(v string, name ...string) NamedStr {
	res := NamedStr{
		K: ParseStrParam(name, ""),
		V: v,
	}

	return res
}

func (t NamedStr) Name() string {
	return t.K
}

func (t NamedStr) Str() string {
	return t.V
}

func (t NamedStr) Diff(raw NamedStr) bool {
	return t.V == raw.V
}

func (t NamedStr) IsEmpty() bool {
	return t.V == ""
}

type NamedStrList []NamedStr

type NamedI32 struct {
	K string `json:"name"`
	V int32  `json:"value"`
}

func NewNamedI32(v int32, name ...string) NamedI32 {
	res := NamedI32{
		K: ParseStrParam(name, ""),
		V: v,
	}

	return res
}

func (t NamedI32) Name() string {
	return t.K
}

func (t NamedI32) Int32() int32 {
	return t.V
}

func (t NamedI32) Diff(raw NamedI32) bool {
	return t.V == raw.V
}

type NamedI32List []NamedI32

type NamedI64 struct {
	K string `json:"name"`
	V int64  `json:"value"`
}

func NewNamedI64(v int64, name ...string) NamedI64 {
	res := NamedI64{
		K: ParseStrParam(name, ""),
		V: v,
	}

	return res
}

func (t NamedI64) Name() string {
	return t.K
}

func (t NamedI64) Int64() int64 {
	return t.V
}

func (t NamedI64) Diff(raw NamedI64) bool {
	return t.V == raw.V
}

type NamedI64List []NamedI64

type NamedF64 struct {
	K string  `json:"name"`
	V float64 `json:"value"`
}

func (t NamedF64) Name() string {
	return t.K
}

func (t NamedF64) Float64() float64 {
	return t.V
}

func (t NamedF64) Diff(raw NamedF64) bool {
	return t.V == raw.V
}

func NewNamedF64(v float64, name ...string) NamedF64 {
	res := NamedF64{
		K: ParseStrParam(name, ""),
		V: v,
	}

	return res
}

type NamedF64List []NamedF64
