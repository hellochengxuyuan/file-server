package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

var (
	pool      *redis.Pool
	redisHost = "127.0.0.1:6379"
	redisPass = "testupload"
)

// 创建redis连接池
func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   30,
		IdleTimeout: 300 * time.Second,
		Dial: func() (conn redis.Conn, err error) {
			// 1. 打开连接
			if conn, err = redis.Dial("tcp", redisHost); err != nil {
				fmt.Println(err)
				return
			}

			// 2. 访问认证
			if _, err = conn.Do("AUTH", redisPass); err != nil {
				conn.Close()
				return
			}
			return
		},
		// 检测redis的健康状态，如果有错误在客户端上关闭redis的连接
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := conn.Do("PING")
			return err
		},
	}
}

func init() {
	pool = newRedisPool()
}

func RedisPool() *redis.Pool {
	return pool
}
