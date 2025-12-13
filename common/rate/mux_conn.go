package rate

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/juju/ratelimit"
	"github.com/sagernet/sing/common/buf"
	"github.com/sagernet/sing/common/bufio"
	"github.com/sagernet/sing/common/network"
)

// MuxConnRateLimiter is a non-blocking rate limiter for mux connections
// It uses buffer-level limiting and dynamic throttling to avoid blocking
type MuxConnRateLimiter struct {
	network.ExtendedConn
	limiter      *ratelimit.Bucket
	readFunc     network.CountFunc
	writeFunc    network.CountFunc
	lastReadTime atomic.Value // time.Time
	readBytes    atomic.Int64
	writeBytes   atomic.Int64
	mu           sync.Mutex
}

func NewMuxConnRateLimiter(conn net.Conn, limiter *ratelimit.Bucket) *MuxConnRateLimiter {
	c := &MuxConnRateLimiter{
		ExtendedConn: bufio.NewExtendedConn(conn),
		limiter:      limiter,
	}
	c.lastReadTime.Store(time.Now())

	c.readFunc = func(n int64) {
		c.readBytes.Add(n)
		c.tryThrottle(n)
	}

	c.writeFunc = func(n int64) {
		c.writeBytes.Add(n)
		c.tryThrottle(n)
	}

	return c
}

// tryThrottle implements non-blocking throttling
func (c *MuxConnRateLimiter) tryThrottle(n int64) {
	// Try to take tokens, but don't block
	available := c.limiter.TakeAvailable(n)

	// If we consumed more than available, we're over limit
	if available < n {
		deficit := n - available
		// Calculate delay based on deficit
		// But we don't block here - just consume future tokens
		_ = deficit
	}
}

func (c *MuxConnRateLimiter) Read(b []byte) (n int, err error) {
	n, err = c.ExtendedConn.Read(b)
	if n > 0 {
		c.tryThrottle(int64(n))
	}
	return
}

func (c *MuxConnRateLimiter) Write(b []byte) (n int, err error) {
	// For mux connections, we throttle before write
	// But use a softer approach
	if len(b) > 0 {
		c.softWait(int64(len(b)))
	}

	n, err = c.ExtendedConn.Write(b)
	if n > 0 && n != len(b) {
		// Adjust if partial write
		c.tryThrottle(int64(n))
	}
	return
}

// softWait implements a gentle throttling that won't block for too long
func (c *MuxConnRateLimiter) softWait(n int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Maximum wait time: 100ms to avoid long blocks
	const maxWait = 100 * time.Millisecond

	delay := c.limiter.Take(n)
	if delay > maxWait {
		// If delay is too long, just take partial tokens and continue
		// This prevents long blocks that break mux
		time.Sleep(maxWait)
		// Consume tokens for the wait we did
		c.limiter.Take(n / 10) // Take 10% as compromise
	} else if delay > 0 {
		time.Sleep(delay)
	}
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

func (c *MuxConnRateLimiter) UnwrapReader() (io.Reader, []network.CountFunc) {
	return c.ExtendedConn, []network.CountFunc{c.readFunc}
}

func (c *MuxConnRateLimiter) UnwrapWriter() (io.Writer, []network.CountFunc) {
	return c.ExtendedConn, []network.CountFunc{c.writeFunc}
}

func (c *MuxConnRateLimiter) Upstream() any {
	return c.ExtendedConn
}
