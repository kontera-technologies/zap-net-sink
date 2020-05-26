package zap_net_sink

import (
	"fmt"
	"net"
	"net/url"

	"go.uber.org/zap"
)

// Register udp and tcp urls with zap.
func init() {
	if err := zap.RegisterSink("udp", NewUDPSink); err != nil {
		panic(err)
	}

	if err := zap.RegisterSink("tcp", NewTCPSink); err != nil {
		panic(err)
	}
}

// NewUDPSink creates a zap sink to the given url.
func NewUDPSink(url *url.URL) (zap.Sink, error) {
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%s", url.Hostname(), url.Port()))
	if err != nil {
		return nil, fmt.Errorf("failed to setup a UDP sink - %w", err)
	}

	return &WriteSyncer{conn}, nil
}

// NewTCPSink creates a zap sink to the given url.
func NewTCPSink(url *url.URL) (zap.Sink, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", url.Hostname(), url.Port()))
	if err != nil {
		return nil, fmt.Errorf("failed to setup a TCP sink - %w", err)
	}

	return &WriteSyncer{conn}, nil
}

type WriteSyncer struct {
	conn net.Conn
}

func (z *WriteSyncer) Close() error {
	return z.conn.Close()
}

func (z *WriteSyncer) Write(p []byte) (n int, err error) {
	return z.conn.Write(p)
}

func (z *WriteSyncer) Sync() error {
	return nil
}
