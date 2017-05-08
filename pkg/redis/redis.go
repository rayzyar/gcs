package rediscli

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

var redisPool = createRedisPool()

func createRedisPool() *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379", redis.DialReadTimeout(time.Second), redis.DialWriteTimeout(time.Second), redis.DialConnectTimeout(time.Second))
		},
		IdleTimeout: time.Second,
		MaxActive:   10,
		MaxIdle:     4,
	}
}

func GetConn() redis.Conn {
	return redisPool.Get()
}
