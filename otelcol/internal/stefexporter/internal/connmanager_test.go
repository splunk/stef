package internal

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type mockConnCreator struct {
	conns    []*mockConn
	connsMux sync.Mutex
}

type mockConn struct {
	closed  bool
	flushed bool
	mux     sync.RWMutex
}

func (m *mockConn) Close() error {
	m.mux.Lock()
	m.closed = true
	m.mux.Unlock()
	return nil
}

func (m *mockConn) Flush() error {
	m.mux.Lock()
	m.flushed = true
	m.mux.Unlock()
	return nil
}

func (m *mockConn) Flushed() bool {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.flushed
}

func (m *mockConn) Closed() bool {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.closed
}

func (m *mockConnCreator) Create(_ context.Context) (Conn, error) {
	conn := &mockConn{}
	m.connsMux.Lock()
	m.conns = append(m.conns, conn)
	m.connsMux.Unlock()
	return conn, nil
}

const connCount = 4

const flushPeriod = time.Hour
const reconnectPeriod = 2 * time.Hour

func withCM(t *testing.T, f func(cm *ConnManager)) *mockConnCreator {
	connCreator := &mockConnCreator{}

	cm := NewConnManager(
		zap.NewNop(),
		connCreator,
		connCount,
		flushPeriod,
		reconnectPeriod,
	)

	cm.Start()

	f(cm)

	assert.NoError(t, cm.Stop(context.Background()))

	for _, conn := range connCreator.conns {
		require.True(t, conn.closed)
	}

	return connCreator
}

func withFakeTimeCM(t *testing.T, f func(cm *ConnManager)) *mockConnCreator {
	connCreator := &mockConnCreator{}

	cm := NewConnManager(
		zap.NewNop(),
		connCreator,
		connCount,
		flushPeriod,
		reconnectPeriod,
	)
	cm.clock = clockwork.NewFakeClock()

	cm.Start()

	f(cm)

	assert.NoError(t, cm.Stop(context.Background()))

	for _, conn := range connCreator.conns {
		require.True(t, conn.closed)
	}

	return connCreator
}

func TestConnManagerStartStop(t *testing.T) {
	withCM(t, func(cm *ConnManager) {})
}

func TestConnManagerAcquireRelease(t *testing.T) {
	withCM(
		t, func(cm *ConnManager) {
			conn, err := cm.Acquire(context.Background())
			require.NoError(t, err)
			require.NotNil(t, conn)
			cm.Release(conn)
		},
	)
}

func TestConnManagerAcquireReleaseMany(t *testing.T) {
	var conns []*ManagedConn
	connCreator := withCM(
		t, func(cm *ConnManager) {
			for i := 0; i < connCount; i++ {
				conn, err := cm.Acquire(context.Background())
				require.NoError(t, err)
				require.NotNil(t, conn)
				conns = append(conns, conn)
			}
			for i := 0; i < connCount; i++ {
				cm.Release(conns[i])
			}
		},
	)
	require.EqualValues(t, connCount, len(connCreator.conns))
	for _, conn := range conns {
		require.True(t, conn.Conn.(*mockConn).flushed)
	}
}

func TestConnManagerAcquireDiscardAcquire(t *testing.T) {
	var conns []*ManagedConn
	connCreator := withCM(
		t, func(cm *ConnManager) {
			conn, err := cm.Acquire(context.Background())
			require.NoError(t, err)

			// Discard the connection. This should restart in a replacement
			// connection creation.
			cm.DiscardAndClose(conn)

			// Make sure we ask for maximum number of possible connections
			// so that we guarantee the discarded connection is replaced.
			for i := 0; i < connCount; i++ {
				conn, err := cm.Acquire(context.Background())
				require.NoError(t, err)
				require.NotNil(t, conn)
				conns = append(conns, conn)
			}
			for i := 0; i < connCount; i++ {
				cm.Release(conns[i])
			}
		},
	)

	// Make sure one more connection is created because we discarded one.
	require.EqualValues(t, connCount+1, len(connCreator.conns))
}

func TestConnManagerAcquireReleaseConcurrent(t *testing.T) {
	withCM(
		t, func(cm *ConnManager) {
			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					conn, err := cm.Acquire(context.Background())
					if err != nil {
						return
					}
					cm.Release(conn)
				}()
			}
			wg.Wait()
		},
	)
}

func TestConnManagerAcquireDiscard(t *testing.T) {
	withCM(
		t, func(cm *ConnManager) {
			conn, err := cm.Acquire(context.Background())
			require.NoError(t, err)
			cm.DiscardAndClose(conn)
		},
	)
}

func TestConnManagerAcquireDiscardConcurrent(t *testing.T) {
	connCreator := withCM(
		t, func(cm *ConnManager) {
			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					conn, err := cm.Acquire(context.Background())
					if err != nil {
						return
					}
					cm.DiscardAndClose(conn)
				}()
			}
			wg.Wait()
		},
	)
	for _, conn := range connCreator.conns {
		require.False(t, conn.flushed)
	}
}

func TestConnManagerFlush(t *testing.T) {
	withFakeTimeCM(
		t, func(cm *ConnManager) {
			conn, err := cm.Acquire(context.Background())
			require.NoError(t, err)

			// Make sure the flush is not done yet.
			assert.False(t, conn.Conn.(*mockConn).Flushed())

			// Advance the clock so that Release() trigger flush.
			cm.clock.(*clockwork.FakeClock).Advance(flushPeriod)
			cm.Release(conn)

			// Make sure the flush is done.
			assert.Eventually(
				t, func() bool {
					return conn.Conn.(*mockConn).Flushed()
				},
				5*time.Second,
				time.Millisecond*10,
			)
		},
	)
}

//func TestConnManagerReconnector(t *testing.T) {
//	var conns []*ManagedConn
//	withFakeTimeCM(
//		t, func(cm *ConnManager) {
//			for i := 0; i < connCount; i++ {
//				conn, err := cm.Acquire(context.Background())
//				require.NoError(t, err)
//				require.NotNil(t, conn)
//				conns = append(conns, conn)
//			}
//			for i := 0; i < connCount; i++ {
//				cm.Release(conns[i])
//			}
//
//			// Advance the clock to trigger reconnects.
//			cm.clock.(*clockwork.FakeClock).Advance(reconnectPeriod)
//			for i := 0; i < connCount; i++ {
//				cm.clock.(*clockwork.FakeClock).Advance(reconnectPeriod)
//				time.Sleep(500 * time.Millisecond)
//			}
//
//			for _, conn := range conns {
//				assert.Eventually(
//					t, func() bool {
//						return conn.Conn.(*mockConn).Closed()
//					}, 5*time.Second, time.Millisecond*10,
//				)
//			}
//		},
//	)
//}
