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

// Lock 锁信息
type Lock interface {
	// Unlock 释放锁
	Unlock(ctx context.Context) (bool, error)
}

// Mutex 锁
type Mutex interface {
	// Lock 获取锁，阻塞等待
	Lock(ctx context.Context) (Lock, error)

	// TryLock 尝试获取锁
	TryLock(ctx context.Context) (Lock, error)
}

// Options mutex options
type Options struct {
	ttl         time.Duration
	waitTimeout time.Duration
	genValueFn  GenValueFn
	tries       int // 重试次数
}

// Option mutex 选项设置
type Option func(*Options)

// WithTTL 设置锁过期时间
func WithTTL(ttl time.Duration) Option {
	return Option(func(o *Options) {
		o.ttl = ttl
	})
}

// WithWaitTimeout 设置获取锁等待时间
func WithWaitTimeout(waitTimeout time.Duration) Option {
	return func(o *Options) {
		o.waitTimeout = waitTimeout
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
