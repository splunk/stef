package internal

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSync2Async(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	var globalID DataID = 0
	pendingAcks := make(
		chan struct {
			dataID   DataID
			ackToken AckChan
		},
	)

	var asyncMux sync.Mutex

	async := func(
		ctx context.Context,
		data any,
		ackToken AckChan,
	) (DataID, error) {
		asyncMux.Lock()
		globalID++
		ackID := globalID
		asyncMux.Unlock()

		pendingAcks <- struct {
			dataID   DataID
			ackToken AckChan
		}{
			dataID:   ackID,
			ackToken: ackToken,
		}

		return ackID, nil
	}
	go func() {
		for pendingAck := range pendingAcks {
			pendingAck.ackToken <- pendingAck.dataID
		}
	}()

	const syncProducers = 1000
	s2a := NewSync2Async(logger, syncProducers, async)

	ctx := context.Background()
	var wg sync.WaitGroup
	const countPerProducer = 1000
	const totalCount = syncProducers * countPerProducer
	for i := 0; i < syncProducers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data := i
			for i := 0; i < countPerProducer; i++ {
				err := s2a.Sync(ctx, data)
				if err != nil {
					require.NoError(t, err)
				}
			}
		}()
	}
	wg.Wait()
	close(pendingAcks)

	require.EqualValues(t, totalCount, globalID)
}
