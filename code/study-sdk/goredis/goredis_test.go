package goredis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"testing"
)

var defalutCtx = context.Background()

func TestGoRedis(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	defer rdb.Close()
	cmd1 := rdb.Set(defalutCtx, "key1", "value1", 0)
	t.Log(cmd1.Result())
	cmd2 := rdb.Get(defalutCtx, "key1")
	t.Log(cmd2.Result())
	cmd3 := rdb.Del(defalutCtx, "key1")
	t.Log(cmd3.Result())
	cmd4 := rdb.Get(defalutCtx, "key1")
	t.Log(cmd4.Result())
}
