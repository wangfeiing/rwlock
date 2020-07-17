package client

import (
	"github.com/go-redis/redis"
	"github.com/wangfeiso/rwlock/lua"
)

var Redis *redis.Client

const LockCmd = "LOCK"
const UnlockCmd = "UNLOCK"

func Init(opt *redis.Options) {
	//opt := &redis.Options{
	//	Network:      "",
	//	Addr:         "",
	//	Password:     "",
	//	WriteTimeout: 10,
	//	ReadTimeout:  10,
	//}
	Redis = redis.NewClient(opt)
	ping, err := Redis.Ping().Result()
	if err != nil {
		return
	}
	if ping != "PONG" {

	}

	InitLua()
}

func InitLua() string {
	hashID, err := Redis.ScriptLoad(lua.ScriptContent).Result()
	if err != nil {
		return ""
	}
	return hashID
}

func Lock() {

}

func Unlock() {

}

func RLock() {

}

func RUnlock() {

}
