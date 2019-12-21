package net

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

// TLSListener is a TLS listener that provides accept with context.
type TLSListener struct {
	tcp       *net.TCPListener
	listener  net.Listener
	heartBeat time.Duration
	closed    uint32
}

// NewTLSListener creates tcp listener.
// Known networks are "tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only).
func NewTLSListener(network string, addr string, cfg *tls.Config, heartBeat time.Duration) (*TLSListener, error) {
	tcp, err := newNetTCPListen(network, addr)
	if err != nil {
		return nil, fmt.Errorf("cannot create new tls listener: %v", err)
	}
	tls := tls.NewListener(tcp, cfg)
	return &TLSListener{
		tcp:       tcp,
		listener:  tls,
		heartBeat: heartBeat,
	}, nil
}

// AcceptContext waits with context for a generic Conn.
func (l *TLSListener) AcceptWithContext(ctx context.Context) (net.Conn, error) {
	for {
		if atomic.LoadUint32(&l.closed) == 1 {
			return nil, ErrServerClosed
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		err := l.SetDeadline(time.Now().Add(l.heartBeat))
		if err != nil {
			return nil, fmt.Errorf("cannot accept connections: %v", err)
		}
		rw, err := l.listener.Accept()
		if err != nil {
			if isTemporary(err) {
				continue
			}
			return nil, fmt.Errorf("cannot accept connections: %v", err)
		}
		return rw, nil
	}
}

// SetDeadline sets deadline for accept operation.
func (l *TLSListener) SetDeadline(t time.Time) error {
	return l.tcp.SetDeadline(t)
}

// Accept waits for a generic Conn.
func (l *TLSListener) Accept() (net.Conn, error) {
	return l.AcceptWithContext(context.Background())
}

// Close closes the connection.
func (l *TLSListener) Close() error {
	if !atomic.CompareAndSwapUint32(&l.closed, 0, 1) {
		return nil
	}
	return l.listener.Close()
}

// Addr represents a network end point address.
func (l *TLSListener) Addr() net.Addr {
	return l.listener.Addr()
}
