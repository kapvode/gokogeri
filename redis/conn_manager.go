package redis

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/kapvode/gokogeri"
)

var _ gokogeri.ConnProvider = (*ConnManager)(nil)

// ConnManager implements the ConnProvider interface and encapsulates the process of establishing and configuring Redis
// connections. The non-dedicated connections use redis.Pool.
type ConnManager struct {
	cfg  *Config
	pool redis.Pool
}

// NewConnManager returns a new instance. Make sure to call Close when you are done.
func NewConnManager(cfg *Config) *ConnManager {
	return &ConnManager{
		cfg: cfg,
		pool: redis.Pool{
			DialContext: func(ctx context.Context) (redis.Conn, error) {
				return redis.DialURLContext(
					ctx,
					cfg.URL,
					redis.DialReadTimeout(cfg.ReadTimeout),
					redis.DialWriteTimeout(cfg.WriteTimeout),
				)
			},
			MaxIdle:     cfg.MaxIdle,
			MaxActive:   cfg.MaxActive,
			IdleTimeout: cfg.IdleTimeout,
			Wait:        true,
		},
	}
}

// Conn implements ConnProvider. It returns connections from a shared pool.
func (cm *ConnManager) Conn(ctx context.Context) (redis.Conn, error) {
	return cm.pool.GetContext(ctx)
}

// DialLongPoll implements ConnProvider.
func (cm *ConnManager) DialLongPoll(ctx context.Context) (redis.Conn, error) {
	return redis.DialURLContext(
		ctx,
		cm.cfg.URL,
		redis.DialReadTimeout(time.Second*time.Duration(cm.cfg.LongPollTimeout)+cm.cfg.ReadTimeout),
		redis.DialWriteTimeout(cm.cfg.WriteTimeout),
	)
}

// Close releases resources used by the connection pool.
func (cm *ConnManager) Close() error {
	return cm.pool.Close()
}
