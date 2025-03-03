package internal

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type DataID uint64
type AckChan chan DataID

type Async func(
	ctx context.Context,
	data any,
	ackChan AckChan,
) (DataID, error)

type Sync2Async struct {
	logger          *zap.Logger
	async           Async
	asyncMux        sync.Mutex
	ackChannelsRing chan chan DataID
}

func NewSync2Async(logger *zap.Logger, concurrency int, async Async) *Sync2Async {
	s := &Sync2Async{
		logger:          logger,
		async:           async,
		ackChannelsRing: make(chan chan DataID, concurrency),
	}

	for i := 0; i < concurrency; i++ {
		s.ackChannelsRing <- make(chan DataID)
	}

	return s
}

func (s *Sync2Async) Sync(ctx context.Context, data any) error {
	// Choose an ackChannel. We are simply doing round-robin here, every Sync()
	var ackChannel chan DataID
	select {
	case ackChannel = <-s.ackChannelsRing:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() {
		s.ackChannelsRing <- ackChannel
	}()

	dataID, err := s.async(ctx, data, ackChannel)
	if err != nil {
		return err
	}
	//fmt.Printf("Wait Ack %04d\n", dataID)

	select {
	case id := <-ackChannel:
		//fmt.Printf("Got  Ack %04d\n", id)
		if id != dataID {
			// Received ack on the wrong data item. This should normally not happen and indicates a bug somewhere.
			s.logger.Error(
				"Received ack on the wrong data item",
				zap.Uint64("expected", uint64(dataID)),
				zap.Uint64("actual", uint64(id)),
			)
		}
	case <-ctx.Done():
		// Abandon the ack channel that was given to async() because we don't know when/if
		// that channel will fire. Just allocate a new channel. The new ackChannel will
		// be returned to ackChannelsRing when this func returns.
		// If after this ack() is called, it will push dataID on an abandoned channel
		// and will have no effect.
		ackChannel = make(chan DataID)
		return ctx.Err()
	}

	return nil
}
