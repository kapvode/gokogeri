package gokogeri

import (
	"context"

	"github.com/gomodule/redigo/redis"
)

// ConnProvider provides Redis connections, while encapsulating the process of establishing and configuring the
// connections.
// It is safe for concurrent use.
//
// The provided Context should only affect the process of establishing a connection. If the context expires afterwards,
// it should not affect the use of the connection.
type ConnProvider interface {
	// Conn returns a connection, which can come from a shared pool. The caller will call Close on the connection when
	// it is done with it.
	Conn(context.Context) (redis.Conn, error)

	// DialLongPoll returns a new, dedicated connection, with a long read timeout and a normal write timeout. The caller
	// will close the connection.
	DialLongPoll(context.Context) (redis.Conn, error)
}
