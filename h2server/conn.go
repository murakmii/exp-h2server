package h2server

import (
	"net"
	"sync/atomic"
)

type (
	Conn interface {
		RemoteAddr() net.Addr
		Write(f Frame)
		Close()
	}

	pseudoConn struct {
		addr   net.Addr
		w      chan Frame
		closed int32
	}
)

var (
	_ Conn = (*pseudoConn)(nil)
)

func newPseudoConn(source net.Conn) *pseudoConn {
	return &pseudoConn{
		addr: source.RemoteAddr(),
		w:    make(chan Frame),
	}
}

func (c *pseudoConn) RemoteAddr() net.Addr {
	return c.addr
}

func (c *pseudoConn) Write(f Frame) {
	if atomic.LoadInt32(&c.closed) == 1 {
		return
	}

	c.w <- f
}

func (c *pseudoConn) Close() {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		close(c.w)
	}
}

func (c *pseudoConn) writeCh() chan Frame {
	return c.w
}
