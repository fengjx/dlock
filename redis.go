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
	cli redis.Cmdable
}

// NewRedisLock 创建redis锁客户端
func NewRedisLock() LockClient {
	c := &RedisLockClient{}
	return c
}

// NewMutex 创建redis锁
func (r RedisLockClient) NewMutex(name string, opts ...Option) Mutex {
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
	unlockSha, _ := r.cli.ScriptLoad(ctx, luaUnlock).Result()
	m := &RedisMutex{
		cli:        r.cli,
		name:       name,
		ttl:        opt.ttl,
		timeout:    opt.timeout,
		tries:      opt.tries,
		genValueFn: opt.genValueFn,
		unlockSha:  unlockSha,
	}
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

func (r RedisMutex) TryLock() error {
	return r.TryLockCtx(context.Background())
}

func (r RedisMutex) TryLockCtx(ctx context.Context) error {
	return r.tryLockCtx(ctx)
}

func (r RedisMutex) tryLockCtx(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	var lerr error
	for i := 0; i < r.tries; i++ {
		val, err := r.genValueFn()
		if err != nil {
			return err
		}
		ok, err := r.cli.SetNX(ctx, r.name, val, r.ttl).Result()
		if err != nil {
			lerr = err
			continue
		}
		if !ok {
			return ErrFailed
		}
		r.val = val
		lerr = nil
	}
	return lerr
}

func (r RedisMutex) Lock() error {
	return r.LockCtx(context.Background())
}

func (r RedisMutex) LockCtx(ctx context.Context) error {
	return r.lockCtx(ctx)
}

func (r RedisMutex) lockCtx(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	var lerr error
	ch := make(chan error, 1)
	go func() {
		for {
			err := r.tryLockCtx(ctx)
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
func (r RedisMutex) Unlock() (bool, error) {
	return r.UnlockCtx(context.Background())
}

// UnlockCtx 释放锁
// true: 删除成功
// false: 删除失败，表示如果是锁的值已经变更，或者是锁不存在。
// 错不存在的情况会返回 ErrLockNotExistOrExpired
func (r RedisMutex) UnlockCtx(ctx context.Context) (bool, error) {
	return r.release(ctx)
}

func (r RedisMutex) release(ctx context.Context) (bool, error) {
	ret, err := r.cli.EvalSha(ctx, r.unlockSha, []string{r.name}, r.val).Int64()
	if err != nil {
		return false, err
	}
	if ret == int64(-1) {
		return false, ErrLockNotExistOrExpired
	}
	return ret != int64(0), nil
}
