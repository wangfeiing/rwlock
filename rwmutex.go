package rwlock

import (
	"github.com/go-redis/redis"
	"github.com/wangfeiso/rwlock/client"
	"sync"
)

type Options struct {
	redis.Options
}

var shaHashID *string

type RWMutex struct {
	shaHashID *string
	lockKey   string
}

func Init(opt *Options) {
	client.Init(&opt.Options)
	tmp := client.InitLua()
	shaHashID = &tmp
}

var l sync.RWMutex

func NewRWMutex() *RWMutex {

	return &RWMutex{
		shaHashID: shaHashID,
	}
}
