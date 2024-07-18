package main

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/fengjx/dlock"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "192.168.6.121:6379",
	})
	lockCli := dlock.NewRedisLock(client)
	mutex := lockCli.NewMutex("examples",
		dlock.WithTTL(10*time.Second),
		dlock.WithWaitTimeout(3*time.Second),
		dlock.WithTries(3),
	)
	ctx := context.Background()
	lock, err := mutex.TryLock(ctx)
	if err != nil {
		log.Panic(err)
	}
	defer lock.Unlock(ctx)
}
