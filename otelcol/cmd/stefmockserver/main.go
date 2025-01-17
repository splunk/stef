package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	tefgrpc "github.com/splunk/stef/go/grpc"
	"github.com/splunk/stef/go/grpc/stef_proto"
	"github.com/splunk/stef/go/otel/oteltef"
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
	log.Printf("Incoming TEF/gRPC connection.\n")

	reader, err := oteltef.NewMetricsReader(grpcReader)
	if err != nil {
		log.Printf("Cannot decode data on incoming TEF/gRPC connection: %v.\n", err)
		return err
	}

	acksSent := 0
	lackAckSent := time.Now()
	for {
		_, err := reader.Read()
		if err != nil {
			log.Printf("Cannot read from TEF/gRPC connection: %v.\n", err)
			return err
		}

		if time.Since(lackAckSent) > 100*time.Millisecond {
			err := ackFunc(reader.RecordCount())
			if err != nil {
				return err
			}
			acksSent++
			lackAckSent = time.Now()
		}

		stats := grpcReader.Stats()
		fmt.Printf(
			//"Sequence Id: %v, Messages: %v, Points: %v, Bytes: %v, Acks: %v\t\t\r",
			"Records: %v, Messages: %v, Bytes: %v, Bytes/point: %.2f, Acks: %v\t\t\r",
			reader.RecordCount(),
			stats.MessagesReceived,
			//reader.Stats().Datapoints,
			stats.BytesReceived,
			float64(stats.BytesReceived)/float64(reader.RecordCount()),
			acksSent,
		)
	}
}

var ListenPort = 0

func main() {
	flag.IntVar(&ListenPort, "port", 4320, "The server listening port")
	flag.Parse()

	grpcServer, listener, serverPort := newGrpcServer(ListenPort)
	fmt.Printf("Listening for STEF/gRPC on port %d\n", serverPort)

	schema, err := oteltef.MetricsWireSchema()
	if err != nil {
		log.Fatalf("Failed to load schema: %v", err)
	}

	settings := tefgrpc.ServerSettings{
		Logger:       nil,
		ServerSchema: schema,
		MaxDictBytes: 0,
		OnStream:     onStream,
	}
	mockServer := tefgrpc.NewStreamServer(settings)
	stef_proto.RegisterSTEFDestinationServer(grpcServer, mockServer)
	_ = grpcServer.Serve(listener)
}
