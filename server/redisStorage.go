package server

import "github.com/gomodule/redigo/redis"

type redisOp struct {
	readisPool *redis.Pool
}
