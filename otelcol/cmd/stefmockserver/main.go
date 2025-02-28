package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"

	tefgrpc "github.com/splunk/stef/go/grpc"
	"github.com/splunk/stef/go/grpc/stef_proto"
	"github.com/splunk/stef/go/otel/oteltef"
	"github.com/splunk/stef/go/pkg"
)

func newGrpcServer(port int) (*grpc.Server, net.Listener, int) {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("Failed to listen on a tcp port: %v", err)
	}
	serverPort := listener.Addr().(*net.TCPAddr).Port
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	return grpcServer, listener, serverPort
}

type mockMetricDestServer struct {
	stef_proto.UnimplementedSTEFDestinationServer
}

func onStream(grpcReader tefgrpc.GrpcReader, ackFunc func(sequenceId uint64) error) error {
	log.Printf("Incoming STEF/gRPC connection.\n")

	reader, err := oteltef.NewMetricsReader(grpcReader)
	if err != nil {
		log.Printf("Cannot decode data on incoming STEF/gRPC connection: %v.\n", err)
		return err
	}

	done := make(chan struct{})
	defer close(done)

	var lastReadRecord atomic.Uint64
	go func() {
		t := time.NewTicker(100 * time.Millisecond)
		var acksSent uint64
		var lastAcked uint64
		for {
			select {
			case <-t.C:
				readRecordCount := lastReadRecord.Load()
				if readRecordCount > lastAcked {
					lastAcked = readRecordCount
					err = ackFunc(lastAcked)
					if err != nil {
						log.Fatalf("Error acking STEF gRPC connection: %v\n", err)
						return
					}
					acksSent++
				}
				stats := grpcReader.Stats()
				fmt.Printf(
					"Records: %v, Messages: %v, Bytes: %v, Bytes/point: %.2f, Acks: %v, Last ACKID: %v  \r",
					readRecordCount,
					stats.MessagesReceived,
					stats.BytesReceived,
					float64(stats.BytesReceived)/float64(readRecordCount),
					acksSent,
					lastAcked,
				)
			case <-done:
				return
			}
		}
	}()

	for {
		err = reader.Read(pkg.ReadOptions{})
		if err != nil {
			log.Printf("Cannot read from STEF/gRPC connection: %v.\n", err)
			return err
		}
		lastReadRecord.Store(reader.RecordCount())
	}
}

var ListenPort = 0

func main() {
	flag.IntVar(&ListenPort, "port", 4320, "The server listening port")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	grpcServer, listener, serverPort := newGrpcServer(ListenPort)
	log.Printf("Listening for STEF/gRPC on port %d\n", serverPort)

	schema, err := oteltef.MetricsWireSchema()
	if err != nil {
		log.Fatalf("Failed to load schema: %v", err)
	}

	settings := tefgrpc.ServerSettings{
		Logger:       nil,
		ServerSchema: &schema,
		MaxDictBytes: 0,
		OnStream:     onStream,
	}
	mockServer := tefgrpc.NewStreamServer(settings)
	stef_proto.RegisterSTEFDestinationServer(grpcServer, mockServer)
	_ = grpcServer.Serve(listener)
}
