package internal

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type DataID uint64

type AsyncResult struct {
	DataID DataID
	Err    error
}
type ResultChan chan AsyncResult

type Async func(
	ctx context.Context,
	data any,
	resultChan ResultChan,
) (DataID, error)

type Sync2Async struct {
	logger             *zap.Logger
	async              Async
	asyncMux           sync.Mutex
	resultChannelsRing chan chan AsyncResult
}

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
	//fmt.Printf("Wait Ack %04d\n", dataID)

	select {
	case result := <-resultChan:
		//fmt.Printf("Got  Ack %04d\n", id)
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
