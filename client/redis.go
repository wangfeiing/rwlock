package client

import "github.com/go-redis/redis"

func init() {
	opt := &redis.Options{
		Network:"",
		Addr:"",
		Password:"",
		WriteTimeout:10,
		ReadTimeout:10,
	}
	redis.NewClient(opt)
}
