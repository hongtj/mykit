package dsp

import (
	"context"
	. "mykit/core/types"
	"sync"
	"sync/atomic"
	"time"
)

type Base struct {
	sync.RWMutex
	tick  time.Time
	batch *uint64
	*ZLogger
}

func (t *Base) BATCH() uint64 {
	return atomic.LoadUint64(t.batch)
}

func (t *Base) Tick() time.Time {
	return t.tick
}

func (t *Base) Cost() time.Duration {
	return time.Now().Sub(t.Tick())
}

func (t *Base) Next(ctx context.Context, tick ...time.Time) {
	t.ZLogger.NewTrace(ctx)

	t.tick = ParseTime(tick)

	atomic.AddUint64(t.batch, 1)

	t.ZLogger.BatchStart(t.BATCH())
}

func NewBase(ctx context.Context, msg string) *Base {
	res := &Base{
		tick:    time.Now(),
		batch:   new(uint64),
		ZLogger: NewZLogger(ctx, msg),
	}

	return res
}
