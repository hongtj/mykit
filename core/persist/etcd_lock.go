package persist

import (
	"context"
	. "mykit/core/types"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

type EtcdLock struct {
	client   *clientv3.Client
	ctx      context.Context
	cancel   context.CancelFunc
	key      string
	revision int64
	ttl      int64
}

func NewEtcdLock(cli *clientv3.Client, key string, ttl ...int64) *EtcdLock {
	res := &EtcdLock{
		client: cli,
		key:    key,
		ttl:    ParseInt64Param(ttl, 10),
	}

	return res
}

func (t *EtcdLock) Lock(ctx context.Context) (err error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		return ErrLockFailed
	}

	// 创建租约
	resp, err := t.client.Grant(ctx, t.ttl)
	if err != nil {
		return err
	}

	t.ctx, t.cancel = context.WithCancel(ctx)
	timeout := deadline.Sub(time.Now())
	go func() {
		for {
			select {
			case <-time.After(timeout / 2):
				_, err = t.client.KeepAliveOnce(t.ctx, resp.ID)
				return
			case <-t.ctx.Done(): // 收到ctx取消信号则退出协程
				return
			}
		}
	}()

	var tx *clientv3.TxnResponse
	cmp := clientv3.Compare(clientv3.CreateRevision(t.key), "=", t.revision)
	put := clientv3.OpPut(t.key, "", clientv3.WithLease(resp.ID))

	for {
		tx, err = t.client.Txn(ctx).
			If(cmp).
			Then(put).
			Else().
			Commit()

		if err == nil && tx.Succeeded {
			t.revision = tx.Header.Revision
			return
		}

		if time.Now().After(deadline) {
			break
		}

		time.Sleep(spinBigInternal)
	}

	return ErrLockTimeout
}

// 释放锁
func (t *EtcdLock) Unlock(ctx context.Context) error {
	t.cancel()

	cmp := clientv3.Compare(clientv3.CreateRevision(t.key), "=", t.revision)
	del := clientv3.OpDelete(t.key)

	_, err := t.client.Txn(ctx).
		If(cmp).
		Then(del).
		Commit()

	return err
}

func (t *EtcdLock) Close() error {
	return t.client.Close()
}

type EtcdLockV2 struct {
	session *concurrency.Session
	mutex   *concurrency.Mutex
}

func NewEtcdLockV2(client *clientv3.Client, key string) (res *EtcdLockV2, err error) {
	s, err := concurrency.NewSession(client)
	if err != nil {
		return nil, err
	}

	res = &EtcdLockV2{
		session: s,
		mutex:   concurrency.NewMutex(s, key),
	}

	return
}

// 加锁
func (l *EtcdLockV2) Lock(ctx context.Context) error {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		err := l.mutex.Lock(ctx)
		if err != nil {
			return
		}
	}()

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		l.mutex.Unlock(ctx)
		<-ch
		return ctx.Err()
	}
}

// 解锁
func (l *EtcdLockV2) Unlock(ctx context.Context) error {
	return l.mutex.Unlock(ctx)
}

// 关闭session
func (l *EtcdLockV2) Close() error {
	return l.session.Close()
}
