# rwlock

### 使用方式

``` 
go get github.com/wangfeiso/rwlock
```
### 特性说明

为确保锁的公平性，用Lua实现了优先级队列FCFS，当多个客户端获取写锁（排它锁）的时候，先到的会先获得锁。


### 快速使用

```
import (
	"github.com/wangfeiso/rwlock"
)

func main() {
    
    // 初始化redis客户端，需要传入redis-server的ip和port
    // 仅支持单机Redis
    rwlock.Init(&rwlock.Options{
        Addr: "127.0.0.1:6379",
    })
    
    //开始使用分布式读写锁
    
    //创建一个锁
    lock := rwlock.New("YourLockKey")
    
    // 加上写锁
    lock.Lock()
    
    // 释放写锁
    lock.Unlock()
    
    // 加上读锁
    lock.RLock()
    
    // 释放读锁
    lock.RUnlock()
    
}

```

### 说明
读写锁之间的互斥性如下

| 互斥性 | 读锁 | 写锁 |
| :-----| ----: | :----: |
| 读锁 | 兼容 | 互斥 |
| 写锁 | 互斥 | 互斥 |

读锁与写锁不能同时存在，读锁和读锁可以同时存在

### TODO
* 补全单元测试
