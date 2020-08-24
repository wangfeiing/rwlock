package rwlock

import (
	"crypto/tls"
	"github.com/go-redis/redis"
	"github.com/wangfeiso/rwlock/client"
	"time"
)

// 从redis侧拷贝
type Options struct {
	// The network type, either tcp or unix.
	// Default is tcp.
	Network string
	// host:port address.
	Addr string

	// Optional password. Must match the password specified in the
	// requirepass server configuration option.
	Password string
	// Database to be selected after connecting to the server.
	DB int

	// Maximum number of retries before giving up.
	// Default is to not retry failed commands.
	MaxRetries int
	// Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff time.Duration
	// Maximum backoff between each retry.
	// Default is 512 milliseconds; -1 disables backoff.
	MaxRetryBackoff time.Duration

	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout time.Duration
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout time.Duration

	// Maximum number of socket connections.
	// Default is 10 connections per every CPU as reported by runtime.NumCPU.
	PoolSize int
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns int
	// Connection age at which client retires (closes) the connection.
	// Default is to not close aged connections.
	MaxConnAge time.Duration
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	// Default is ReadTimeout + 1 second.
	PoolTimeout time.Duration
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout time.Duration
	// Frequency of idle checks made by idle connections reaper.
	// Default is 1 minute. -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency time.Duration

	// Enables read only queries on slave nodes.
	readOnly bool

	// TLS Config to use. When set TLS will be negotiated.
	TLSConfig *tls.Config
}

type RWMutex struct {
	lockKey   string
	uniqID    string
	expire    int64
	retryTime int64
}

// Init
// 初始化redis客户端
func Init(opt *Options) {
	redisOpts := redis.Options{
		Network:            opt.Network,
		Addr:               opt.Addr,
		Password:           opt.Password,
		DB:                 opt.DB,
		MaxRetries:         opt.MaxRetries,
		MinRetryBackoff:    opt.MinRetryBackoff,
		MaxRetryBackoff:    opt.MaxRetryBackoff,
		DialTimeout:        opt.DialTimeout,
		ReadTimeout:        opt.ReadTimeout,
		WriteTimeout:       opt.WriteTimeout,
		PoolSize:           opt.PoolSize,
		MinIdleConns:       opt.MinIdleConns,
		MaxConnAge:         opt.MaxConnAge,
		PoolTimeout:        opt.PoolTimeout,
		IdleTimeout:        opt.IdleTimeout,
		IdleCheckFrequency: opt.IdleCheckFrequency,
		TLSConfig:          opt.TLSConfig,
	}
	// redis属于基础资源 如果redis的客户端初始化失败，直接panic，没得商量
	if err := client.Init(&redisOpts); err != nil {
		panic("redis client init ")
	}
}
