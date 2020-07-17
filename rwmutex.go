package rwlock

import (
	"github.com/go-redis/redis"
	"github.com/wangfeiso/rwlock/client"
	"github.com/wangfeiso/rwlock/tool"
	"sync"
)

type Options struct {
	redis.Options
}

var shaHashID *string

type RWMutex struct {
	lockKey   string
	uniqID    string
	expire    int64
	retryTime int64
}

func Init(opt *Options) {
	client.Init(&opt.Options)
	client.InitLua()

}

var l sync.RWMutex

func NewRWMutex() *RWMutex {

	return &RWMutex{
		uniqID:  tool.GetUUID(),
		lockKey: tool.GetUUID(),
		expire:  10,
	}
}

func (rw *RWMutex) Lock() {
	client.Lock(rw.lockKey, rw.uniqID, rw.expire)
}

func (rw *RWMutex) Unlock() {
	client.Unlock(rw.lockKey, rw.uniqID)
}

func (rw *RWMutex) RLock() {
	client.RLock(rw.lockKey)
}

func (rw *RWMutex) RUnLock() {
	client.RUnlock(rw.lockKey)
}
