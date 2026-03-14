package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	stefgrpc "github.com/splunk/stef/go/grpc"
	"github.com/splunk/stef/go/grpc/stef_proto"
	"github.com/splunk/stef/go/pkg/schema"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type stdLogger struct{}

func (stdLogger) Debugf(_ context.Context, format string, v ...interface{}) {
	log.Printf("[DEBUG] "+format, v...)
}

func (stdLogger) Errorf(_ context.Context, format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

func main() {
	if err := run(); err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: %s <server|client> [flags]", os.Args[0])
	}

	switch os.Args[1] {
	case "server":
		return runServer(os.Args[2:])
	case "client":
		return runClient(os.Args[2:])
	default:
		return fmt.Errorf("unknown mode %q", os.Args[1])
	}
}

func runServer(args []string) error {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	listen := fs.String("listen", "127.0.0.1:0", "address to listen on")
	rootStruct := fs.String("root-struct", "TestStruct", "root struct name")
	fieldCount := fs.Int("field-count", 2, "number of fields in wire schema")
	expectedChunkHex := fs.String("expected-chunk-hex", "", "expected chunk bytes as hex")
	ackID := fs.Uint64("ack-id", 1, "ack id to send after receiving expected chunk")
	timeout := fs.Duration("timeout", 10*time.Second, "max time to wait for one stream exchange")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *fieldCount <= 0 {
		return errors.New("field-count must be > 0")
	}

	expected, err := hex.DecodeString(*expectedChunkHex)
	if err != nil {
		return fmt.Errorf("invalid expected-chunk-hex: %w", err)
	}

	wireSchema := makeWireSchema(*rootStruct, *fieldCount)
	if wireSchema == nil {
		return errors.New("failed to build wire schema")
	}

	listener, err := net.Listen("tcp", *listen)
	if err != nil {
		return fmt.Errorf("listen failed: %w", err)
	}
	defer listener.Close()

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	doneCh := make(chan error, 1)
	streamServer := stefgrpc.NewStreamServer(stefgrpc.ServerSettings{
		Logger:       stdLogger{},
		ServerSchema: wireSchema,
		MaxDictBytes: 1024,
		Callbacks: stefgrpc.Callbacks{OnStream: func(reader stefgrpc.GrpcReader, stream stefgrpc.STEFStream) error {
			got := make([]byte, len(expected))
			if _, err := io.ReadFull(reader, got); err != nil {
				doneCh <- fmt.Errorf("failed to read expected bytes: %w", err)
				return err
			}
			if !bytes.Equal(got, expected) {
				err := fmt.Errorf("unexpected chunk bytes: got=%x want=%x", got, expected)
				doneCh <- err
				return err
			}
			if err := stream.SendDataResponse(&stef_proto.STEFDataResponse{AckRecordId: *ackID}); err != nil {
				doneCh <- fmt.Errorf("failed to send ack: %w", err)
				return err
			}
			doneCh <- nil
			return nil
		}},
	})

	stef_proto.RegisterSTEFDestinationServer(grpcServer, streamServer)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			doneCh <- fmt.Errorf("grpc server failed: %w", err)
		}
	}()

	fmt.Printf("READY %s\n", listener.Addr().String())
	log.Printf("listening on %s", listener.Addr().String())

	select {
	case err := <-doneCh:
		return err
	case <-time.After(*timeout):
		return fmt.Errorf("timed out after %s waiting for client exchange", *timeout)
	}
}

func runClient(args []string) error {
	fs := flag.NewFlagSet("client", flag.ContinueOnError)
	target := fs.String("target", "", "server target host:port")
	rootStruct := fs.String("root-struct", "TestStruct", "root struct name")
	fieldCount := fs.Int("field-count", 2, "number of fields in wire schema")
	headerHex := fs.String("header-hex", "", "chunk header bytes as hex")
	contentHex := fs.String("content-hex", "", "chunk payload bytes as hex")
	expectAck := fs.Uint64("expect-ack", 1, "expected ack id")
	timeout := fs.Duration("timeout", 10*time.Second, "max time to complete exchange")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *target == "" {
		return errors.New("target is required")
	}
	if *fieldCount <= 0 {
		return errors.New("field-count must be > 0")
	}

	header, err := hex.DecodeString(*headerHex)
	if err != nil {
		return fmt.Errorf("invalid header-hex: %w", err)
	}
	content, err := hex.DecodeString(*contentHex)
	if err != nil {
		return fmt.Errorf("invalid content-hex: %w", err)
	}

	conn, err := grpc.NewClient(
		*target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to create grpc connection: %w", err)
	}
	defer conn.Close()

	ackCh := make(chan uint64, 1)
	client, err := stefgrpc.NewClient(stefgrpc.ClientSettings{
		Logger:     stdLogger{},
		GrpcClient: stef_proto.NewSTEFDestinationClient(conn),
		ClientSchema: stefgrpc.ClientSchema{
			RootStructName: *rootStruct,
			WireSchema:     makeWireSchema(*rootStruct, *fieldCount),
		},
		Callbacks: stefgrpc.ClientCallbacks{
			OnDisconnect: func(err error) {
				if err != nil {
					log.Printf("disconnect: %v", err)
				}
			},
			OnAck: func(ackID uint64) error {
				select {
				case ackCh <- ackID:
				default:
				}
				return nil
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	writer, _, err := client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}

	if err := writer.WriteChunk(header, content); err != nil {
		return fmt.Errorf("write chunk failed: %w", err)
	}

	select {
	case ackID := <-ackCh:
		if ackID != *expectAck {
			return fmt.Errorf("unexpected ack id: got=%d want=%d", ackID, *expectAck)
		}
	case <-ctx.Done():
		return fmt.Errorf("timed out waiting for ack: %w", ctx.Err())
	}

	disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer disconnectCancel()
	if err := client.Disconnect(disconnectCtx); err != nil {
		return fmt.Errorf("disconnect failed: %w", err)
	}

	return nil
}

func makeWireSchema(rootName string, fieldCount int) *schema.WireSchema {
	st := schema.NewStruct()
	st.Name = rootName
	st.IsRoot = true
	for i := 0; i < fieldCount; i++ {
		st.AddField(&schema.StructField{
			Name: fmt.Sprintf("f%d", i),
			FieldType: schema.FieldType{
				Primitive: &schema.PrimitiveType{Type: schema.PrimitiveTypeUint64},
			},
		})
	}

	sch := &schema.Schema{
		Structs:   map[string]*schema.Struct{rootName: st},
		Multimaps: map[string]*schema.Multimap{},
		Enums:     map[string]*schema.Enum{},
	}
	wireSchema := schema.NewWireSchema(sch, rootName)
	return &wireSchema
}
