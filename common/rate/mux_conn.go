package rate

import (
	"net"
	"sync"
	"time"

	"github.com/juju/ratelimit"
	"github.com/sagernet/sing/common/buf"
	"github.com/sagernet/sing/common/bufio"
	"github.com/sagernet/sing/common/network"
)

// MuxConnRateLimiter is a soft rate limiter for mux connections.
// It caps each wait at a short duration to avoid long blocks that can break mux streams.
//
// Important: do not implement network.ReadCounter/network.WriteCounter (UnwrapReader/UnwrapWriter)
// here, otherwise sing's UnwrapCountReader/UnwrapCountWriter can bypass this limiter.
type MuxConnRateLimiter struct {
	network.ExtendedConn
	limiter *ratelimit.Bucket
	mu      sync.Mutex
}

func NewMuxConnRateLimiter(conn net.Conn, limiter *ratelimit.Bucket) *MuxConnRateLimiter {
	return &MuxConnRateLimiter{
		ExtendedConn: bufio.NewExtendedConn(conn),
		limiter:      limiter,
	}
}

// tryThrottle consumes tokens without blocking.
func (c *MuxConnRateLimiter) tryThrottle(n int64) {
	if c.limiter == nil || n <= 0 {
		return
	}
	_ = c.limiter.TakeAvailable(n)
}

// softWait implements gentle throttling that won't block for too long.
func (c *MuxConnRateLimiter) softWait(n int64) {
	if c.limiter == nil || n <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	const maxWait = 100 * time.Millisecond
	delay := c.limiter.Take(n)
	if delay <= 0 {
		return
	}
	if delay > maxWait {
		delay = maxWait
	}
	time.Sleep(delay)
}

func (c *MuxConnRateLimiter) Read(b []byte) (n int, err error) {
	n, err = c.ExtendedConn.Read(b)
	if n > 0 {
		c.tryThrottle(int64(n))
	}
	return
}

func (c *MuxConnRateLimiter) Write(b []byte) (n int, err error) {
	if len(b) > 0 {
		c.softWait(int64(len(b)))
	}

	n, err = c.ExtendedConn.Write(b)
	if n > 0 && n != len(b) {
		c.tryThrottle(int64(n))
	}
	return
}

func (c *MuxConnRateLimiter) ReadBuffer(buffer *buf.Buffer) error {
	err := c.ExtendedConn.ReadBuffer(buffer)
	if err != nil {
		return err
	}
	if buffer.Len() > 0 {
		c.tryThrottle(int64(buffer.Len()))
	}
	return nil
}

func (c *MuxConnRateLimiter) WriteBuffer(buffer *buf.Buffer) error {
	dataLen := int64(buffer.Len())
	if dataLen > 0 {
		c.softWait(dataLen)
	}
	return c.ExtendedConn.WriteBuffer(buffer)
}

func (c *MuxConnRateLimiter) Upstream() any {
	return c.ExtendedConn
}
