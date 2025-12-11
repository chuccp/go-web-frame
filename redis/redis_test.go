package redis

import (
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestName(t *testing.T) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

}
