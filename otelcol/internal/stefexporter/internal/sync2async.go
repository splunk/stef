package internal

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// Async is a function that executes an operation asynchronously.
// It takes some data and a channel to report the result.
// It returns a DataID, that will be reported in the resultChan when
// the operation completes.
type Async func(
	ctx context.Context,
	data any,
	resultChan ResultChan,
) (DataID, error)

// DataID identifies the result of an asynchronous operation.
// Since Async function can be called repeatedly, the DataID is used to differentiate
// which of the operations started via Async() call has completed when indicated
// by the resultChan.
type DataID uint64

// AsyncResult is the result of an asynchronous operation completion.
type AsyncResult struct {
	// DataID is the ID of the data that was processed. This matches the DataID
	// returned by the Async function.
	DataID DataID

	// Err is the error that occurred during the processing of the data or nil
	// if the operation completed successfully.
	Err error
}

// ResultChan is a channel used to report the result of an asynchronous
// operation completion.
type ResultChan chan AsyncResult

// Sync2Async is an API converter that allows an asynchronous implementation
// of a function (Async) to be called synchronously.
type Sync2Async struct {
	logger             *zap.Logger
	async              Async
	asyncMux           sync.Mutex
	resultChannelsRing chan chan AsyncResult
}

// NewSync2Async creates a new Sync2Async instance.
// concurrency is the number of concurrent async calls that can be in flight.
// If more than concurrency Sync() calls are made, the caller will block until
// one of the async calls completes and returns a result.
func NewSync2Async(logger *zap.Logger, concurrency int, async Async) *Sync2Async {
	s := &Sync2Async{
		logger:             logger,
		async:              async,
		resultChannelsRing: make(chan chan AsyncResult, concurrency),
	}

	for i := 0; i < concurrency; i++ {
		// We need 1 element in the channel to make sure reporting the results via channel is not
		// blocked when the recipient of the channel gave up.
		s.resultChannelsRing <- make(chan AsyncResult, 1)
	}

	return s
}

// Sync performs a synchronous operation. It will trigger the execution of the
// provided Async operation with supplied data and will block until the async
// operation completes (until the resultChan receives a result).
//
// If the number of calls to Sync() exceeds the concurrency limit, the caller will block
// until one of the previous calls completes. This means that even if the number of
// executing Sync() calls exceeds concurrency limit, the number of Async calls
// will never exceed the concurrency limit.
//
// Cancelling the ctx will cause the async operation to be abandoned and error
// to be returned. The ctx is also passed to the Async function.
// If ctx is cancelled before the async operation is started (e.g. due to being
// blocked on concurrency limit) then an error will be returned.
func (s *Sync2Async) Sync(ctx context.Context, data any) error {
	// Choose an resultChan. We are simply doing round-robin here, every Sync()
	var resultChan chan AsyncResult
	select {
	case resultChan = <-s.resultChannelsRing:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() {
		s.resultChannelsRing <- resultChan
	}()

	dataID, err := s.async(ctx, data, resultChan)
	if err != nil {
		return err
	}

	select {
	case result := <-resultChan:
		if result.DataID != dataID {
			// Received ack on the wrong data item. This should normally not happen and indicates a bug somewhere.
			s.logger.Error(
				"Received ack on the wrong data item",
				zap.Uint64("expected", uint64(dataID)),
				zap.Uint64("actual", uint64(result.DataID)),
			)
		}
		return result.Err

	case <-ctx.Done():
		// Abandon the ack channel that was given to async() because we don't know when/if
		// that channel will fire. Just allocate a new channel. The new resultChan will
		// be returned to resultChannelsRing when this func returns.
		// If after this ack() is called, it will push dataID on an abandoned channel
		// and will have no effect.
		resultChan = make(chan AsyncResult)
		return ctx.Err()
	}
}
