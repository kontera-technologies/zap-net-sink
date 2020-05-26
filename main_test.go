package zap_net_sink_test

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	_ "kontera-technologies/zap-net-sink"
)

func TestUDPSink(t *testing.T) {
	server := setupUDP(t, 1234)
	serverOutC := decToChan(t, json.NewDecoder(server))

	writer, cleanup, err := zap.Open("udp://127.0.0.1:1234")
	fatalIfErr(t, err)
	t.Cleanup(cleanup)

	logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), writer, zap.DebugLevel)).Sugar()

	logger.Warnw("hello", "from", "warn")
	var actual map[string]interface{}
	select {
	case actual = <-serverOutC:
		actualEqualsExpected(t, actual, map[string]interface{}{"msg": "hello", "level":"warn", "from": "warn", "ts": actual["ts"]})
	case <-time.After(time.Millisecond):
		t.Fatal("expected message within 1ms")
	}

	logger.Infow("hello", "from", "info")
	select {
	case actual = <-serverOutC:
		actualEqualsExpected(t, actual, map[string]interface{}{"msg": "hello", "level":"info", "from": "info", "ts": actual["ts"]})
	case <-time.After(time.Millisecond):
		t.Fatal("expected message within 1ms")
	}
}

func TestTCPSink(t *testing.T) {
	server := setupTCP(t, 1234)
	var serverOutC <-chan map[string]interface{}
	var serverWG sync.WaitGroup
	serverWG.Add(1)
	go func() {
		conn, err := server.Accept()
		fatalIfErr(t, err)
		serverWG.Done()
		serverOutC = decToChan(t, json.NewDecoder(conn))
	}()

	writer, cleanup, err := zap.Open("tcp://127.0.0.1:1234")
	fatalIfErr(t, err)
	t.Cleanup(cleanup)

	logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), writer, zap.DebugLevel)).Sugar()

	serverWG.Wait()

	logger.Warnw("hello", "from", "warn")
	var actual map[string]interface{}
	select {
	case actual = <-serverOutC:
		actualEqualsExpected(t, actual, map[string]interface{}{"msg": "hello", "level":"warn", "from": "warn", "ts": actual["ts"]})
	case <-time.After(time.Millisecond):
		t.Fatal("expected message within 1ms")
	}

	logger.Infow("hello", "from", "info")
	select {
	case actual = <-serverOutC:
		actualEqualsExpected(t, actual, map[string]interface{}{"msg": "hello", "level":"info", "from": "info", "ts": actual["ts"]})
	case <-time.After(time.Millisecond):
		t.Fatal("expected message within 1ms")
	}
}

func actualEqualsExpected(tb testing.TB, actual, expected interface{}) {
	tb.Helper()
	if !reflect.DeepEqual(actual, expected) {
		tb.Errorf(`Actual doesn't equal expected:
Actual:   %#v
Expected: %#v`, actual, expected)
	}
}

func decToChan(tb testing.TB, dec *json.Decoder) <-chan map[string]interface{} {
	tb.Helper()
	serverOutC := make(chan map[string]interface{}, 1)
	go func() {
		for {
			var v map[string]interface{}
			if err := dec.Decode(&v); err != nil {
				return
			}
			serverOutC <- v
		}
	}()
	return serverOutC
}

func setupUDP(tb testing.TB, port uint) net.Conn {
	laddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
	fatalIfErr(tb, err)
	logServer, err := net.ListenUDP("udp", laddr)
	fatalIfErr(tb, err)
	tb.Cleanup(func() { _ = logServer.Close() })
	return logServer
}

func setupTCP(tb testing.TB, port uint) net.Listener {
	laddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	fatalIfErr(tb, err)
	listener, err := net.ListenTCP("tcp", laddr)
	fatalIfErr(tb, err)
	tb.Cleanup(func() { _ = listener.Close() })
	return listener
}

func fatalIfErr(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatal(err)
	}
}
