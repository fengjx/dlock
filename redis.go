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

// RedisLock redis 锁信息
type RedisLock struct {
	mutex *RedisMutex
	val   string
}

// Unlock 释放锁
func (r RedisLock) Unlock(ctx context.Context) (bool, error) {
	return r.mutex.release(ctx, r.val)
}

// NewMutex 创建redis锁
func (c *RedisLockClient) NewMutex(name string, opts ...Option) Mutex {
	if _, ok := c.mtxs[name]; ok {
		panic("mutex name already exists")
	}
	opt := &Options{
		ttl:         defaultTTL,
		waitTimeout: defaultTimeout,
		tries:       defaultTries,
		genValueFn:  defaultGenValueFn,
	}
	for _, o := range opts {
		o(opt)
	}
	ctx := context.Background()
	unlockSha := c.cli.ScriptLoad(ctx, luaUnlock).Val()
	m := &RedisMutex{
		cli:         c.cli,
		name:        name,
		ttl:         opt.ttl,
		waitTimeout: opt.waitTimeout,
		tries:       opt.tries,
		genValueFn:  opt.genValueFn,
		unlockSha:   unlockSha,
	}
	c.mtxs[name] = m
	return m
}

// RedisMutex redis实现的分布式锁
type RedisMutex struct {
	cli         redis.Cmdable          // redis 客户端
	name        string                 // 锁名称
	ttl         time.Duration          // 持有锁时长
	waitTimeout time.Duration          // 阻塞获得锁最长等待时长
	genValueFn  func() (string, error) // 生成锁 value
	tries       int                    // 获取锁异常时，尝试次数
	unlockSha   string                 // 删除锁脚本sha
}

// TryLock 尝试获取锁，获取不到不会阻塞
func (m *RedisMutex) TryLock(ctx context.Context) (lock Lock, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	for i := 0; i < m.tries; i++ {
		var val string
		val, err = m.genValueFn()
		if err != nil {
			return nil, err
		}
		var ok bool
		ok, err = m.cli.SetNX(ctx, m.name, val, m.ttl).Result()
		if err != nil {
			continue
		}
		if !ok {
			return nil, ErrFailed
		}
		lock = &RedisLock{
			mutex: m,
			val:   val,
		}
		return lock, nil
	}
	return
}

// Lock 获取锁，获取不到会阻塞，最多等待 waitTimeout 时长
func (m *RedisMutex) Lock(ctx context.Context) (lock Lock, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ticker := time.NewTicker(10 * time.Millisecond)
	timer := time.NewTimer(m.waitTimeout)
	for range ticker.C {
		lock, err = m.TryLock(ctx)
		if err == nil {
			return lock, nil
		}
		if !errors.Is(err, ErrFailed) {
			return nil, err
		}
		select {
		case <-timer.C:
			return nil, ErrTimeout
		default:
		}
	}
	return
}

// 释放锁
func (m *RedisMutex) release(ctx context.Context, val string) (bool, error) {
	ret, err := m.cli.EvalSha(ctx, m.unlockSha, []string{m.name}, val).Int64()
	if err != nil {
		return false, err
	}
	if ret == int64(-1) {
		return false, ErrLockNotExistOrExpired
	}
	return ret != int64(0), nil
}
