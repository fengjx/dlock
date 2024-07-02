package dlock

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
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
	m := &RedisMutex{
		name: name,
	}
	for _, o := range opts {
		o.Apply(m)
	}
	return m
}

// RedisMutex redis实现的分布式锁
type RedisMutex struct {
	name       string
	ttl        time.Duration
	timeout    time.Duration
	genValueFn func() (string, error)

	tries int
}

func (r RedisMutex) Lock() error {
	return r.LockCtx(context.Background())
}

func (r RedisMutex) LockCtx(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (r RedisMutex) TryLock() error {
	return r.TryLockCtx(context.Background())
}

func (r RedisMutex) TryLockCtx(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (r RedisMutex) Unlock() (bool, error) {
	return r.UnlockCtx(context.Background())
}

func (r RedisMutex) UnlockCtx(ctx context.Context) (bool, error) {
	//TODO implement me
	panic("implement me")
}

// WithRedisGenValueFunc 自定义 value 生成
func WithRedisGenValueFunc(genValueFn func() (string, error)) Option {
	return OptionFunc(func(m Mutex) {
		if mtx, ok := m.(*RedisMutex); ok {
			mtx.genValueFn = genValueFn
		}
		panic("option only supported for RedisMutex")
	})
}

// WithRedisTTL 设置锁过期时间
func WithRedisTTL(ttl time.Duration) Option {
	return OptionFunc(func(m Mutex) {
		if mtx, ok := m.(*RedisMutex); ok {
			mtx.ttl = ttl
		}
		panic("option only supported for RedisMutex")
	})
}

// WithRedisTimeout 设置获取锁超时时间
func WithRedisTimeout(timeout time.Duration) Option {
	return OptionFunc(func(m Mutex) {
		if mtx, ok := m.(*RedisMutex); ok {
			mtx.timeout = timeout
		}
		panic("option only supported for RedisMutex")
	})
}

// WithRedisTries 获取锁失败重试次数
func WithRedisTries(tries int) Option {
	return OptionFunc(func(m Mutex) {
		if mtx, ok := m.(*RedisMutex); ok {
			mtx.tries = tries
		}
		panic("option only supported for RedisMutex")
	})
}
