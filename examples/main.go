package main

import (
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
		dlock.WithTimeout(3*time.Second),
		dlock.WithTries(3),
	)
	err := mutex.TryLock()
	if err != nil {
		log.Panic(err)
	}
	defer mutex.Unlock()
}
