package mux

import (
	"errors"
	"github.com/integration-system/isp-lib/atomic"
	"net"
	"sync"
)

var (
	ErrListenerClosed = errors.New("mux: listener closed")
	ErrNotMatched     = errors.New("mux: connection not matched by a matcher")
)

type Mux interface {
	Match(Matcher) net.Listener
	Serve() error
	OnError(func(error))
	Close() error
}

type matchersListener struct {
	matcher  Matcher
	listener *muxListener
}

type mux struct {
	listener   net.Listener
	ml         []matchersListener
	connBufLen int
	doneCh     chan struct{}
	errHandler func(error)
}

func (m *mux) Match(matcher Matcher) net.Listener {
	ml := newMuxListener(m.listener, m.connBufLen)
	m.ml = append(m.ml, matchersListener{matcher: matcher, listener: ml})
	return ml
}

func (m *mux) Serve() error {
	var wg sync.WaitGroup

	defer func() {
		close(m.doneCh)
		wg.Wait()

		for _, sl := range m.ml {
			_ = sl.listener.Close()
			for c := range sl.listener.connCh {
				_ = c.Close()
			}
			close(sl.listener.connCh)
		}
	}()

	for {
		c, err := m.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				continue
			}
			return err
		}

		wg.Add(1)
		go m.serve(c, &wg)
	}
}

func (m *mux) Close() error {
	return m.listener.Close()
}

func (m *mux) OnError(f func(error)) {
	m.errHandler = f
}

func (m *mux) serve(c net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	buf := getBuffer()
	n, err := c.Read(buf)
	data := buf[:n]
	for _, matcher := range m.ml {
		matched := matcher.matcher(data)
		if matched {
			if matcher.listener.closed.Get() {
				break
			}
			muc := newMuxConn(c, data, err)
			select {
			case matcher.listener.connCh <- muc:
			case <-m.doneCh:
				putBuffer(buf)
				_ = c.Close()
			}
			return
		}
	}

	putBuffer(buf)
	_ = c.Close()
	if m.errHandler != nil {
		m.errHandler(ErrNotMatched)
	}
}

func (m *mux) handleErr(err error) bool {
	if ne, ok := err.(net.Error); ok {
		return ne.Temporary()
	}
	return false
}

func New(l net.Listener) Mux {
	return &mux{
		listener:   l,
		connBufLen: 1024,
		doneCh:     make(chan struct{}),
	}
}

type muxListener struct {
	root   net.Listener
	connCh chan net.Conn
	closed *atomic.AtomicBool
}

func (l *muxListener) Accept() (net.Conn, error) {
	if l.closed.Get() {
		return nil, ErrListenerClosed
	}
	c, ok := <-l.connCh
	if !ok || l.closed.Get() {
		return nil, ErrListenerClosed
	}
	return c, nil
}

// Close stops listening on address.
// Already Accepted connections are not closed.
func (l *muxListener) Close() error {
	l.closed.Set(true)
	return nil
}

func (l *muxListener) Addr() net.Addr {
	return l.root.Addr()
}

func newMuxListener(root net.Listener, connBufLen int) *muxListener {
	return &muxListener{
		root:   root,
		connCh: make(chan net.Conn, connBufLen),
		closed: atomic.NewAtomicBool(false),
	}
}

type muxConn struct {
	net.Conn
	buf     []byte
	lastErr error
}

func newMuxConn(c net.Conn, buf []byte, lastErr error) *muxConn {
	return &muxConn{
		Conn:    c,
		buf:     buf,
		lastErr: lastErr,
	}
}

func (m *muxConn) Read(p []byte) (int, error) {
	if m.buf != nil {
		n := copy(p, m.buf)
		if len(m.buf) == n {
			putBuffer(m.buf)
			m.buf = nil
		} else {
			m.buf = m.buf[n:]
		}
		return n, m.lastErr
	}
	return m.Conn.Read(p)
}
