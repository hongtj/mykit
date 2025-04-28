package dsp

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	. "mykit/core/types"
	"net/url"
	"reflect"
	"time"
)

// FNV哈希算法

func Fnv32(key string) int {
	hash := uint32(2166136261)
	l := len(key)
	for i := 0; i < l; i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}

	return int(hash)
}

func EnsureJsonStr(raw string) string {
	return DeStrParam(raw, StrOfNullJson)
}

func EnsureJsonStrFromByte(raw []byte) string {
	return BytesToString(EnsureJsonByte(raw))
}

func EnsureJsonByte(raw []byte) []byte {
	if len(raw) == 0 {
		return ByteOfNullJson
	}

	if len(raw) == 4 &&
		raw[0] == ByteOfNull[0] && raw[1] == ByteOfNull[1] &&
		raw[2] == ByteOfNull[2] && raw[3] == ByteOfNull[3] {
		return ByteOfNullJson
	}

	return raw
}

func EnsureJsonRawMessage(raw string) json.RawMessage {
	return EnsureJsonByte(StringToBytes(raw))
}

func EnsureListByteFromStr(raw string) []byte {
	return DeByteListParam(StringToBytes(raw), ByteOfNullList)
}

func EnsureListByte(raw []byte) []byte {
	return DeByteListParam(raw, ByteOfNullList)
}

func EnsureListJsonRawMessage(raw json.RawMessage) json.RawMessage {
	return DeByteListParam(raw, RawMessageOfNullList)
}

func RespJsonMarshal(obj interface{}) []byte {
	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		return RespJsonMarshalValue(reflect.ValueOf(obj))
	}

	return mustJsonMarshal(1, obj)
}

func RespJsonMarshalValue(value reflect.Value) []byte {
	if value.IsNil() {
		return ByteOfNullJson
	}

	return mustJsonMarshal(1, value.Interface())
}

func MustJsonMarshal(obj interface{}) []byte {
	return mustJsonMarshal(1, obj)
}

func JsonRawMessage(obj interface{}) json.RawMessage {
	if obj == nil {
		return RawMessageOfNullJson
	}

	return mustJsonMarshal(1, obj)
}

func mustJsonMarshal(k int, obj interface{}) []byte {
	k++

	res, err := json.Marshal(obj)
	if err != nil {
		LogS(k).Error("mustJsonMarshal",
			LogError(err),
		)

		return nil
	}

	return res
}

func ToJsonStr(obj interface{}) string {
	str, ok := obj.(string)
	if ok {
		return str
	}

	return BytesToString(mustJsonMarshal(1, obj))
}

func UnmarshalJson(raw []byte, v interface{}) error {
	err := json.Unmarshal(raw, &v)
	if err != nil {
		LogS1.Error("UnmarshalJson",
			LogError(err),
		)
	}

	return err
}

func UnmarshalJsonStr(raw string, v interface{}) error {
	err := json.Unmarshal(StringToBytes(raw), &v)
	if err != nil {
		LogS1.Error("UnmarshalJsonStr",
			LogError(err),
		)
	}

	return err
}

func unmarshalJson(k int, raw []byte, v interface{}) error {
	k++

	err := json.Unmarshal(raw, &v)
	if err != nil {
		LogS(k).Error("unmarshalJson",
			LogError(err),
		)
	}

	return err
}

func TransToJrw(raw string, obj interface{}) json.RawMessage {
	err := unmarshalJson(1, StringToBytes(raw), &obj)
	if err != nil {
		return ByteOfNullJson
	}

	return mustJsonMarshal(1, obj)
}

func TransToJrwList(raw string, obj interface{}) json.RawMessage {
	err := unmarshalJson(1, StringToBytes(raw), &obj)
	if err != nil {
		return ByteOfNullList
	}

	return mustJsonMarshal(1, obj)
}

func PrettyJson(raw interface{}) string {
	s := mustJsonMarshal(1, raw)

	var d bytes.Buffer
	err := json.Indent(&d, s, "", "  ")
	if err != nil {
		return ""
	}

	return d.String()
}

func UrlEncode(s string) string {
	return url.QueryEscape(s)
}

func ParseObjToBuffer(obj interface{}) (buf bytes.Buffer, err error) {
	err = json.NewEncoder(&buf).Encode(obj)

	return
}

func ParseObjToReader(obj interface{}) io.Reader {
	var p []byte
	if b, ok := obj.([]byte); ok {
		p = b

	} else if s, ok2 := obj.(string); ok2 {
		p = StringToBytes(s)

	} else {
		p = mustJsonMarshal(1, obj)
	}

	return bytes.NewBuffer(p)
}

func ParseObjToReaderCloser(obj interface{}) io.ReadCloser {
	return ioutil.NopCloser(ParseObjToReader(obj))
}

func ZipObj(raw interface{}) []byte {
	var p []byte
	if b, ok := raw.([]byte); ok {
		p = b

	} else if s, ok2 := raw.(string); ok2 {
		p = StringToBytes(s)

	} else {
		p = mustJsonMarshal(1, raw)
	}

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(p)
	w.Close()

	res := b.Bytes()

	return res
}

func UnzipObj(raw []byte, obj interface{}) (err error) {
	return unzipObj(1, raw, obj)
}

func UnzipFromByte(raw []byte) (res []byte, err error) {
	r, err := zlib.NewReader(bytes.NewBuffer(raw))
	if err != nil {
		return
	}

	err = r.Close()
	if err != nil {
		return
	}

	var b bytes.Buffer
	_, err = io.Copy(&b, r)
	if err != nil {
		return
	}

	res = b.Bytes()

	return
}

func unzipObj(k int, raw []byte, obj interface{}) (err error) {
	k++

	data, err := UnzipFromByte(raw)
	if err != nil {
		return err
	}

	err = unmarshalJson(k, data, &obj)

	return
}

func ZipObjToBase64(raw interface{}) string {
	p := ZipObj(raw)
	res := base64.StdEncoding.EncodeToString(p)

	return res
}

func UnzipObjFromBase64(raw string, obj interface{}) (err error) {
	p, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return err
	}

	return unzipObj(1, p, obj)
}

func UnzipFromStr(raw string) (res []byte, err error) {
	return UnzipFromByte(StringToBytes(raw))
}

func UnzipFromBase64(raw string) (res []byte, err error) {
	p, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return
	}

	return UnzipFromByte(p)
}

func UnzipList(raw []byte) (res []byte) {
	res = ByteOfNullList

	data, err := UnzipFromByte(raw)
	if err == nil {
		res = data
	}

	return
}

func ParseTimeFromResp(raw interface{}) time.Time {
	if raw == nil {
		return time.Time{}
	}

	t, ok := raw.(string)
	if ok {
		return MustParseTimeStr(t)
	}

	n, ok := raw.(json.Number)
	if ok {
		tick, err := n.Int64()
		if err == nil {
			return time.Unix(tick, 0)
		}
	}

	return time.Time{}
}
