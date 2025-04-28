package transfer

import (
	"time"

	"github.com/levigross/grequests"
)

func ParseRequestOptions(param []grequests.RequestOptions) grequests.RequestOptions {
	if len(param) == 0 {
		return grequests.RequestOptions{
			DialTimeout:    time.Second * 1,
			RequestTimeout: time.Second * 10,
		}
	}

	return param[0]
}

func RequestWithTimeout(d time.Duration) grequests.RequestOptions {
	res := grequests.RequestOptions{
		DialTimeout: d,
		Headers: map[string]string{
			ContentType: JsonContentType,
		},
	}

	return res
}
