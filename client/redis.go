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

var currentVersion = ""
var shaHashID string

func Init(opt *redis.Options) {
	Redis = redis.NewClient(opt)
	_, err := Redis.Ping().Result()
	if err != nil {
		panic(err)
		return
	}
	opts = opt
	currentVersion = strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond)))
	InitLua()
}

func InitLua() {
	hashID, err := Redis.ScriptLoad(lua.ScriptContent).Result()
	if err != nil {
		return
	}
	fmt.Println(hashID)
	SetShaHasID(hashID)
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
	shaHashID := GetShaHashID()
	for {
		res, err := send(shaHashID, key, uniqID, LockCmd, expireTime)
		if err != nil {
			handleError(err)
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
	res, err := send(GetShaHashID(), key, uniqID, UnlockCmd, 0)
	if res.Success() {
		return
	}
	if res.IsError() {
		panic(res.Error())
	}
	if err != nil {
		handleError(err)
	}
}

func RLock(key string) {
	for {
		res, err := send(GetShaHashID(), key, "", RLockCmd, 0)
		if res.Success() {
			return
		}
		if err != nil {
			handleError(err)
		}
		time.Sleep(getRandomSleepTime())
	}
}

func RUnlock(key string) {
	res, err := send(GetShaHashID(), key, "", RUnlockCmd, 0)
	if res.Success() {
		return
	}
	if err != nil {
		panic(err)
		handleError(err)
	}
}

func getRandomSleepTime() time.Duration {
	return time.Duration(tool.Rand(10, 20)) * time.Millisecond
}

func send(shaHashID, key string, uniqID, lockCmd string, expireTime int64) (*responseLock, error) {
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

func handleError(err error) {
	if err.Error() == EofError {
		handleEofError()
	}
	if err.Error() == NoScriptError {
		fmt.Println(err.Error())
		handleNoScriptError()
	}
}

// redis重启
func handleEofError() {
	Init(opts)
}

//script
func handleNoScriptError() {
	InitLua()
}
