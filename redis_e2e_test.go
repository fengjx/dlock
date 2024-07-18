//go:build e2e

package dlock

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// 集成测试

func getRedisLockClient() LockClient {
	client := redis.NewClient(&redis.Options{
		Addr: "192.168.6.121:6379",
	})
	lock := NewRedisLock(client)
	return lock
}

func TestE2ELock(t *testing.T) {
	mutex := getRedisLockClient().NewMutex("e2e-lock",
		WithTTL(10*time.Second),
		WithWaitTimeout(3*time.Second),
		WithTries(3),
	)
	ctx := context.Background()
	lock, err := mutex.Lock(ctx)
	assert.NoError(t, err)
	_, err = mutex.Lock(ctx)
	assert.Equal(t, ErrTimeout, err)
	ok, err := lock.Unlock(ctx)
	assert.NoError(t, err)
	assert.Equal(t, true, ok)
}

func TestE2ESum(t *testing.T) {
	mutex := getRedisLockClient().NewMutex("e2e-lock-sum",
		WithTTL(10*time.Second),
		WithWaitTimeout(60*time.Second),
		WithTries(3),
	)
	var sum int
	var wg sync.WaitGroup
	wg.Add(100)
	for i := 1; i <= 100; i++ {
		go func(idx int) {
			defer wg.Done()
			ctx := context.Background()
			lock, err := mutex.Lock(ctx)
			if err != nil {
				t.Fatal(idx, err)
			}
			defer func() {
				ok, err := lock.Unlock(ctx)
				if !ok {
					t.Fatal(idx, "unlock failed", ok)
				}
				if err != nil {
					t.Fatal(idx, "unlock err", err)
				}
			}()
			t.Log(idx)
			sum += idx
		}(i)
	}
	wg.Wait()
	assert.Equal(t, 5050, sum)
}

func TestE2ETryLock(t *testing.T) {
	mutex := getRedisLockClient().NewMutex("e2e-try-lock",
		WithTTL(10*time.Second),
		WithWaitTimeout(60*time.Second),
		WithTries(3),
	)
	ctx := context.Background()
	lock, err := mutex.TryLock(ctx)
	assert.NoError(t, err)
	_, err = mutex.TryLock(ctx)
	assert.Equal(t, ErrFailed, err)
	ok, err := lock.Unlock(ctx)
	assert.NoError(t, err)
	assert.Equal(t, true, ok)
}
