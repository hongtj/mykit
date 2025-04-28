package dsp

import (
	. "mykit/core/types"
	"reflect"
	"sync"
)

func Json2Map(raw []byte) (s map[string]string, err error) {
	var result map[string]string
	err = UnmarshalJson(raw, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type Map struct {
	v reflect.Type
	m *sync.Map
}

func NewMap(v interface{}) *Map {
	res := &Map{
		v: reflect.TypeOf(v),
		m: new(sync.Map),
	}

	return res
}

func (t *Map) Set(k string, obj interface{}) {
	if reflect.TypeOf(obj) != t.v {
		return
	}

	t.m.Store(k, obj)
}

func (t *Map) Get(k string) (v interface{}, ok bool) {
	v, ok = t.m.Load(k)
	return
}

type ObjMap struct {
	l sync.RWMutex
	m map[string]interface{}
}

func NewObjMap() *ObjMap {
	res := &ObjMap{
		m: map[string]interface{}{},
	}

	return res
}

func (t *ObjMap) Set(k string, obj interface{}) {
	t.l.Lock()
	t.m[k] = obj
	t.l.Unlock()
}

func (t *ObjMap) Get(k string) (v interface{}, ok bool) {
	t.l.RLock()
	v, ok = t.m[k]
	t.l.RUnlock()

	return
}

type BytesMap struct {
	l sync.RWMutex
	m map[string][]byte
}

func NewBytesMap() *BytesMap {
	res := &BytesMap{
		m: map[string][]byte{},
	}

	return res
}

func (t *BytesMap) Set(k string, payload []byte) {
	if k == "" || len(payload) == 0 {
		return
	}

	t.l.Lock()
	t.m[k] = payload
	t.l.Unlock()
}

func (t *BytesMap) Load(k string, obj interface{}) error {
	t.l.RLock()
	defer t.l.RUnlock()

	value, ok := t.m[k]
	if !ok {
		return ErrNotExist
	}

	err := UnmarshalJson(value, &obj)

	return err
}

func (t *BytesMap) Get(k string) (res []byte) {
	t.l.RLock()
	defer t.l.RUnlock()

	res, ok := t.m[k]
	if !ok {
		return ByteOfNullJson
	}

	return
}

func (t *BytesMap) GetList(k string) (res []byte) {
	t.l.RLock()
	defer t.l.RUnlock()

	res, ok := t.m[k]
	if !ok {
		return ByteOfNullList
	}

	return
}

func (t *BytesMap) BatchGetWithSource(s ByteSource, k ...string) (res [][]byte) {
	l := len(k)
	if l == 0 {
		return
	}

	res = make([][]byte, l)

	t.l.RLock()
	defer t.l.RUnlock()

	for i, v := range k {
		b, ok := t.m[v]
		if ok {
			res[i] = b
		} else {
			res[i] = s(v)
		}
	}

	return
}

func (t *BytesMap) BatchGet(k ...string) (res [][]byte) {
	var s ByteSource = func(s string) []byte {
		return ByteOfNullJson
	}

	return t.BatchGetWithSource(s, k...)
}

func (t *BytesMap) BatchGetList(k ...string) (res [][]byte) {
	var s ByteSource = func(s string) []byte {
		return ByteOfNullList
	}

	return t.BatchGetWithSource(s, k...)
}
