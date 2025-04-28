package types

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

// interfaceToSliceInterface convert interface to slice interface
func InterfaceToSliceInterface(docs interface{}) []interface{} {
	kt := reflect.TypeOf(docs)
	kt = DeType(kt)

	if kt.Kind() != reflect.Slice {
		return nil
	}

	s := reflect.ValueOf(docs)
	s = DeValue(s)
	if s.Len() == 0 {
		return nil
	}

	var sDocs []interface{}
	for i := 0; i < s.Len(); i++ {
		sDocs = append(sDocs, s.Index(i).Interface())
	}

	return sDocs
}

func ExpandStructFields(raw interface{}, tag string, res map[string]interface{}) {
	t := GetTypeOfStruct(raw)
	v := GetValueOfStruct(raw)

	n := v.NumField()
	for i := 0; i < n; i++ {
		obj := v.Field(i).Interface()
		if obj == nil {
			continue
		}

		f := t.Field(i)
		k := f.Tag.Get(tag)
		if k == "-" {
			continue
		}

		if strings.Contains(k, TagDelimiter) {
			k = strings.Split(k, TagDelimiter)[0]
		}

		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			ExpandStructFields(obj, tag, res)
		} else {
			res[k] = obj
		}
	}
}

func TransStructToMapStrInterfaceByTag(raw interface{}, tag string) map[string]interface{} {
	if tag == "" {
		panic("need tag")
	}

	res := map[string]interface{}{}
	ExpandStructFields(raw, tag, res)

	return res
}

func DeType(raw reflect.Type) reflect.Type {
	for raw.Kind() == reflect.Ptr {
		raw = raw.Elem()
	}

	return raw
}

func DeValue(raw reflect.Value) reflect.Value {
	for raw.Kind() == reflect.Ptr {
		raw = raw.Elem()
	}

	return raw
}

func GetTypeOfStruct(raw interface{}) reflect.Type {
	rt := reflect.TypeOf(raw)
	rt = DeType(rt)

	if rt.Kind() != reflect.Struct {
		panic("gto need struct")
	}

	return rt
}

func GetValueOfStruct(raw interface{}) reflect.Value {
	rv := reflect.ValueOf(raw)
	rv = DeValue(rv)

	if rv.Kind() != reflect.Struct {
		panic("gvo need struct")
	}

	return rv
}

func SetField(obj interface{}, name string, value interface{}) error {
	structValue := reflect.ValueOf(obj).Elem()
	structFieldValue := structValue.FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return errors.New("Provided value type didn't match obj field type")
	}

	structFieldValue.Set(val)
	return nil
}

func SetFiledToStruct(raw interface{}, f interface{}) {
	filedName := GetTypeOfStruct(f).Name()
	s := GetValueOfStruct(raw)
	t := s.FieldByName(filedName)
	t.Set(reflect.ValueOf(f))
}

func IsNil(i interface{}) bool {
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}

	return false
}

func IsPtr(raw interface{}) bool {
	return ValidObject(raw, reflect.Ptr)
}

func IsStruct(raw interface{}) bool {
	return ValidObject(raw, reflect.Struct)
}

func IsSlice(raw interface{}) bool {
	return ValidObject(raw, reflect.Slice)
}

func IsMap(raw interface{}) bool {
	return ValidObject(raw, reflect.Map)
}

func ValidObject(raw interface{}, t reflect.Kind) bool {
	return reflect.TypeOf(raw).Kind() == t
}

func IsSliceElem(raw interface{}) bool {
	return ValidElem(raw, reflect.Slice)
}

func ValidElem(raw interface{}, t reflect.Kind) bool {
	return reflect.TypeOf(raw).Elem().Kind() == t
}

func IsStructPtr(raw interface{}) bool {
	t, ok := raw.(reflect.Type)
	if ok {
		return IsStructPtrKind(t)
	}

	return IsStructPtrKind(reflect.TypeOf(raw))
}

func IsSlicePtr(raw interface{}) bool {
	t, ok := raw.(reflect.Type)
	if ok {
		return IsSlicePtrKind(t)
	}

	return IsSlicePtrKind(reflect.TypeOf(raw))
}

func IsStructPtrKind(raw reflect.Type) bool {
	if raw.Kind() != reflect.Ptr {
		return false
	}

	return raw.Elem().Kind() == reflect.Struct
}

func IsSlicePtrKind(raw reflect.Type) bool {
	if raw.Kind() != reflect.Ptr {
		return false
	}

	return raw.Elem().Kind() == reflect.Slice
}

func IsContextKind(raw reflect.Type) bool {
	if raw.Kind() == CtxKind {
		return true
	}

	return reflect.TypeOf(raw).Kind() == CtxKind
}

func IsErrorKind(raw reflect.Type) bool {
	if raw.Kind() == ErrKind {
		return true
	}

	return reflect.TypeOf(raw).Kind() == ErrKind
}

func SumStructInt32Filed(raw interface{}) int32 {
	var sum int32

	v := GetValueOfStruct(raw)
	n := v.NumField()
	for i := 0; i < n; i++ {
		s, ok := v.Field(i).Interface().(int32)
		if ok {
			sum += s
		}
	}

	return sum
}

func SumStructInt64Filed(raw interface{}) int64 {
	var sum int64

	v := GetValueOfStruct(raw)
	n := v.NumField()
	for i := 0; i < n; i++ {
		s, ok := v.Field(i).Interface().(int64)
		if ok {
			sum += s
		}
	}

	return sum
}

func SumStructFloat64Filed(raw interface{}) float64 {
	var sum float64

	v := GetValueOfStruct(raw)
	n := v.NumField()
	for i := 0; i < n; i++ {
		s, ok := v.Field(i).Interface().(float64)
		if ok {
			sum += s
		}
	}

	return sum
}

func Sizeof(v interface{}) uintptr {
	return unsafe.Sizeof(v)
}
