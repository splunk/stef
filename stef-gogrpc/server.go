package stefgrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/splunk/stef/stef-go/schema"

	"github.com/splunk/stef/stef-gogrpc/internal"
	"github.com/splunk/stef/stef-gogrpc/stef_proto"
	"github.com/splunk/stef/stef-gogrpc/types"
)

type grpcMsgSource interface {
	// RecvMsg receives a TEF/gRPC message containing TEF bytes.
	RecvMsg() (tefBytes []byte, isEndOfChunk bool, err error)
}

type GrpcReader interface {
	Read(p []byte) (n int, err error)
	Stats() GrpcReaderStats
}

type GrpcReaderStats struct {
	MessagesReceived uint64
	BytesReceived    uint64
}

type chunkAssembler struct {
	source    grpcMsgSource
	buf       []byte
	readIndex int

	statsMux sync.RWMutex
	stats    GrpcReaderStats
}

var _ io.Reader = (*chunkAssembler)(nil)

func (g *chunkAssembler) recvMsg() (chunkBytes []byte, err error) {
	var chunkBuf []byte
	for {
		bytes, isEndOfChunk, err := g.source.RecvMsg()
		if err != nil {
			return nil, err
		}
		if chunkBuf == nil {
			chunkBuf = bytes
		} else {
			chunkBuf = append(chunkBuf, bytes...)
		}

		if isEndOfChunk {
			// These bytes are ending a chunk. Return what is accumulated.
			g.statsMux.Lock()
			g.stats.MessagesReceived++
			g.stats.BytesReceived += uint64(len(chunkBuf))
			g.statsMux.Unlock()
			return chunkBuf, nil
		}
	}
}

func (g *chunkAssembler) Read(p []byte) (n int, err error) {
	if g.readIndex >= len(g.buf) {
		data, err := g.recvMsg()
		if err != nil {
			return 0, err
		}
		g.buf = data
		g.readIndex = 0
	}
	n = copy(p, g.buf[g.readIndex:])
	g.readIndex += n
	return n, nil
}

func (g *chunkAssembler) Stats() GrpcReaderStats {
	g.statsMux.RLock()
	defer g.statsMux.RUnlock()
	return g.stats
}

func newChunkAssembler(source grpcMsgSource) GrpcReader {
	return &chunkAssembler{source: source}
}

type grpcChunkSource struct {
	serverStream     stef_proto.STEFDestination_StreamServer
	message          stef_proto.STEFServerMessage
	response         stef_proto.STEFDataResponse
	messagesReceived uint64
}

func newGrpcChunkSource(serverStream stef_proto.STEFDestination_StreamServer) *grpcChunkSource {
	s := &grpcChunkSource{
		serverStream: serverStream,
	}
	s.message = stef_proto.STEFServerMessage{
		Message: &stef_proto.STEFServerMessage_Response{Response: &s.response},
	}
	return s
}

func (r *grpcChunkSource) RecvMsg() (tefBytes []byte, isEndOfChunk bool, err error) {
	response, err := r.serverStream.Recv()
	if err != nil {
		return nil, false, err
	}
	r.messagesReceived++
	return response.StefBytes, response.IsEndOfChunk, nil
}

func (r *grpcChunkSource) AckRecordId(recordId uint64) error {
	// Acknowledge receipt.
	r.response.AckRecordId = recordId
	return r.serverStream.Send(&r.message)
}

type StreamServer struct {
	stef_proto.UnimplementedSTEFDestinationServer

	logger       types.Logger
	serverSchema *schema.Schema
	maxDictBytes uint64
	onStream     func(reader GrpcReader, ackFunc func(sequenceId uint64) error) error
}

var _ stef_proto.STEFDestinationServer = (*StreamServer)(nil)

type ServerSettings struct {
	Logger       types.Logger
	ServerSchema *schema.Schema
	MaxDictBytes uint64
	OnStream     func(reader GrpcReader, ackFunc func(sequenceId uint64) error) error
}

func NewStreamServer(settings ServerSettings) *StreamServer {
	if settings.Logger == nil {
		settings.Logger = internal.NopLogger{}
	}
	//settings.ServerSchema.Minify()
	return &StreamServer{
		logger:       settings.Logger,
		serverSchema: settings.ServerSchema,
		maxDictBytes: settings.MaxDictBytes,
		onStream:     settings.OnStream,
	}
}

func (s *StreamServer) Stream(server stef_proto.STEFDestination_StreamServer) error {
	schemaJSON, err := json.Marshal(s.serverSchema)
	if err != nil {
		return fmt.Errorf("could not marshal server schema: %w", err)
	}

	// Send capabilities message to the client.
	message := stef_proto.STEFServerMessage{
		Message: &stef_proto.STEFServerMessage_Capabilities{
			Capabilities: &stef_proto.STEFDestinationCapabilities{
				DictionaryLimits: &stef_proto.STEFDictionaryLimits{MaxDictBytes: s.maxDictBytes},
				SchemaJson:       schemaJSON,
			},
		},
	}
	err = server.Send(&message)
	if err != nil {
		s.logger.Errorf(context.Background(), "cannot send message to the client: %v", err)
		return err
	}

	grpcStream := newGrpcChunkSource(server)
	reader := newChunkAssembler(grpcStream)
	return s.onStream(reader, grpcStream.AckRecordId)
}
