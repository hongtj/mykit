package dsp

import (
	"errors"
	. "mykit/core/types"
	"strings"
)

func InvalidParamErr(err error) bool {
	return errors.Is(err, ErrInvalidParam)
}

func AlreadyExistErr(err error) bool {
	return errors.Is(err, ErrAlreadyExist)
}

func NotExistErr(err error) bool {
	return errors.Is(err, ErrNotExist)
}

func NotFoundErr(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func ReachLimitErr(err error) bool {
	return errors.Is(err, ErrReachLimit)
}

func TimeoutErr(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "Client.Timeout")
}

type CodeResp struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

func (t CodeResp) Err() error {
	if t.Code == 0 {
		return nil
	}

	return NewErrorCode(t.Msg, -1*t.Code)
}

type ErrorCode struct {
	code int32
	s    string
}

func NewErrorCode(text string, code ...int32) *ErrorCode {
	return &ErrorCode{code: ParseInt32Param(code, -1), s: text}
}

func NewError(text string) error {
	return NewErrorCode(text)
}

func SuccessCode(text ...string) error {
	return &ErrorCode{code: RspCodeSuccess, s: ParseStrParam(text, RspMsgSuccess)}
}

func SuccessWithCode(code int32) error {
	return &ErrorCode{code: code, s: RspMsgSuccess}
}

func MsgCode(text string) error {
	return &ErrorCode{code: RspCodeMsg, s: text}
}

func (t *ErrorCode) Code() int32 {
	return t.code
}

func (t *ErrorCode) Error() string {
	return t.s
}

var (
	ErrorCodeInvalidParam     = NewErrorCode("invalid param")
	ErrorCodeInvalidBodyParam = NewErrorCode("invalid body")
	ErrorCodeNotExist         = NewErrorCode("not exist")
	ErrorCodeNotFound         = NewErrorCode("not found")
	ErrorCodeReachLimit       = NewErrorCode("reach limit")
	ErrorCodeNoPermission     = NewErrorCode("no permission")
	ErrorCodeTimeout          = NewErrorCode("timeout", CodeDeadlineExceeded)
)
