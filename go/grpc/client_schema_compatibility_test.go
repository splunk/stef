package stefgrpc

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/splunk/stef/go/grpc/stef_proto"
	"github.com/splunk/stef/go/pkg/schema"
)

// mockLogger implements types.Logger for testing
type mockLogger struct{}

func (l *mockLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	log.Printf("[DEBUG] "+format, args...)
}

func (l *mockLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

// mockDestinationServerWithCapabilities extends mockDestinationServer to handle capabilities
type mockDestinationServerWithCapabilities struct {
	stef_proto.UnimplementedSTEFDestinationServer
	serverSchema     *schema.WireSchema
	receivedMessages [][]byte
	mu               sync.Mutex
	acksSent         []uint64
	onStreamFunc     func(server stef_proto.STEFDestination_StreamServer) error
}

func (s *mockDestinationServerWithCapabilities) Stream(server stef_proto.STEFDestination_StreamServer) error {
	if s.onStreamFunc != nil {
		return s.onStreamFunc(server)
	}

	// Receive first message from client
	clientMsg, err := server.Recv()
	if err != nil {
		return fmt.Errorf("failed to receive first message: %w", err)
	}

	// Verify it's a first message with root struct name
	if clientMsg.FirstMessage == nil {
		return fmt.Errorf("expected first message, got nil")
	}

	// Send capabilities response with server schema
	var schemaBytes bytes.Buffer
	err = s.serverSchema.Serialize(&schemaBytes)
	if err != nil {
		return fmt.Errorf("could not marshal server schema: %w", err)
	}

	capabilities := &stef_proto.STEFDestinationCapabilities{
		Schema: schemaBytes.Bytes(),
		DictionaryLimits: &stef_proto.STEFDictionaryLimits{
			MaxDictBytes: 1024 * 1024, // 1MB
		},
	}

	capabilitiesMsg := &stef_proto.STEFServerMessage{
		Message: &stef_proto.STEFServerMessage_Capabilities{
			Capabilities: capabilities,
		},
	}

	if err := server.Send(capabilitiesMsg); err != nil {
		return fmt.Errorf("failed to send capabilities: %w", err)
	}

	// Continue receiving data messages and sending acks
	ackId := uint64(1)
	for {
		msg, err := server.Recv()
		if err != nil {
			return err
		}

		s.mu.Lock()
		s.receivedMessages = append(s.receivedMessages, msg.StefBytes)
		s.mu.Unlock()

		// Send ack
		ackMsg := &stef_proto.STEFServerMessage{
			Message: &stef_proto.STEFServerMessage_Response{
				Response: &stef_proto.STEFDataResponse{
					AckRecordId: ackId,
				},
			},
		}

		if err := server.Send(ackMsg); err != nil {
			return fmt.Errorf("failed to send ack: %w", err)
		}

		s.mu.Lock()
		s.acksSent = append(s.acksSent, ackId)
		s.mu.Unlock()

		ackId++
	}
}

// createClientSchema creates a simple client schema with 2 fields
func createClientSchema() *schema.WireSchema {
	clientSchema := &schema.WireSchema{
		StructFieldCount: map[string]uint{
			"TestStruct": 2, // Client has 2 fields
		},
	}
	return clientSchema
}

// createServerSchema creates a server schema with 3 fields (superset of client)
func createServerSchema() *schema.WireSchema {
	serverSchema := &schema.WireSchema{
		StructFieldCount: map[string]uint{
			"TestStruct": 3, // Server has 3 fields (superset)
		},
	}
	return serverSchema
}

func TestClientSchemaCompatibility_ServerSuperset(t *testing.T) {
	// Setup gRPC server
	grpcServer, listener, serverPort := newGrpcServer()

	// Create schemas
	clientSchema := createClientSchema()
	serverSchema := createServerSchema()

	// Verify compatibility before testing
	compatibility, err := serverSchema.Compatible(clientSchema)
	require.NoError(t, err)
	assert.Equal(t, schema.CompatibilitySuperset, compatibility, "Server schema should be superset of client schema")

	// Setup mock server with capabilities
	mockServer := &mockDestinationServerWithCapabilities{
		serverSchema: serverSchema,
	}
	stef_proto.RegisterSTEFDestinationServer(grpcServer, mockServer)

	go func() {
		grpcServer.Serve(listener)
	}()
	defer grpcServer.Stop()

	// Create gRPC client connection
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", serverPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	grpcClient := stef_proto.NewSTEFDestinationClient(conn)

	// Track callbacks
	var disconnectErr error
	var acksReceived []uint64
	var mu sync.Mutex

	// Setup client
	settings := ClientSettings{
		Logger:     &mockLogger{},
		GrpcClient: grpcClient,
		ClientSchema: ClientSchema{
			RootStructName: "TestStruct",
			WireSchema:     clientSchema,
		},
		Callbacks: ClientCallbacks{
			OnDisconnect: func(err error) {
				mu.Lock()
				disconnectErr = err
				mu.Unlock()
			},
			OnAck: func(ackId uint64) error {
				mu.Lock()
				acksReceived = append(acksReceived, ackId)
				mu.Unlock()
				return nil
			},
		},
	}

	client, err := NewClient(settings)
	require.NoError(t, err)

	// Connect and get writer
	writer, opts, err := client.Connect(context.Background())
	require.NoError(t, err)
	require.NotNil(t, writer)

	// Verify writer options are set correctly for superset compatibility
	assert.True(t, opts.IncludeDescriptor, "IncludeDescriptor should be true for superset compatibility")
	assert.Equal(t, clientSchema, opts.Schema, "Schema should be set to client schema")
	assert.Equal(t, uint(1024*1024), opts.MaxTotalDictSize, "MaxTotalDictSize should be set from server capabilities")

	// Write some test data
	testHeader := []byte("header")
	testContent := []byte("content")

	err = writer.WriteChunk(testHeader, testContent)
	require.NoError(t, err)

	// Wait for ack to be received
	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(acksReceived) > 0
	}, 2*time.Second, 10*time.Millisecond, "Should receive at least one ack")

	// Verify ack was received
	mu.Lock()
	assert.Len(t, acksReceived, 1, "Should receive exactly one ack")
	assert.Equal(t, uint64(1), acksReceived[0], "First ack should have ID 1")
	assert.Nil(t, disconnectErr, "Should not have disconnect error")
	mu.Unlock()

	// Verify server received the data
	mockServer.mu.Lock()
	assert.Len(t, mockServer.receivedMessages, 1, "Server should receive one message")
	expectedData := append(testHeader, testContent...)
	assert.Equal(t, expectedData, mockServer.receivedMessages[0], "Server should receive correct data")
	mockServer.mu.Unlock()

	// Disconnect
	err = client.Disconnect(context.Background())
	require.NoError(t, err)
}

func TestClientSchemaCompatibility_ClientSuperset(t *testing.T) {
	// Setup gRPC server
	grpcServer, listener, serverPort := newGrpcServer()

	// Create schemas
	clientSchema := &schema.WireSchema{
		StructFieldCount: map[string]uint{
			"TestStruct": 3, // Client has 3 fields
		},
	}
	serverSchema := &schema.WireSchema{
		StructFieldCount: map[string]uint{
			"TestStruct": 2, // Server has only 2 fields - incompatible
		},
	}

	// Verify compatibility before testing
	compatibility, err := clientSchema.Compatible(serverSchema)
	require.NoError(t, err)
	assert.Equal(t, schema.CompatibilitySuperset, compatibility, "Client schema should be superset of server schema")

	// Setup mock server with capabilities
	mockServer := &mockDestinationServerWithCapabilities{
		serverSchema: serverSchema,
	}
	stef_proto.RegisterSTEFDestinationServer(grpcServer, mockServer)

	go func() {
		grpcServer.Serve(listener)
	}()
	defer grpcServer.Stop()

	// Create gRPC client connection
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", serverPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	grpcClient := stef_proto.NewSTEFDestinationClient(conn)

	// Track callbacks
	var disconnectErr error
	var acksReceived []uint64
	var mu sync.Mutex

	// Setup client
	settings := ClientSettings{
		Logger:     &mockLogger{},
		GrpcClient: grpcClient,
		ClientSchema: ClientSchema{
			RootStructName: "TestStruct",
			WireSchema:     clientSchema,
		},
		Callbacks: ClientCallbacks{
			OnDisconnect: func(err error) {
				mu.Lock()
				disconnectErr = err
				mu.Unlock()
			},
			OnAck: func(ackId uint64) error {
				mu.Lock()
				acksReceived = append(acksReceived, ackId)
				mu.Unlock()
				return nil
			},
		},
	}

	client, err := NewClient(settings)
	require.NoError(t, err)

	// Connect and get writer
	writer, opts, err := client.Connect(context.Background())
	require.NoError(t, err)
	require.NotNil(t, writer)

	// Verify writer options are set correctly for superset compatibility
	assert.True(t, opts.IncludeDescriptor, "IncludeDescriptor should be true for superset compatibility")
	assert.Equal(t, clientSchema, opts.Schema, "Schema should be set to client schema")
	assert.Equal(t, uint(1024*1024), opts.MaxTotalDictSize, "MaxTotalDictSize should be set from server capabilities")

	// Write some test data
	testHeader := []byte("header")
	testContent := []byte("content")

	err = writer.WriteChunk(testHeader, testContent)
	require.NoError(t, err)

	// Wait for ack to be received
	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(acksReceived) > 0
	}, 2*time.Second, 10*time.Millisecond, "Should receive at least one ack")

	// Verify ack was received
	mu.Lock()
	assert.Len(t, acksReceived, 1, "Should receive exactly one ack")
	assert.Equal(t, uint64(1), acksReceived[0], "First ack should have ID 1")
	assert.Nil(t, disconnectErr, "Should not have disconnect error")
	mu.Unlock()

	// Verify server received the data
	mockServer.mu.Lock()
	assert.Len(t, mockServer.receivedMessages, 1, "Server should receive one message")
	expectedData := append(testHeader, testContent...)
	assert.Equal(t, expectedData, mockServer.receivedMessages[0], "Server should receive correct data")
	mockServer.mu.Unlock()

	// Disconnect
	err = client.Disconnect(context.Background())
	require.NoError(t, err)
}

func TestClientSchemaCompatibility_IncompatibleSchemas(t *testing.T) {
	// Setup gRPC server
	grpcServer, listener, serverPort := newGrpcServer()

	// Create truly incompatible schemas (different struct names)
	clientSchema := &schema.WireSchema{
		StructFieldCount: map[string]uint{
			"ClientStruct": 2, // Client has a struct that server doesn't
		},
	}
	serverSchema := &schema.WireSchema{
		StructFieldCount: map[string]uint{
			"ServerStruct": 2, // Server has a different struct that client doesn't
		},
	}

	// Verify schemas are incompatible in both directions
	compatibility, err := serverSchema.Compatible(clientSchema)
	assert.Error(t, err, "Should return error for incompatible schemas")
	assert.Equal(t, schema.CompatibilityIncompatible, compatibility, "Should be incompatible")

	compatibility, err = clientSchema.Compatible(serverSchema)
	assert.Error(t, err, "Should return error for incompatible schemas in reverse direction")
	assert.Equal(t, schema.CompatibilityIncompatible, compatibility, "Should be incompatible in reverse direction")

	// Setup mock server
	mockServer := &mockDestinationServerWithCapabilities{
		serverSchema: serverSchema,
	}
	stef_proto.RegisterSTEFDestinationServer(grpcServer, mockServer)

	go func() {
		grpcServer.Serve(listener)
	}()
	defer grpcServer.Stop()

	// Create gRPC client connection
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", serverPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	grpcClient := stef_proto.NewSTEFDestinationClient(conn)

	// Setup client
	settings := ClientSettings{
		Logger:     &mockLogger{},
		GrpcClient: grpcClient,
		ClientSchema: ClientSchema{
			RootStructName: "ClientStruct",
			WireSchema:     clientSchema,
		},
		Callbacks: ClientCallbacks{
			OnDisconnect: func(err error) {},
			OnAck:        func(ackId uint64) error { return nil },
		},
	}

	client, err := NewClient(settings)
	require.NoError(t, err)

	// Connect should fail due to incompatible schemas
	_, _, err = client.Connect(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "incompatble", "Error should mention schema incompatibility")
}

func TestClientSchemaCompatibility_ExactMatch(t *testing.T) {
	// Setup gRPC server
	grpcServer, listener, serverPort := newGrpcServer()

	// Create identical schemas
	clientSchema := &schema.WireSchema{
		StructFieldCount: map[string]uint{
			"TestStruct": 2,
		},
	}
	serverSchema := &schema.WireSchema{
		StructFieldCount: map[string]uint{
			"TestStruct": 2, // Exact match
		},
	}

	// Verify exact compatibility
	compatibility, err := serverSchema.Compatible(clientSchema)
	require.NoError(t, err)
	assert.Equal(t, schema.CompatibilityExact, compatibility, "Should be exact match")

	// Setup mock server
	mockServer := &mockDestinationServerWithCapabilities{
		serverSchema: serverSchema,
	}
	stef_proto.RegisterSTEFDestinationServer(grpcServer, mockServer)

	go func() {
		grpcServer.Serve(listener)
	}()
	defer grpcServer.Stop()

	// Create gRPC client connection
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", serverPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	grpcClient := stef_proto.NewSTEFDestinationClient(conn)

	// Setup client
	settings := ClientSettings{
		Logger:     &mockLogger{},
		GrpcClient: grpcClient,
		ClientSchema: ClientSchema{
			RootStructName: "TestStruct",
			WireSchema:     clientSchema,
		},
		Callbacks: ClientCallbacks{
			OnDisconnect: func(err error) {},
			OnAck:        func(ackId uint64) error { return nil },
		},
	}

	client, err := NewClient(settings)
	require.NoError(t, err)

	// Connect should succeed
	writer, opts, err := client.Connect(context.Background())
	require.NoError(t, err)
	require.NotNil(t, writer)

	// For exact match, IncludeDescriptor should be false
	assert.False(t, opts.IncludeDescriptor, "IncludeDescriptor should be false for exact match")
	assert.Nil(t, opts.Schema, "Schema should be nil for exact match")
	assert.Equal(t, uint(1024*1024), opts.MaxTotalDictSize, "MaxTotalDictSize should still be set")

	// Disconnect
	err = client.Disconnect(context.Background())
	require.NoError(t, err)
}
