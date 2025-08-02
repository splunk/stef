package stefgrpc

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/splunk/stef/go/grpc/stef_proto"
	"github.com/splunk/stef/go/pkg/schema"
)

func newGrpcServer() (*grpc.Server, net.Listener, int) {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:0"))
	if err != nil {
		log.Fatalf("Failed to listen on a tcp port: %v", err)
	}
	serverPort := listener.Addr().(*net.TCPAddr).Port
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	return grpcServer, listener, serverPort
}

// mockDestinationServer unified mock server that handles both basic functionality and schema capabilities
type mockDestinationServer struct {
	stef_proto.UnimplementedSTEFDestinationServer
	serverSchema  *schema.WireSchema
	receivedData  []byte
	receivedCount atomic.Int64
	useSchemaMode bool // determines whether to use schema capabilities mode or basic mode
}

// testLogger implements a simple logger for testing
type testLogger struct{}

func (l *testLogger) Debugf(ctx context.Context, format string, v ...interface{}) {
	log.Printf("[DEBUG] "+format, v...)
}

func (l *testLogger) Errorf(ctx context.Context, format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

func (s *mockDestinationServer) Stream(server stef_proto.STEFDestination_StreamServer) error {
	// Handle schema capabilities exchange if in schema mode
	if s.useSchemaMode {
		if err := s.handleCapabilitiesExchange(server); err != nil {
			return err
		}
		// Continue to unified message processing
		return s.processMessages(server)
	}

	// Basic mode - use chunk assembler for simple data processing
	reader := newChunkAssembler(newGrpcChunkSource(server))
	for {
		buf := make([]byte, 1024)
		n, err := reader.Read(buf)
		if err != nil {
			return err
		}
		s.receivedData = append(s.receivedData, buf[:n]...)
		s.receivedCount.Add(1)
	}
}

// handleCapabilitiesExchange handles the initial capabilities exchange for schema mode
func (s *mockDestinationServer) handleCapabilitiesExchange(server stef_proto.STEFDestination_StreamServer) error {
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
	serverCapabilities := &stef_proto.STEFDestinationCapabilities{}

	if s.serverSchema != nil {
		// Properly serialize server schema using the same method as original tests
		var schemaBytes bytes.Buffer
		err = s.serverSchema.Serialize(&schemaBytes)
		if err != nil {
			return fmt.Errorf("could not marshal server schema: %w", err)
		}
		serverCapabilities.Schema = schemaBytes.Bytes()
	}

	capabilitiesMsg := &stef_proto.STEFServerMessage{
		Message: &stef_proto.STEFServerMessage_Capabilities{
			Capabilities: serverCapabilities,
		},
	}

	if err := server.Send(capabilitiesMsg); err != nil {
		return fmt.Errorf("failed to send capabilities: %w", err)
	}
	return nil
}

// processMessages handles the main message processing loop for schema mode
func (s *mockDestinationServer) processMessages(server stef_proto.STEFDestination_StreamServer) error {
	ackId := uint64(1) // Simple counter for ack IDs
	
	for {
		msg, err := server.Recv()
		if err != nil {
			return err
		}

		if len(msg.StefBytes) > 0 {
			// Send ack response with incrementing ID
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
			
			ackId++ // Increment for next ack
		}
	}
}

func TestGrpcReaderDestinationServer(t *testing.T) {
	grpcServer, listener, serverPort := newGrpcServer()
	mockServer := &mockDestinationServer{
		useSchemaMode: false, // Use basic mode for this test
	}
	stef_proto.RegisterSTEFDestinationServer(grpcServer, mockServer)
	go func() {
		grpcServer.Serve(listener)
	}()

	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", serverPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()
	client := stef_proto.NewSTEFDestinationClient(conn)

	stream, err := client.Stream(context.Background())
	require.NoError(t, err)

	request := &stef_proto.STEFClientMessage{
		StefBytes: []byte{1, 2, 3},
	}
	err = stream.Send(request)
	require.NoError(t, err)

	request = &stef_proto.STEFClientMessage{
		StefBytes:    []byte{4, 5},
		IsEndOfChunk: true,
	}
	err = stream.Send(request)
	require.NoError(t, err)

	require.Eventually(
		t, func() bool {
			return mockServer.receivedCount.Load() == 1
		},
		5*time.Second,
		5*time.Millisecond,
	)

	assert.EqualValues(t, []byte{1, 2, 3, 4, 5}, mockServer.receivedData)

	grpcServer.Stop()
}

func TestSchemaCompatibility(t *testing.T) {
	testCases := []struct {
		name                      string
		clientFieldCount          uint
		serverFieldCount          uint
		clientStructName          string
		serverStructName          string
		expectedCompatibility     schema.Compatibility
		expectedIncludeDescriptor bool
		expectedSchemaNotNil      bool
		expectConnectError        bool
	}{
		{
			name:                      "ServerSuperset",
			clientFieldCount:          2,
			serverFieldCount:          3,
			clientStructName:          "TestStruct",
			serverStructName:          "TestStruct",
			expectedCompatibility:     schema.CompatibilitySuperset,
			expectedIncludeDescriptor: true,
			expectedSchemaNotNil:      true,
			expectConnectError:        false,
		},
		{
			name:                      "ClientSuperset",
			clientFieldCount:          3,
			serverFieldCount:          2,
			clientStructName:          "TestStruct",
			serverStructName:          "TestStruct",
			expectedCompatibility:     schema.CompatibilitySuperset,
			expectedIncludeDescriptor: true,
			expectedSchemaNotNil:      true,
			expectConnectError:        false,
		},
		{
			name:                      "ExactMatch",
			clientFieldCount:          2,
			serverFieldCount:          2,
			clientStructName:          "TestStruct",
			serverStructName:          "TestStruct",
			expectedCompatibility:     schema.CompatibilityExact,
			expectedIncludeDescriptor: false,
			expectedSchemaNotNil:      false,
			expectConnectError:        false,
		},
		{
			name:                      "IncompatibleSchemas",
			clientFieldCount:          2,
			serverFieldCount:          2,
			clientStructName:          "ClientStruct",
			serverStructName:          "ServerStruct",
			expectedCompatibility:     schema.CompatibilityIncompatible,
			expectedIncludeDescriptor: false,
			expectedSchemaNotNil:      false,
			expectConnectError:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup gRPC server
			grpcServer, listener, serverPort := newGrpcServer()

			// Create schemas
			clientSchema := &schema.WireSchema{
				StructFieldCount: map[string]uint{
					tc.clientStructName: tc.clientFieldCount,
				},
			}
			serverSchema := &schema.WireSchema{
				StructFieldCount: map[string]uint{
					tc.serverStructName: tc.serverFieldCount,
				},
			}

			// Setup mock server with schema capabilities
			mockServer := &mockDestinationServer{
				serverSchema:  serverSchema,
				useSchemaMode: true, // Use schema mode for compatibility tests
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

			// Setup client with mock logger
			mockLogger := &testLogger{}

			settings := ClientSettings{
				Logger:     mockLogger,
				GrpcClient: grpcClient,
				ClientSchema: ClientSchema{
					RootStructName: tc.clientStructName,
					WireSchema:     clientSchema,
				},
				Callbacks: ClientCallbacks{
					OnDisconnect: func(err error) {},
					OnAck: func(ackId uint64) error {
						return nil
					},
				},
			}

			client, err := NewClient(settings)
			require.NoError(t, err)

			// Connect - handle both compatible and incompatible cases
			writer, opts, err := client.Connect(context.Background())

			if tc.expectConnectError {
				// For incompatible schemas, connect should fail
				require.Error(t, err, "Connect should fail for incompatible schemas")
				assert.Contains(t, err.Error(), "incompatble", "Error should mention schema incompatibility")
			} else {
				// For compatible schemas, connect should succeed
				require.NoError(t, err, "Connect should succeed for compatible schemas")
				require.NotNil(t, writer, "Writer should not be nil")

				// Assert schema compatibility behavior in connect() function
				assert.Equal(t, tc.expectedIncludeDescriptor, opts.IncludeDescriptor,
					"IncludeDescriptor should be %v for %s compatibility", tc.expectedIncludeDescriptor, tc.name)

				if tc.expectedSchemaNotNil {
					assert.Equal(t, clientSchema, opts.Schema, "Schema should be set to client schema for superset compatibility")
				} else {
					assert.Nil(t, opts.Schema, "Schema should be nil for exact match")
				}
			}
			grpcServer.Stop()
		})
	}
}
