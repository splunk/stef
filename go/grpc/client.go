package stefgrpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/splunk/stef/go/pkg"
	"github.com/splunk/stef/go/pkg/schema"

	"github.com/splunk/stef/go/grpc/internal"
	"github.com/splunk/stef/go/grpc/stef_proto"
	"github.com/splunk/stef/go/grpc/types"
)

type ClientCallbacks struct {
	// OnDisconnect is called when the stream is disconnected.
	OnDisconnect func(err error)

	// Callback when an ack is received from the server.
	// Return error will result in disconnecting the stream.
	OnAck func(ackId uint64) error
}

// Client is a client for communicating over STEF/gRPC protocol.
type Client struct {
	grpcClient   stef_proto.STEFDestinationClient
	stream       stef_proto.STEFDestination_StreamClient
	callbacks    ClientCallbacks
	clientSchema ClientSchema
	logger       types.Logger

	// Running state
	waitCh     chan struct{}
	cancelFunc context.CancelFunc
}

// SendError will be returned by Client's sending functions. It may get wrapped by
// STEF writers, but can be checked via errors.Is(err, stefgrpc.SendError{})
// and is useful for distinguishing gRPC connection (sending) errors that are
// transient and can be retried, from STEF encoding errors which are permanent and
// should not be retried.
type SendError struct {
	err error
}

var _ error = (*SendError)(nil)

func (e SendError) Error() string {
	return "stefgrpc write error: " + e.err.Error()
}

func (e SendError) Unwrap() error { return e.err }

func (e SendError) Is(target error) bool {
	switch target.(type) {
	case SendError:
		return true
	default:
		return false
	}
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

	if err := w.stream.Send(&w.request); err != nil {
		return SendError{err: err}
	}
	return nil
}

var ErrServerInvalidResponse = errors.New("invalid server response")

// ClientSettings contains configuration settings for creating a Client.
type ClientSettings struct {
	// Logger instance used for logging client operations.
	Logger types.Logger

	// gRPC stream to send data over.
	GrpcClient stef_proto.STEFDestinationClient

	// ClientSchema of the client.
	ClientSchema ClientSchema

	// Callbacks for handling events such as acknowledgments and disconnections.
	Callbacks ClientCallbacks
}

type ClientSchema struct {
	// The name of the root struct of the schema.
	RootStructName string

	// The wire schema of the client.
	WireSchema *schema.WireSchema
}

// NewClient creates a new instance of the Client.
//
// Requirements:
// - The `RootStructName` in `ClientSchema` must not be empty.
// - The `WireSchema` in `ClientSchema` must not be nil.
//
// Example:
//
//	clientSettings := stefgrpc.ClientSettings{
//	    GrpcClient:   grpcClient,
//	    ClientSchema: stefgrpc.ClientSchema{RootStructName: "Metrics", WireSchema: &schema},
//	    Callbacks:    stefgrpc.ClientCallbacks{},
//	}
//	client, err := stefgrpc.NewClient(clientSettings)
//	if err != nil {
//	    log.Fatalf("Failed to create client: %v", err)
//	}
func NewClient(settings ClientSettings) (*Client, error) {
	if settings.Logger == nil {
		settings.Logger = internal.NopLogger{}
	}

	if settings.Callbacks.OnDisconnect == nil {
		settings.Callbacks.OnDisconnect = func(err error) {}
	}

	if settings.ClientSchema.RootStructName == "" {
		return nil, fmt.Errorf("client schema root struct name is empty")
	}
	if settings.ClientSchema.WireSchema == nil {
		return nil, fmt.Errorf("client schema wire schema is nil")
	}

	client := &Client{
		grpcClient:   settings.GrpcClient,
		callbacks:    settings.Callbacks,
		clientSchema: settings.ClientSchema,
		logger:       settings.Logger,
		waitCh:       make(chan struct{}),
	}

	return client, nil
}

func (c *Client) Connect(ctx context.Context) (pkg.ChunkWriter, pkg.WriterOptions, error) {
	c.logger.Debugf(context.Background(), "Begin connecting (client=%p)", c)

	opts := pkg.WriterOptions{}

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

	// Send the first message to the server, include the root struct name.
	clientMsg := &stef_proto.STEFClientFirstMessage{
		RootStructName: c.clientSchema.RootStructName,
	}
	err = stream.Send(
		&stef_proto.STEFClientMessage{
			FirstMessage: clientMsg,
		},
	)
	if err != nil {
		return nil, opts, fmt.Errorf("failed to send to server: %w", err)
	}

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
	var serverSchema schema.WireSchema
	buf := bytes.NewBuffer(capabilities.Capabilities.Schema)
	err = serverSchema.Deserialize(buf)
	if err != nil {
		return nil, opts, fmt.Errorf("failed to unmarshal capabilities schema: %w", err)
	}

	// Check if server schema is backward compatible with client schema.
	compatibility, err := serverSchema.Compatible(c.clientSchema.WireSchema)
	switch compatibility {
	case schema.CompatibilityExact:
		// Schemas match exactly, nothing else is needed, can start sending data.

	case schema.CompatibilitySuperset:
		// ServerStream schema is superset of client schema. The client MUST specify its schema
		// in the STEF header.
		opts.IncludeDescriptor = true
		opts.Schema = c.clientSchema.WireSchema

	case schema.CompatibilityIncompatible:
		// It is neither exact match nor is server schema a superset, but server schema maybe subset.
		// Check the opposite direction: if client schema is backward compatible with server schema.
		compatibility, err = serverSchema.Compatible(c.clientSchema.WireSchema)

		if err != nil || compatibility == schema.CompatibilityIncompatible {
			return nil, opts, fmt.Errorf("client and server schemas are incompatble: %w", err)
		}

		if compatibility == schema.CompatibilitySuperset {
			// Client schema is superset of server schema. The client MUST downgrade its schema.
			opts.IncludeDescriptor = true
			opts.Schema = c.clientSchema.WireSchema
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

	c.logger.Debugf(context.Background(), "Begin receiving acks (client=%p)", c)

	for {
		resp, err := c.stream.Recv()
		if err != nil {
			if err == io.EOF {
				c.callbacks.OnDisconnect(nil)
				return
			}
			c.logger.Errorf(context.Background(), "Error receiving acks: %v (client=%p)", err, c)
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
