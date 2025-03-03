package internal

import (
	"context"
	"sync"
	"testing"
	"time"

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
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockConn) Flush() error {
	m.flushed = true
	return nil
}

func (m *mockConnCreator) Create(_ context.Context) (Conn, error) {
	conn := &mockConn{}
	m.connsMux.Lock()
	m.conns = append(m.conns, conn)
	m.connsMux.Unlock()
	return conn, nil
}

const connCount = 4

func withCM(t *testing.T, f func(cm *ConnManager)) *mockConnCreator {
	connCreator := &mockConnCreator{}

	cm := NewConnManager(
		zap.NewNop(),
		connCreator,
		connCount,
		time.Hour,
		time.Hour,
	)

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
