package types

import (
	"errors"
	"fmt"
	"strings"
)

type Errors []error

func (t *Errors) Add(err ...error) {
	for _, v := range err {
		*t = append(*t, v)
	}
}

func (t Errors) Error() error {
	for _, v := range t {
		if v != nil {
			return v
		}
	}

	return nil
}

var (
	ErrInvalidParam     = errors.New("invalid param")
	ErrInvalidBodyParam = errors.New("invalid body")
	ErrAlreadyExist     = errors.New("already exist")
	ErrNotExist         = errors.New("not exist")
	ErrNotFound         = errors.New("not found")
	ErrReachLimit       = errors.New("reach limit")
	ErrFailed           = errors.New("failed")
)

func HandleInitErr(msg string, err error, must ...bool) {
	if err == nil {
		return
	}

	if !strings.HasSuffix(msg, "\n") {
		msg = fmt.Sprintf("%v err, %v", msg, err)
	} else {
		msg = fmt.Sprintf("%v err: %v", msg, err)
	}

	if ParseBool(must) {
		panic(msg)
	}
}
