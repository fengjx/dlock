package dlock

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	defaultTTL     = 1 * time.Minute
	defaultTimeout = 10 * time.Second
	defaultTries   = 3
)

// GenValueFn lock value 生成器
type GenValueFn func() (string, error)

var defaultGenValueFn GenValueFn = func() (string, error) {
	return uuid.NewString(), nil
}

// LockClient 锁客户端
type LockClient interface {
	// NewMutex 创建锁
	NewMutex(name string, opts ...Option) Mutex
}

type Mutex interface {
	// Lock 获取锁，阻塞等待
	Lock() error

	// LockCtx 获取锁，阻塞等待
	LockCtx(ctx context.Context) error

	// TryLock 尝试获取锁
	TryLock() error

	// TryLockCtx 尝试获取锁
	TryLockCtx(ctx context.Context) error

	// Unlock 释放锁
	Unlock() (bool, error)

	// UnlockCtx 释放锁
	UnlockCtx(ctx context.Context) (bool, error)
}

// Options mutex options
type Options struct {
	ttl        time.Duration
	timeout    time.Duration
	genValueFn GenValueFn
	tries      int // 重试次数
}

// Option mutex 选项设置
type Option func(*Options)

// WithTTL 设置锁过期时间
func WithTTL(ttl time.Duration) Option {
	return Option(func(o *Options) {
		o.ttl = ttl
	})
}

// WithTimeout 设置获取锁超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.timeout = timeout
	}
}

// WithTries 获取锁失败重试次数
func WithTries(tries int) Option {
	return func(o *Options) {
		o.tries = tries
	}
}

// WithGenValueFunc 自定义 value 生成
func WithGenValueFunc(genValueFn GenValueFn) Option {
	return func(o *Options) {
		o.genValueFn = genValueFn
	}
}
