package stefreceiver

import (
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	stefgrpc "github.com/splunk/stef/go/grpc"
	"github.com/splunk/stef/go/grpc/stef_proto"
)

type responder struct {
	logger *zap.Logger

	// Stream to send responses to.
	stream stefgrpc.STEFStream

	// Channel to stop the responder goroutine.
	stopCh chan struct{}

	nextResponse    stef_proto.STEFDataResponse
	nextResponseMux sync.RWMutex

	// The next ack ID to send to the client.
	NextAckID atomic.Uint64

	// An error that occurred while sending responses.
	LastError atomic.Value

	// Channel to receive bad data information from.
	BadDataCh chan BadData
}

func newResponder(logger *zap.Logger, stream stefgrpc.STEFStream) *responder {
	return &responder{
		logger:    logger,
		stream:    stream,
		stopCh:    make(chan struct{}),
		BadDataCh: make(chan BadData, 10),
	}
}

func (r *responder) stop() {
	close(r.stopCh)
}

func (r *responder) run() {
	t := time.NewTicker(10 * time.Millisecond)
	var acksSent uint64
	var lastAckedID uint64

	// Preallocate to avoid allocations in the loop.
	badDataResponse := &stef_proto.STEFDataResponse{
		BadDataRecordIdRanges: []*stef_proto.STEFIDRange{{}},
	}
	ackResponse := &stef_proto.STEFDataResponse{}

	for {
		select {
		case badData := <-r.BadDataCh:
			r.composeBadDataResponse(badDataResponse, badData)
			if err := r.stream.SendDataResponse(badDataResponse); err != nil {
				r.logger.Error("Error acking STEF gRPC connection", zap.Error(err))
				r.LastError.Store(err)
				return
			}
			acksSent++
			lastAckedID = badDataResponse.AckRecordId

		case <-t.C:
			readRecordID := r.NextAckID.Load()
			if readRecordID > lastAckedID {
				lastAckedID = readRecordID
				ackResponse.AckRecordId = lastAckedID
				if err := r.stream.SendDataResponse(ackResponse); err != nil {
					r.logger.Error("Error acking STEF gRPC connection", zap.Error(err))
					r.LastError.Store(err)
					return
				}
				acksSent++
			}
			// TODO: get stats from grcpReader and record then in obsReport.

		case <-r.stopCh:
			return
		}
	}
}

func (r *responder) composeBadDataResponse(response *stef_proto.STEFDataResponse, badData BadData) {
	response.AckRecordId = badData.toID

	// First bad data range.
	response.BadDataRecordIdRanges = response.BadDataRecordIdRanges[:1]
	response.BadDataRecordIdRanges[0].FromId = badData.fromID
	response.BadDataRecordIdRanges[0].ToId = badData.toID

	// See if there is more bad data we can report in the same response.
	for {
		select {
		case moreBadData := <-r.BadDataCh:
			// Add a range.
			response.BadDataRecordIdRanges = append(
				response.BadDataRecordIdRanges, &stef_proto.STEFIDRange{
					FromId: moreBadData.fromID,
					ToId:   moreBadData.toID,
				},
			)

			// Use the last ID value for AckRecordID
			if response.AckRecordId < moreBadData.toID {
				response.AckRecordId = moreBadData.toID
			}
		default:
			return
		}
	}
}
