--基于lua的读写锁，可以保证原子性

local rProfix = "_read_for_lock__"
local wProfix = "_write_for_lock__"
local lockKey = KEYS[1]
local cmdKey  = KEYS[2]
local lockUniqKey = ARGV[1]
local expireNum = tonumber(ARGV[2])


local readLockKey = rProfix .. lockKey
local writeLockKey = wProfix .. lockKey
local errorString = ""
local debugString = ""
local Ok =  "OK"

local function get(key)
    return redis.call("GET", key)
end

local function hget(key , field)
    return  redis.call("HGET", key, field)
end

local function hset(key, field, value)
    return  redis.call("HSET", key , field , value)
end

local function hgetall(key)
    local ret =  redis.call("HGETALL", key)
    local step = 2
    local returnTable = {}

    for i = 1 , table.getn(ret) - 1, step  do
        returnTable[ret[i]] = ret[i+1]
    end
    return returnTable
end

local function hlen(key)
    return  redis.call("HLEN", key)
end

local function hdel(key)
    return  redis.call("HDEL" , key)
end

local function set(key , value)
    local ret = redis.call("SET", key, value)
    return ret['ok']
end

local function incr(key)

    return  redis.call("INCR", key)
end

local function decr(key)
    return  redis.call("DECR", key)
end

local function expire(key, sec)
    return  redis.call("EXPIRE", key, tonumber(sec))
end

local function del(key)
    return  redis.call("DEL" , key)
end

-- write lock
local function lock()
    -- 如果有读锁在用
    -- 直接则返回false
    -- 表示上锁失败
    local ret = get(readLockKey)
    if  ret ~= false and  tonumber(ret) > 0
    then
        debugString = "read lock number(".. ret ..") > 0"
        return false
    end
    -- 表示写锁已经被加上
    -- 写锁失败
    -- 返回false
    local wret = get(writeLockKey)
    if wret ~= false and string.len(wret) > 0
    then
        debugString = "write lock be set by other"
        return false
    end
    --  开始加锁
    local incrRet = set(writeLockKey, lockUniqKey)
    if incrRet ~= Ok
    then
        debugString = "write lock set fail,key=" .. writeLockKey .. ",lockUniqKey=" .. lockUniqKey
        return false
    end
    -- 设置过期时间
    local expireRet = expire(writeLockKey, expireNum)
    if expireRet <= 0
    then
        -- 回滚
        del(writeLockKey)
        debugString = "write lock expire fail,key=" .. writeLockKey .. ",expireNum=" .. expireNum
        return false
    end

    return true
end

local function unlock()
    local ret = get(writeLockKey)
    -- 如果当前锁不存在
    if ret == false or string.len(ret) <= 0
    then
        errorString = "Unlock of unlocked RWMutex"
        return false
    end

    -- 判断当前锁是否不是自己加的
    -- 证明锁超时了被释放，并且被别人抢走
    if ret ~= lockUniqKey
    then
        debugString = "write unlock one timeout key ,key==" .. writeLockKey .. ",expectUniqKey=" .. lockUniqKey .. ",newUniqKey=" .. ret
        return true
    end
    -- 删除当前的key
    local retDel = del(writeLockKey)
    if retDel <= 0
    then
        debugString = "write unlock del fail,key==" .. writeLockKey
        return false
    end

    return true
end

local function rlock()
    local wlock = get(writeLockKey)

    -- 查一下是否有写锁
    if wlock ~= false and string.len(wlock) > 0
    then
        debugString = "read rlock fail,write lock occupy now,key==" .. writeLockKey .. ",occupyUniqKey=" .. wlock
        return false
    end

    local retIncr = incr(readLockKey)
    if retIncr > 0
    then
        return true
    end

    debugString = "read rlock fail,incr not expect,key==" .. readLockKey .. ",retIncr=" .. retIncr
    return false
end

local function runlock(key)
    local rlock = get(readLockKey)
    if rlock == false or tonumber(rlock) <= 0
    then
        errorString = "RUnlock of unlocked"
        return false
    end
    decr(readLockKey)
    return true
end

local function handleLock()
    if cmdKey == "LOCK"
    then
        if string.len(lockUniqKey) <= 0
        then
            errorString = "unque key is nil"
            return false
        end
        if string.len(writeLockKey) <= 0
        then
            errorString = "Lock key is nil"
            return false
        end
        return lock()
    end

    if cmdKey == "UNLOCK"
    then
        if string.len(lockUniqKey) <= 0
        then
            errorString = "unque key is nil"
            return false
        end
        if string.len(writeLockKey) <= 0
        then
            errorString = "Lock key is nil"
            return false
        end
        return unlock()
    end

    if cmdKey == "RLOCK"
    then
        if string.len(readLockKey) <= 0
        then
            errorString = "Rlock key is nil"
            return false
        end
        return rlock()
    end
    if cmdKey == "RUNLOCK"
    then
        if string.len(readLockKey) <= 0
        then
            errorString = "RUnlock key is nil"
            return false
        end
        return runlock()
    end

    errorString = "Unkown rwlock Command"
    return false
end

local opRet = handleLock()

return cjson.encode({
    opRet = opRet,
    debug = debugString,
    errMsg = errorString
})