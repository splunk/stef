package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/tigrannajaryan/stef/stef-go/pkg"
	stefgrpc "github.com/tigrannajaryan/stef/stef-gogrpc"

	"github.com/tigrannajaryan/stef/stef-otel/oteltef"

	"github.com/tigrannajaryan/stef/stef-gogrpc/stef_proto"
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
		ServerSchema: schema,
		OnStream: func(source stefgrpc.GrpcReader, ackFunc func(sequenceId uint64) error) error {
			reader, err := oteltef.NewMetricsReader(source)
			require.NoError(t, err)

			// Read and verify that received records match what was sent.
			record, err := reader.Read()
			require.NoError(t, err)
			require.EqualValues(t, "abc", record.Metric().Name())

			// Send acknowledgment to the client.
			err = ackFunc(reader.RecordCount())
			if err != nil {
				log.Printf("Error sending ack record id to server: %v", err)
			}

			// Signal that verification is done.
			close(recordsReceivedAndVerified)
			return nil
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
		ClientSchema: schema,
		Callbacks:    callbacks,
	}
	client := stefgrpc.NewClient(clientSettings)

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

	conn.Close()
	grpcServer.Stop()
}
