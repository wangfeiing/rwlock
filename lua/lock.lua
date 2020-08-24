-- 基于lua的读写锁，可以保证原子性
-- Redis对Lua的支持是有限制，不支持require，只能做成单文件

local rProfix = "_read_for_lock__"
local wProfix = "_write_for_lock__"
local lockKey = KEYS[1]
local cmdKey  = KEYS[2]

-- 公平锁队列
local queueKey = "_wait_queue__" .. lockKey

-- 判断队列内某个元素是否存在
local existHashKey = "_wait_queue_hash_set__" .. lockKey

-- 锁的唯一ID
local lockUniqKey = ARGV[1]
if ARGV[2] == nil
then
    ARGV[2] = "100"
end

--超时时间
local expireNum = tonumber(ARGV[2])

-- 客户端是否等待监测
local onlineKey = "_waiting_exipre_lock_key__" .. lockKey .. "_uniqueID__" ..lockUniqKey


--读锁key
local readLockKey = rProfix .. lockKey
--写锁key
local writeLockKey = wProfix .. lockKey
local errorString = ""
local debugString = ""
local Ok =  "OK"

local function getOnlineKey(uniqKey)
    return "_online_exipre_lock_key__" .. lockKey .. "_uniqueID__" .. uniqKey
end

local function get(key)
    return redis.call("GET", key)
end

local function hset(key, field, value)
    return  redis.call("HSET", key , field , value)
end

local function hdel(key, field)
    return  redis.call("HDEL", key , field)
end

local function hexists(key , field)
    local ret =  redis.call("HEXISTS", key , field )
    if ret > 0
    then
        return true
    end
    return false
end

local function llen(key)
    return  redis.call("LLEN", key)
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

local function exists(key)
    return  redis.call("EXISTS" , key)
end

local function rpush(key , val)
    return redis.call("RPUSH" , key , val)
end

local function lpop(key)
    return redis.call("LPOP",key)
end

local function lrem(key, count , val)
    return redis.call("LREM",key, count, val)
end

local function lindex(key , index)
    return redis.call("LINDEX" , key ,index)
end

local function range(key , startIdx , endIdx)
    return redis.call("LRANGE",key ,startIdx ,endIdx)
end


-- ------- 公平锁逻辑 ------
-- 刷新hearbeat
local function onlineHeartbeat()
    set(onlineKey  , 1)
    expire(onlineKey, 1)
end

local function isOnline(uniqueID)
    local tmpOnlineKey = getOnlineKey(uniqueID)
    local ret = exists(tmpOnlineKey)
    if ret > 0
    then
        return true
    end
    return false
end

-- 队列入队
local function existQueue()
    return hexists(existHashKey , lockUniqKey)
end

local function enQueue()
    local uniqueID = lockUniqKey
    local exist =  existQueue(uniqueID)
    if exist
    then
        return true
    end
    hset(existHashKey, uniqueID, 1)
    local ret =  rpush(queueKey , uniqueID)
    if ret <= 0
    then
--     如果rpush失败，就回滚
        hdel(existHashKey , uniqueID)
        return false
    end

    return true
end

-- 队列第一个元素出队
local function deQueue()
    return lpop(queueKey)
end

-- 读队列第一个元素
local function front()
    return lindex(queueKey , 0)
end

-- 队列的长度

local function countQueue()
    return llen(queueKey)
end


local function isSelf()
--    如果队列没有元素，直接让自己获取锁
    local count = countQueue()
    if count == 0
    then
        return true
    end

    local frontOne = front()
    if frontOne == lockUniqKey
    then
        return true
    end
    return false
end


local function handleLockFail()
    local count = countQueue()
    if count > 0
    then
        --   读取队列第一个元素
        local frontOne = front()
        --    判断第一个元素是否在线
        local frontOneOnline = isOnline(frontOne)
        -- 如果不在线 就从队列移除
        if frontOneOnline == false
        then
            deQueue()
            --从删除hash表中删除
            hdel(existHashKey, frontOne)
        end
    end

--  自身入队
    enQueue()

end
--处理加锁成功的情况
local function handleLockSuccess()

    -- 自身出队
    deQueue()

    --  删除
    del(onlineKey)

    --从删除 hash 队列中删除
    hdel(existHashKey, lockUniqKey)
end

-- write lock
local function lock()
    --   维护一下自己的心跳
    onlineHeartbeat()

    -- 如果有读锁在用
    -- 直接则返回false
    -- 表示上锁失败
    local ret = get(readLockKey)
    if  ret ~= false and  tonumber(ret) > 0
    then
        debugString = "read lock number(".. ret ..") > 0"
--      如果拿不到锁，就进入单独的逻辑处理一下
        handleLockFail()
        return false
    end
    -- 表示写锁已经被加上
    -- 写锁失败
    -- 返回false
    local wret = get(writeLockKey)
    if wret ~= false and string.len(wret) > 0
    then
        debugString = "write lock be set by other"
--        锁被别人占用了，mmp
        handleLockFail()
        return false
    end
--    检查是否轮到自己
    local isTurnMe = isSelf()
    if isTurnMe == false
    then
        debugString = "lock is free,but not turn me,lockKey="..lockKey.."uniqueID=" .. lockUniqKey
        handleLockFail()
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
--    处理加锁成功
    handleLockSuccess()
    return true
end

local function unlock()
    local ret = get(writeLockKey)
    -- 如果当前锁不存在
    if ret == false or string.len(ret) <= 0
    then
        debugString = "Unlock of unlocked RWMutex"
        return true
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

local function runlock()
    local rlock = get(readLockKey)
    if rlock == false or tonumber(rlock) <= 0
    then
        errorString = "RUnlock of unlocked"
        return false
    end
    decr(readLockKey)
    return true
end

-- 处理锁逻辑
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