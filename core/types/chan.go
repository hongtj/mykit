package types

import (
	"reflect"
)

func CountChanRemain(raw interface{}) int {
	if reflect.TypeOf(raw).Kind() != reflect.Chan {
		return -1
	}

	TearDownWait()

	return reflect.ValueOf(raw).Len()
}
