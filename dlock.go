package dlock

import (
	"context"
)

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

// Option mutex 设置接口
type Option interface {
	Apply(Mutex)
}

// OptionFunc mutex 选项设置
type OptionFunc func(Mutex)

// Apply 调用 mutex 配置
func (f OptionFunc) Apply(mutex Mutex) {
	f(mutex)
}
