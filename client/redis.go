package client

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/wangfeiso/rwlock/lua"
	"github.com/wangfeiso/rwlock/tool"
	"strconv"
	"time"
)

var Redis *redis.Client
var opts *redis.Options

// error 定义
const NoScriptError = "NOSCRIPT No matching script. Please use EVAL."
const EofError = "EOF"

const LockCmd = "LOCK"
const UnlockCmd = "UNLOCK"
const RLockCmd = "RLOCK"
const RUnlockCmd = "RUNLOCK"

var shaHashID string

func Init(opt *redis.Options) error {
	Redis = redis.NewClient(opt)
	if _, err := Redis.Ping().Result(); err != nil {
		return err
	}
	opts = opt

	if err := InitLua(); err != nil {
		return err
	}
	return nil
}

func InitLua() error {
	hashID, err := Redis.ScriptLoad(lua.ScriptContent).Result()
	if err != nil {
		return err
	}
	fmt.Println(hashID)
	SetShaHasID(hashID)
	return nil

}
func GetShaHashID() string {
	return shaHashID
}
func SetShaHasID(str string) {
	shaHashID = str
}

type responseLock struct {
	OpRet  bool   `json:"opRet"`
	ErrMsg string `json:"errMsg"`
	Debug  string `json:"debug"`
}

func (r responseLock) IsError() bool {
	if len(r.ErrMsg) > 0 {
		return true
	}
	return false
}
func (r responseLock) Success() bool {
	return r.OpRet
}
func (r responseLock) Error() string {
	return r.ErrMsg
}

func Lock(key string, uniqID string, expireTime int64) {
	if len(key) < 0 {
		panic("lock key is nil")
	}
	if expireTime <= 0 {
		expireTime = 5
	}
	for {
		res, err := sendLock(GetShaHashID(), key, uniqID, LockCmd, expireTime)
		if err != nil {
			handleError(err)
			time.Sleep(getRandomSleepTime())
			continue
		}
		if res != nil && res.IsError() {
			panic(res.Error())
		}
		if res != nil && res.Success() {
			return
		}

		time.Sleep(getRandomSleepTime())
	}
}

func Unlock(key, uniqID string) {
	i := 10
	for {
		res, err := sendLock(GetShaHashID(), key, uniqID, UnlockCmd, 0)
		if res != nil && res.Success() {
			return
		}
		if res != nil && res.IsError() {
			panic(res.Error())
		}
		if err != nil {
			handleError(err)
		}
		if i--; i <= 0 {
			return
		}
		time.Sleep(getRandomSleepTime())
	}
}

func RLock(key string) {
	for {
		res, err := sendLock(GetShaHashID(), key, "", RLockCmd, 0)
		if res != nil && res.Success() {
			return
		}
		if err != nil {
			handleError(err)
		}

		time.Sleep(getRandomSleepTime())
	}
}

func RUnlock(key string) {
	if len(key) <= 0 {
		panic("runlock nil key")
	}
	i := 10
	for {
		res, err := sendLock(GetShaHashID(), key, "", RUnlockCmd, 0)
		if res != nil && res.Success() {
			return
		}
		if err != nil {
			handleError(err)
		}

		if i--; i <= 0 {
			return
		}
		time.Sleep(getRandomSleepTime())
	}
}

func getRandomSleepTime() time.Duration {
	return time.Duration(tool.Rand(10, 20)) * time.Millisecond
}

func sendLock(shaHashID, key string, uniqID, lockCmd string, expireTime int64) (*responseLock, error) {
	var ret interface{}
	var err error
	switch lockCmd {
	case LockCmd:
		ret, err = Redis.EvalSha(shaHashID, []string{key, lockCmd}, []string{uniqID, strconv.Itoa(int(expireTime))}).Result()
	case UnlockCmd:
		ret, err = Redis.EvalSha(shaHashID, []string{key, lockCmd}, []string{uniqID}).Result()
	case RLockCmd, RUnlockCmd:
		ret, err = Redis.EvalSha(shaHashID, []string{key, lockCmd}, []string{}).Result()
	}

	if err != nil {
		return nil, err
	}
	var retJson = ret.(string)
	var res responseLock
	if err := json.Unmarshal([]byte(retJson), &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func handleError(err error) bool {
	if err == nil {
		return false
	}
	switch err.Error() {
	case EofError:
		if err := handleEofError(); err != nil {
			return false
		}
		return true
	case NoScriptError:
		if err := handleNoScriptError(); err != nil {
			return false
		}
		return true
	default:
		return false
	}

}

// redis重启
func handleEofError() error {
	return Init(opts)
}

// script 不存在
func handleNoScriptError() error {
	return InitLua()
}
