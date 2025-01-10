package stefgrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/splunk/stef/stef-go/pkg"
	"github.com/splunk/stef/stef-go/schema"

	"github.com/splunk/stef/stef-gogrpc/internal"
	"github.com/splunk/stef/stef-gogrpc/stef_proto"
	"github.com/splunk/stef/stef-gogrpc/types"
)

type ClientCallbacks struct {
	// OnDisconnect is called when the stream is disconnected.
	OnDisconnect func(err error)

	// Callback when an ack is received from the server.
	// Return error will result in disconnecting the stream.
	OnAck func(ackId uint64) error
}

type Client struct {
	grpcClient   stef_proto.STEFDestinationClient
	stream       stef_proto.STEFDestination_StreamClient
	callbacks    ClientCallbacks
	clientSchema *schema.Schema
	logger       types.Logger

	// Running state
	waitCh     chan struct{}
	cancelFunc context.CancelFunc
}

type grpcWriter struct {
	stream  stef_proto.STEFDestination_StreamClient
	request stef_proto.STEFClientMessage
	onAck   func(ackId uint64)
}

func (w *grpcWriter) WriteChunk(header []byte, content []byte) error {
	w.request.StefBytes = w.request.StefBytes[:0]
	w.request.StefBytes = append(w.request.StefBytes, header...)
	w.request.StefBytes = append(w.request.StefBytes, content...)
	w.request.IsEndOfChunk = true

	// TODO: split the chunk into multiple messages if it is too big to fit in one gRPC message.

	return w.stream.Send(&w.request)
}

var ErrServerInvalidResponse = errors.New("invalid server response")

type ClientSettings struct {
	Logger types.Logger
	// gRPC stream to send data over.
	GrpcClient   stef_proto.STEFDestinationClient
	ClientSchema *schema.Schema
	Callbacks    ClientCallbacks
}

func NewClient(settings ClientSettings) *Client {
	if settings.Logger == nil {
		settings.Logger = internal.NopLogger{}
	}

	if settings.Callbacks.OnDisconnect == nil {
		settings.Callbacks.OnDisconnect = func(err error) {}
	}

	client := &Client{
		grpcClient:   settings.GrpcClient,
		callbacks:    settings.Callbacks,
		clientSchema: settings.ClientSchema,
		logger:       settings.Logger,
		waitCh:       make(chan struct{}),
	}

	return client
}

func (c *Client) Connect(ctx context.Context) (pkg.ChunkWriter, pkg.WriterOptions, error) {
	opts := pkg.WriterOptions{
		FrameRestartFlags: pkg.RestartDictionaries,
	}

	ctx, cancelFunc := context.WithCancel(ctx)
	c.cancelFunc = cancelFunc

	stream, err := c.grpcClient.Stream(ctx)
	if err != nil {
		return nil, opts, fmt.Errorf("failed to gRPC stream: %w", err)
	}

	c.stream = stream

	isError := true
	closeOnErr := func() {
		if isError {
			if err := stream.CloseSend(); err != nil {
				c.logger.Debugf(ctx, "CloseSend failed: %v", err)
			}
		}
	}
	defer closeOnErr()

	// The server must send capabilities message.
	message, err := stream.Recv()
	if err != nil {
		return nil, opts, fmt.Errorf("failed to receive from server: %w", err)
	}

	capabilities, ok := message.Message.(*stef_proto.STEFServerMessage_Capabilities)
	if !ok || capabilities == nil || capabilities.Capabilities == nil {
		return nil, opts, ErrServerInvalidResponse
	}

	// Apply dictionary limits.
	if capabilities.Capabilities.DictionaryLimits != nil {
		opts.MaxTotalDictSize = uint(capabilities.Capabilities.DictionaryLimits.MaxDictBytes)
	}

	// Unmarshal server schema.
	var serverSchema schema.Schema
	err = json.Unmarshal(capabilities.Capabilities.SchemaJson, &serverSchema)
	if err != nil {
		return nil, opts, fmt.Errorf("failed to unmarshal capabilities schema: %w", err)
	}

	// Check if server schema is backward compatible with client schema.
	compatibility, err := serverSchema.Compatible(c.clientSchema)
	switch compatibility {
	case schema.CompatibilityExact:
		// Schemas match exactly, nothing else is needed, can start sending data.

	case schema.CompatibilitySuperset:
		// ServerStream schema is superset of client schema. The client MUST specify its schema
		// in the TEF header.
		opts.IncludeDescriptor = true
		opts.Schema = c.clientSchema

	case schema.CompatibilityIncompatible:
		// It is neither exact match nor is server schema a superset, but server schema maybe subset.
		// Check the opposite direction: if client schema is backward compatible with server schema.
		compatibility, err = serverSchema.Compatible(c.clientSchema)

		if err != nil || compatibility == schema.CompatibilityIncompatible {
			return nil, opts, fmt.Errorf("client and server schemas are incompatble: %w", err)
		}

		if compatibility == schema.CompatibilitySuperset {
			// Client schema is superset of server schema. The client MUST downgrade its schema.
			opts.IncludeDescriptor = true
			opts.Schema = c.clientSchema
		}
	}

	isError = false

	writer := &grpcWriter{
		stream: stream,
	}
	go c.receive()

	return writer, opts, nil
}

func (c *Client) Disconnect(ctx context.Context) error {
	// This will cancel and close the stream and terminate receive() method.
	c.cancelFunc()

	// Wait until receive() ends.
	select {
	case <-c.waitCh:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (c *Client) receive() {
	defer close(c.waitCh)

	for {
		resp, err := c.stream.Recv()
		if err != nil {
			if err == io.EOF {
				c.callbacks.OnDisconnect(nil)
				return
			}
			c.logger.Errorf(context.Background(), "Error receiving acks: %v", err)
			c.callbacks.OnDisconnect(err)
			return
		}

		err = c.callbacks.OnAck(resp.Message.(*stef_proto.STEFServerMessage_Response).Response.AckRecordId)
		if err != nil {
			c.callbacks.OnDisconnect(err)
			return
		}
	}
}
