package stefgrpc

import (
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

	"github.com/splunk/stef/stef-gogrpc/stef_proto"
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

type mockDestinationServer struct {
	stef_proto.UnimplementedSTEFDestinationServer
	receivedData  []byte
	receivedCount atomic.Int64
}

func (s *mockDestinationServer) Stream(server stef_proto.STEFDestination_StreamServer) error {
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
	return nil
}

func TestGrpcReaderDestinationServer(t *testing.T) {
	grpcServer, listener, serverPort := newGrpcServer()
	mockServer := &mockDestinationServer{}
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
