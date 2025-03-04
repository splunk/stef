package internal

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jonboulle/clockwork"
	"go.uber.org/zap"
)

type ConnManager struct {
	logger *zap.Logger

	clock clockwork.Clock

	// Number of connections desirable to maintain.
	targetConnCount uint

	// Number of current connections, in all pools or acquired.
	curConnCount atomic.Int64

	// ConnCreator is used to create new connections.
	connCreator ConnCreator

	// Connection pools. curConnCount connections are either in one
	// of these pools, or are acquired temporarily.
	idleConns     chan *ManagedConn // Ready to be acquired.
	flushConns    chan *ManagedConn // Pending to be flushed.
	recreateConns chan *ManagedConn // Pending to be recreated.

	// Period to flush connections.
	flushPeriod time.Duration

	// Period to reconnect connections.
	reconnectPeriod time.Duration

	// Flags to indicate if the goroutines are stopped.
	flusherStopped     bool
	reconnectorStopped bool
	discarderStopped   bool
	// stoppedCond is used to wait until all goroutines are stopped.
	stoppedCond *CancellableCond

	// stopSignal is closed to signal all goroutines to stop.
	stopSignal chan struct{}
}

type ManagedConn struct {
	Conn       Conn
	startTime  time.Time
	lastFlush  time.Time
	needsFlush bool
	isAcquired bool
}

type ConnCreator interface {
	// Create a new connection. May be called concurrently.
	Create(ctx context.Context) (Conn, error)
}

type Conn interface {
	// Close the connection. The connection will be discarded
	// after this call returns.
	Close() error

	// Flush the connection. This is typically to send any buffered data.
	// Will be called periodically (see ConnManager flushPeriod) and
	// before ConnManager.Stop returns.
	Flush() error
}

func NewConnManager(
	logger *zap.Logger,
	creator ConnCreator,
	targetConnCount uint,
	flushPeriod time.Duration,
	reconnectPeriod time.Duration,
) *ConnManager {
	return &ConnManager{
		logger:          logger,
		clock:           clockwork.NewRealClock(),
		connCreator:     creator,
		targetConnCount: targetConnCount,
		idleConns:       make(chan *ManagedConn, targetConnCount),
		flushConns:      make(chan *ManagedConn, targetConnCount),
		recreateConns:   make(chan *ManagedConn, targetConnCount),
		flushPeriod:     flushPeriod,
		reconnectPeriod: reconnectPeriod,
		stoppedCond:     NewCancellableCond(),
		stopSignal:      make(chan struct{}),
	}
}

// Start starts the connection manager. It will start creating targetConnCount
// available connections.
func (c *ConnManager) Start() {
	// Create connections in recreateConns pool. They will be created
	// by recreator.
	for i := uint(0); i < c.targetConnCount; i++ {
		c.recreateConns <- &ManagedConn{}
		c.curConnCount.Add(1)
	}

	go c.flusher()
	go c.reconnector()
	go c.recreator()
}

// Stop stops the connection manager. It will wait until all acquired
// connections are returned. Then it will flush connections that
// are marked as needing to flush, and then will close all connections.
func (c *ConnManager) Stop(ctx context.Context) error {
	// Signal goroutines to stop
	close(c.stopSignal)

	// Wait until all goroutines stop
	err := c.stoppedCond.Wait(
		ctx, func() bool {
			return c.flusherStopped && c.reconnectorStopped && c.discarderStopped
		},
	)
	if err != nil {
		c.logger.Error("Failed to stop connection manager", zap.Error(err))
		return err
	}

	return c.closeAll(ctx)
}

// Close all connections.
func (c *ConnManager) closeAll(ctx context.Context) error {
	// We must close exactly targetConnCount connections in total.
	// All goroutines are stopped at this point, so they won't interfere.
	// All connections are either in the between idleConns, flushConns and recreateConns
	// pools or are acquired and will be returned to one of the pools soon.

	cnt := c.curConnCount.Load()
	c.logger.Debug("closing connections", zap.Int64("count", cnt))

	var errs []error
	for i := int64(0); i < cnt; i++ {
		// Get a connection from one of the connection pools.
		var conn *ManagedConn
		var discarded bool
		select {
		case conn = <-c.recreateConns:
			discarded = true
		case <-ctx.Done():
			return ctx.Err()
		case conn = <-c.idleConns:
		case conn = <-c.flushConns:
		}

		if conn.Conn != nil {
			// Flush if needs a flush and is not discarded.
			if !discarded && conn.needsFlush {
				if err := conn.Conn.Flush(); err != nil {
					c.logger.Debug("Failed to flush connection", zap.Error(err))
					errs = append(errs, err)
					continue
				}
			}

			// And close the connection.
			if err := conn.Conn.Close(); err != nil {
				c.logger.Debug("Failed to close connection", zap.Error(err))
				errs = append(errs, err)
				continue
			}
		}
	}

	// Join all errors (if any)
	return errors.Join(errs...)
}

// Acquire an idle connection for exclusive use.
// Must call Release() or DiscardAndClose() when done.
// Returns an error if the connection is not available til ctx is done
// or if the manager is stopped.
func (c *ConnManager) Acquire(ctx context.Context) (*ManagedConn, error) {
	select {
	case conn := <-c.idleConns:
		if conn.isAcquired {
			panic("connection is not acquired")
		}
		conn.isAcquired = true
		return conn, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.stopSignal:
		return nil, errors.New("connection manager is stopped")
	}
}

// Release returns the previously acquired connection.
// If the connection was last flushed more than flushPeriod ago, it will
// be flushed first otherwise it becomes available for other clients to
// Acquire immediately. Either way the connection will be marked as
// needing a flush.
func (c *ConnManager) Release(conn *ManagedConn) {
	if !conn.isAcquired {
		panic("connection is not acquired")
	}
	conn.isAcquired = false
	conn.needsFlush = true
	if c.clock.Since(conn.lastFlush) >= c.flushPeriod {
		c.flushConns <- conn
		return
	}
	c.idleConns <- conn
}

// DiscardAndClose discards the acquired connection and closes it.
// This is normally used when the connection goes bad in some way
// and should not be reused.
func (c *ConnManager) DiscardAndClose(conn *ManagedConn) {
	if !conn.isAcquired {
		panic("connection is not acquired")
	}
	conn.needsFlush = false
	c.recreateConns <- conn
}

func (c *ConnManager) flusher() {
	defer func() {
		c.stoppedCond.Cond.L.Lock()
		c.flusherStopped = true
		c.stoppedCond.Cond.L.Unlock()
		c.stoppedCond.Cond.Broadcast()
	}()

	for {
		select {
		case <-c.stopSignal:
			return
		case conn := <-c.flushConns:
			if conn.needsFlush {
				conn.needsFlush = false
				if err := conn.Conn.Flush(); err != nil {
					c.logger.Error("Failed to flush connection. Closing connection.", zap.Error(err))
					c.recreateConns <- conn
					continue
				}
				conn.lastFlush = c.clock.Now()
			}
			c.idleConns <- conn
		}
	}
}

// reconnector periodically checks connections and reconnects them if they
// were connected for more than reconnectPeriod. It will stagger the reconnections
// to avoid all connections reconnecting at the same c.clock.
func (c *ConnManager) reconnector() {
	defer func() {
		c.stoppedCond.Cond.L.Lock()
		c.reconnectorStopped = true
		c.stoppedCond.Cond.L.Unlock()
		c.stoppedCond.Cond.Broadcast()
	}()

	// Periodically reconnect idle connections.
	ticker := c.clock.NewTicker(c.reconnectPeriod / time.Duration(c.targetConnCount))
	defer ticker.Stop()
	for {
		select {
		case <-c.stopSignal:
			return
		case <-ticker.Chan():
			// Find an idle connection
			var conn *ManagedConn
			select {
			case <-c.stopSignal:
				return
			case conn = <-c.idleConns:
			}

			// Check if it is time to reconnect.
			if c.clock.Since(conn.startTime) >= c.reconnectPeriod {
				if conn.needsFlush {
					conn.needsFlush = false
					// Flush it first.
					if err := conn.Conn.Flush(); err != nil {
						c.logger.Error("Failed to flush connection. Closing connection.", zap.Error(err))
					}
				}

				// Send it for reconnection.
				c.recreateConns <- conn
			} else {
				// Put it back, too soon to reconnect.
				c.idleConns <- conn
			}
		}
	}
}

// recreator closes connections from recreateConns pool
// and replaces them by new connections.
func (c *ConnManager) recreator() {
	defer func() {
		c.stoppedCond.Cond.L.Lock()
		c.discarderStopped = true
		c.stoppedCond.Cond.L.Unlock()
		c.stoppedCond.Cond.Broadcast()
	}()

	for {
		select {
		case <-c.stopSignal:
			return
		case conn := <-c.recreateConns:
			if c.curConnCount.Add(-1) < 0 {
				panic("negative connection count")
			}
			if conn.Conn != nil {
				// Close the connection.
				if err := conn.Conn.Close(); err != nil {
					c.logger.Error("Failed to close connection", zap.Error(err))
				}
			}
			c.createNewConn()
		}
	}
}

// contextFromStopSignal creates a context that is cancelled when the
// provided channel is closed.
func contextFromStopSignal(stopSignal <-chan struct{}) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-stopSignal
		cancel()
	}()
	return ctx, cancel
}

// createNewConn creates a new connection and adds it to the idleConns pool.
// It will keep trying with backoff until connection creation succeeds or
// until ConnManager is stopped.
func (c *ConnManager) createNewConn() {
	c.logger.Debug("Creating new connection")

	// Create a context that will be cancelled when the manager is stopped.
	ctx, cancel := contextFromStopSignal(c.stopSignal)
	defer cancel()

	// Try to connect, retrying until succeeding, with exponential backoff.
	bo := backoff.NewExponentialBackOff()
	ticker := backoff.NewTicker(bo)
	defer ticker.Stop()

	for {
		// Wait until the next retry or until the manager is stopped.
		select {
		case <-c.stopSignal:
			return
		case <-ticker.C:
		}

		c.logger.Debug("calling Create() connection")
		conn, err := c.connCreator.Create(ctx)
		if err != nil {
			c.logger.Info("Failed to create connection. Will retry.", zap.Error(err))
			continue
		}

		now := c.clock.Now()
		managedConn := &ManagedConn{
			Conn:      conn,
			startTime: now,
			lastFlush: now,
		}
		c.curConnCount.Add(1)
		c.idleConns <- managedConn
		return
	}
}
