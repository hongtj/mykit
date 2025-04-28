package persist

import "errors"

var (
	ErrLockFailed  = errors.New("lock failed")
	ErrLockTimeout = errors.New("lock timeout")
)
