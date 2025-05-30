package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	stefgrpc "github.com/splunk/stef/go/grpc"
	"github.com/splunk/stef/go/pkg"

	"github.com/splunk/stef/go/otel/oteltef"

	"github.com/splunk/stef/go/grpc/stef_proto"
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

func TestGrpcWriteRead(t *testing.T) {
	grpcServer, listener, serverPort := newGrpcServer()

	schema, err := oteltef.MetricsWireSchema()
	require.NoError(t, err)

	recordsReceivedAndVerified := make(chan struct{})
	serverSettings := stefgrpc.ServerSettings{
		ServerSchema: &schema,
		Callbacks: stefgrpc.Callbacks{
			OnStream: func(source stefgrpc.GrpcReader, stream stefgrpc.STEFStream) error {
				reader, err := oteltef.NewMetricsReader(source)
				require.NoError(t, err)

				// Read and verify that received records match what was sent.
				err = reader.Read(pkg.ReadOptions{})
				require.NoError(t, err)
				require.EqualValues(t, "abc", reader.Record.Metric().Name())

				// Send acknowledgment to the client.
				err = stream.SendDataResponse(&stef_proto.STEFDataResponse{AckRecordId: reader.RecordCount()})
				if err != nil {
					log.Printf("Error sending ack record id to server: %v", err)
				}

				// Signal that verification is done.
				close(recordsReceivedAndVerified)
				return nil
			},
		},
	}

	server := stefgrpc.NewStreamServer(serverSettings)

	// Start a server.
	stef_proto.RegisterSTEFDestinationServer(grpcServer, server)
	go func() {
		grpcServer.Serve(listener)
	}()

	// Connect to the server.
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", serverPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	// Open a stream to the server.
	grpcClient := stef_proto.NewSTEFDestinationClient(conn)

	// Last Id acked by the server.
	var lastAckId atomic.Uint64

	// Prepare client callbacks.
	callbacks := stefgrpc.ClientCallbacks{
		OnAck: func(ackId uint64) error {
			lastAckId.Store(ackId)
			return nil
		},
	}
	require.NoError(t, err)

	clientSettings := stefgrpc.ClientSettings{
		GrpcClient:   grpcClient,
		ClientSchema: stefgrpc.ClientSchema{WireSchema: &schema, RootStructName: oteltef.MetricsStructName},
		Callbacks:    callbacks,
	}
	client, err := stefgrpc.NewClient(clientSettings)
	require.NoError(t, err)

	w, opts, err := client.Connect(context.Background())
	require.NoError(t, err)

	opts.Compression = pkg.CompressionZstd

	// Create record writer.
	writer, err := oteltef.NewMetricsWriter(w, opts)
	require.NoError(t, err)

	// Write the records.
	writer.Record.Metric().SetName("abc")
	err = writer.Write()
	require.NoError(t, err)
	lastSentId := writer.RecordCount()

	// Make sure data is sent.
	err = writer.Flush()
	require.NoError(t, err)

	// Make sure all data is acknowledged by the server.
	require.Eventually(
		t, func() bool {
			// Wait until the last piece of data tha was sent by the client is acked by the server.
			return lastAckId.Load() == lastSentId
		}, 5*time.Second, 10*time.Millisecond,
	)

	// Wait until all records are decoded by the server and verified.
	select {
	case <-recordsReceivedAndVerified:
	case <-time.Tick(5 * time.Second):
		t.Error("Timed out waiting for records")
	}

	// Close connection so that client's next sending fails.
	conn.Close()
	grpcServer.Stop()

	// Write a record.
	err = writer.Write()
	require.NoError(t, err)

	// Try to send the data.
	err = writer.Flush()

	// Make sure we see the expected error.
	require.True(t, errors.Is(err, stefgrpc.SendError{}))
}

func TestDictReset(t *testing.T) {
	grpcServer, listener, serverPort := newGrpcServer()
	defer grpcServer.Stop()

	schema, err := oteltef.MetricsWireSchema()
	require.NoError(t, err)

	const nameLen = 1000

	// This sequence is crafted to catch a bug when Reader fails to resets
	// its dict in sync with Writer.
	metricNames := []string{
		strings.Repeat("a", nameLen), // elem 1 in dict
		strings.Repeat("b", nameLen), // elem 2 in dict
		strings.Repeat("c", nameLen), // elem 3 in dict, hit the dict limit, next record is a new frame
		"d",                          // elem 1 in dict, This starts a new frame and resets dicts
		strings.Repeat("c", nameLen), // elem 2 in dict, this adds a new element to dicts
		"d",                          // elem 1 in dict, but if we have a bug in reader reset this will be elem 4
	}

	recordsReceivedAndVerified := make(chan struct{})
	serverSettings := stefgrpc.ServerSettings{
		MaxDictBytes: uint64(2*nameLen + 500), // Fit 2 elems, 3rd is over limit.
		ServerSchema: &schema,
		Callbacks: stefgrpc.Callbacks{
			OnStream: func(source stefgrpc.GrpcReader, stream stefgrpc.STEFStream) error {
				reader, err := oteltef.NewMetricsReader(source)
				require.NoError(t, err)

				// Read and verify that received records match what was sent.
				for i, metricName := range metricNames {
					err := reader.Read(pkg.ReadOptions{})
					require.NoError(t, err)
					assert.EqualValues(t, metricName, reader.Record.Metric().Name(), i)
				}

				// Send acknowledgment to the client.
				err = stream.SendDataResponse(&stef_proto.STEFDataResponse{AckRecordId: reader.RecordCount()})
				if err != nil {
					log.Printf("Error sending ack record id to server: %v", err)
				}

				// Signal that verification is done.
				close(recordsReceivedAndVerified)
				return nil
			},
		},
	}

	server := stefgrpc.NewStreamServer(serverSettings)

	// Start a server.
	stef_proto.RegisterSTEFDestinationServer(grpcServer, server)
	go func() {
		grpcServer.Serve(listener)
	}()

	// Connect to the server.
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", serverPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	// Open a stream to the server.
	grpcClient := stef_proto.NewSTEFDestinationClient(conn)

	// Last Id acked by the server.
	var lastAckId atomic.Uint64

	// Prepare client callbacks.
	callbacks := stefgrpc.ClientCallbacks{
		OnAck: func(ackId uint64) error {
			lastAckId.Store(ackId)
			return nil
		},
	}
	require.NoError(t, err)

	clientSettings := stefgrpc.ClientSettings{
		GrpcClient:   grpcClient,
		ClientSchema: stefgrpc.ClientSchema{WireSchema: &schema, RootStructName: oteltef.MetricsStructName},
		Callbacks:    callbacks,
	}
	client, err := stefgrpc.NewClient(clientSettings)
	require.NoError(t, err)

	w, opts, err := client.Connect(context.Background())
	require.NoError(t, err)

	opts.Compression = pkg.CompressionZstd

	// Create record writer.
	writer, err := oteltef.NewMetricsWriter(w, opts)
	require.NoError(t, err)

	// Write the records.
	for _, metricName := range metricNames {
		writer.Record.Metric().SetName(metricName)
		err = writer.Write()
		require.NoError(t, err)
	}

	lastSentId := writer.RecordCount()

	// Make sure data is sent.
	err = writer.Flush()
	require.NoError(t, err)

	// Make sure all data is acknowledged by the server.
	require.Eventually(
		t, func() bool {
			// Wait until the last piece of data tha was sent by the client is acked by the server.
			return lastAckId.Load() == lastSentId
		}, 5*time.Second, 10*time.Millisecond,
	)

	// Wait until all records are decoded by the server and verified.
	select {
	case <-recordsReceivedAndVerified:
	case <-time.Tick(5 * time.Second):
		t.Error("Timed out waiting for records")
	}
}

func TestGrpcClientError(t *testing.T) {
	schema, err := oteltef.MetricsWireSchema()
	require.NoError(t, err)

	clientSettings := stefgrpc.ClientSettings{
		GrpcClient:   nil,
		ClientSchema: stefgrpc.ClientSchema{WireSchema: &schema},
	}
	_, err = stefgrpc.NewClient(clientSettings)
	require.Error(t, err)

	clientSettings = stefgrpc.ClientSettings{
		GrpcClient:   nil,
		ClientSchema: stefgrpc.ClientSchema{RootStructName: oteltef.MetricsStructName},
	}
	_, err = stefgrpc.NewClient(clientSettings)
	require.Error(t, err)
}
