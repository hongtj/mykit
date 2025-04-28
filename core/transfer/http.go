package transfer

import (
	"bytes"
	"encoding/json"
	"fmt"
	. "mykit/core/types"
	"sync/atomic"

	"github.com/levigross/grequests"
)

var (
	httpDebug = new(int32)
)

func HttpDebug(n ...int32) {
	atomic.StoreInt32(httpDebug, ParseInt32Param(n, 1))
}

func IsHttpDebug() bool {
	return atomic.LoadInt32(httpDebug) == 1
}

func JsonPost(msg, url string, req interface{}, o ...grequests.RequestOptions) (body []byte, err error) {
	payload, err := json.Marshal(req)
	if err != nil {
		fmt.Println(msg+" post req", err)
		return
	}

	if IsHttpDebug() {
		fmt.Println(url)
		fmt.Println(string(payload))
	}

	option := ParseRequestOptions(o)

	option.RequestBody = bytes.NewReader(payload)

	resp, err := grequests.Post(url, &option)
	if err != nil {
		fmt.Println(msg, err)
		return
	}

	body = resp.Bytes()
	if IsHttpDebug() {
		fmt.Println(resp.RawResponse)
	}

	return
}

func PostJson(msg, url string, req, rsp interface{}, o ...grequests.RequestOptions) (err error) {
	resp, err := JsonPost(msg, url, req, o...)
	if err != nil {
		return
	}

	err = json.Unmarshal(resp, &rsp)

	return
}

func HttpGet(msg, url string, o ...grequests.RequestOptions) (body []byte, err error) {
	option := ParseRequestOptions(o)

	resp, err := grequests.Get(url, &option)
	if err != nil {
		fmt.Println(msg+" get req", err)
		return
	}

	if IsHttpDebug() {
		fmt.Println(url)
	}

	body = resp.Bytes()
	if IsHttpDebug() {
		fmt.Println(resp.RawResponse)
	}

	return
}

func JsonCall(url string, req interface{}) (rsp JsonCallRes, err error) {
	resp, err := JsonPost("JsonCall", url, req)
	if err != nil {
		return JsonCallRes{}, err
	}

	err = json.Unmarshal(resp, &rsp)

	return
}
