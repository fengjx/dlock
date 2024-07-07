package dlock

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	//go:embed lua/unlock.lua
	luaUnlock string
)

// RedisLockClient 基于redis实现的分布式锁客户端
type RedisLockClient struct {
	cli  redis.Cmdable
	mtxs map[string]*RedisMutex
}

// NewRedisLock 创建redis锁客户端
func NewRedisLock(cli redis.Cmdable) LockClient {
	c := &RedisLockClient{
		cli:  cli,
		mtxs: make(map[string]*RedisMutex),
	}
	return c
}

// NewMutex 创建redis锁
func (c *RedisLockClient) NewMutex(name string, opts ...Option) Mutex {
	if _, ok := c.mtxs[name]; ok {
		panic("mutex name already exists")
	}
	opt := &Options{
		ttl:        defaultTTL,
		timeout:    defaultTimeout,
		tries:      defaultTries,
		genValueFn: defaultGenValueFn,
	}
	for _, o := range opts {
		o(opt)
	}
	ctx := context.Background()
	unlockSha := c.cli.ScriptLoad(ctx, luaUnlock).Val()
	m := &RedisMutex{
		cli:        c.cli,
		name:       name,
		ttl:        opt.ttl,
		timeout:    opt.timeout,
		tries:      opt.tries,
		genValueFn: opt.genValueFn,
		unlockSha:  unlockSha,
	}
	c.mtxs[name] = m
	return m
}

// RedisMutex redis实现的分布式锁
type RedisMutex struct {
	cli        redis.Cmdable          // redis 客户端
	name       string                 // 锁名称
	ttl        time.Duration          // 持有锁时长
	timeout    time.Duration          // 阻塞获得锁最长等待时长
	genValueFn func() (string, error) // 生成锁 value
	tries      int                    // 获取锁异常时，尝试次数
	val        string                 // 当前锁 value
	unlockSha  string                 // 删除锁脚本sha
}

func (m *RedisMutex) TryLock() error {
	return m.TryLockCtx(context.Background())
}

func (m *RedisMutex) TryLockCtx(ctx context.Context) error {
	return m.tryLockCtx(ctx)
}

func (m *RedisMutex) tryLockCtx(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	var lerr error
	for i := 0; i < m.tries; i++ {
		val, err := m.genValueFn()
		if err != nil {
			return err
		}
		ok, err := m.cli.SetNX(ctx, m.name, val, m.ttl).Result()
		if err != nil {
			lerr = err
			continue
		}
		if !ok {
			return ErrFailed
		}
		m.val = val
		lerr = nil
		return nil
	}
	return lerr
}

func (m *RedisMutex) Lock() error {
	return m.LockCtx(context.Background())
}

func (m *RedisMutex) LockCtx(ctx context.Context) error {
	return m.lockCtx(ctx)
}

func (m *RedisMutex) lockCtx(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()
	var lerr error
	ch := make(chan error, 1)
	go func() {
		for {
			err := m.tryLockCtx(ctx)
			if errors.Is(err, ErrFailed) {
				time.Sleep(10 * time.Millisecond) // Sleep for a while before retrying
				continue
			}
			if err != nil {
				ch <- err
				break
			}
			ch <- nil
		}
	}()
	select {
	case <-ctx.Done():
		lerr = ErrTimeout
	case err := <-ch:
		lerr = err
	}
	return lerr
}

// Unlock 释放锁
// true: 删除成功
// false: 删除失败，表示如果是锁的值已经变更，或者是锁不存在。
// 错不存在的情况会返回 ErrLockNotExistOrExpired
func (m *RedisMutex) Unlock() (bool, error) {
	return m.UnlockCtx(context.Background())
}

// UnlockCtx 释放锁
// true: 删除成功
// false: 删除失败，表示如果是锁的值已经变更，或者是锁不存在。
// 错不存在的情况会返回 ErrLockNotExistOrExpired
func (m *RedisMutex) UnlockCtx(ctx context.Context) (bool, error) {
	return m.release(ctx)
}

func (m *RedisMutex) release(ctx context.Context) (bool, error) {
	ret, err := m.cli.EvalSha(ctx, m.unlockSha, []string{m.name}, m.val).Int64()
	if err != nil {
		return false, err
	}
	if ret == int64(-1) {
		return false, ErrLockNotExistOrExpired
	}
	return ret != int64(0), nil
}
