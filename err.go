package dlock

import "errors"

var (
	// ErrFailed 获取锁失败
	ErrFailed = errors.New("dlock: failed to acquire lock")

	// ErrTimeout TryLock 超时
	ErrTimeout = errors.New("dlock: acquire lock timeout")

	// ErrLockNotExistOrExpired 锁不存在或已过期
	ErrLockNotExistOrExpired = errors.New("dlock: failed to unlock, lock not exist or already expired")
)
