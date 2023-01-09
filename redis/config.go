package redis

import "time"

// Config holds the configuration for Redis connections.
type Config struct {
	// URL is used to connect to Redis.
	URL string

	// LongPollTimeout is the timeout in seconds for the BRPOP Redis command that reads from the queues. Zero means no
	// timeout, but it is better that you set a value, because it also has the effect of pinging the connection, to make
	// sure it is still active.
	LongPollTimeout int

	// MaxIdle is a parameter for redis.Pool.
	MaxIdle int

	// MaxActive is a parameter for redis.Pool.
	MaxActive int

	// IdleTimeout is a parameter for redis.Pool.
	IdleTimeout time.Duration

	// ReadTimeout is a parameter for redis.DialReadTimeout.
	ReadTimeout time.Duration

	// WriteTimeout is a parameter for redis.DialWriteTimeout.
	WriteTimeout time.Duration
}

// NewDefaultConfig returns the default configuration.
func NewDefaultConfig() *Config {
	return &Config{
		URL:             "redis://localhost",
		LongPollTimeout: 30,
		MaxIdle:         2,
		MaxActive:       2,
		IdleTimeout:     time.Minute,
		ReadTimeout:     time.Second * 5,
		WriteTimeout:    time.Second * 5,
	}
}
