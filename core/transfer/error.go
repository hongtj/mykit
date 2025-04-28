package transfer

import (
	"errors"
)

var (
	ErrOffline             = errors.New("offline")
	ErrRpcFailed           = errors.New("rpc failed")
	ErrInvalidTraceParam   = errors.New("invalid trace param")
	ErrInvalidSegmentParam = errors.New("invalid segment param")
	ErrInvalidNsqParam     = errors.New("invalid nsq param")
	ErrInvalidNatsParam    = errors.New("invalid nats param")
)
