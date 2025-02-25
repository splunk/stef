// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package stefreceiver // import "go.opentelemetry.io/collector/receiver/otlpreceiver"

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componentstatus"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	stefgrpc "github.com/splunk/stef/go/grpc"
	"github.com/splunk/stef/go/grpc/stef_proto"
	"github.com/splunk/stef/go/otel/oteltef"
	stefpdatametrics "github.com/splunk/stef/go/pdata/metrics"
)

// stefReceiver is the type that exposes Metrics reception.
type stefReceiver struct {
	cfg        *Config
	serverGRPC *grpc.Server

	nextMetrics consumer.Metrics
	shutdownWG  sync.WaitGroup

	settings *receiver.Settings
}

// newStefReceiver just creates the OpenTelemetry receiver services. It is the caller's
// responsibility to invoke the respective Start*Reception methods as well
// as the various Stop*Reception methods to end it.
func newStefReceiver(cfg *Config, set *receiver.Settings, nextMetrics consumer.Metrics) (*stefReceiver, error) {
	r := &stefReceiver{
		cfg:         cfg,
		nextMetrics: nextMetrics,
		settings:    set,
	}
	return r, nil
}

func (r *stefReceiver) startGRPCServer(host component.Host) error {
	var err error
	if r.serverGRPC, err = r.cfg.ServerConfig.ToServer(
		context.Background(), host, r.settings.TelemetrySettings,
	); err != nil {
		return err
	}

	r.serverGRPC = grpc.NewServer()

	r.settings.Logger.Info("Starting GRPC server", zap.String("endpoint", r.cfg.ServerConfig.NetAddr.Endpoint))
	var gln net.Listener
	if gln, err = r.cfg.ServerConfig.NetAddr.Listen(context.Background()); err != nil {
		return err
	}

	schema, err := oteltef.MetricsWireSchema()
	if err != nil {
		log.Fatalf("Failed to load schema: %v", err)
	}

	settings := stefgrpc.ServerSettings{
		Logger:       nil,
		ServerSchema: &schema,
		MaxDictBytes: 0,
		OnStream:     r.onStream,
	}
	stefSrv := stefgrpc.NewStreamServer(settings)
	stef_proto.RegisterSTEFDestinationServer(r.serverGRPC, stefSrv)

	r.shutdownWG.Add(1)
	go func() {
		defer r.shutdownWG.Done()

		if errGrpc := r.serverGRPC.Serve(gln); errGrpc != nil && !errors.Is(errGrpc, grpc.ErrServerStopped) {
			componentstatus.ReportStatus(host, componentstatus.NewFatalErrorEvent(errGrpc))
		}
	}()
	return nil
}

// Start runs the trace receiver on the gRPC server. Currently
// it also enables the metrics receiver too.
func (r *stefReceiver) Start(ctx context.Context, host component.Host) error {
	if err := r.startGRPCServer(host); err != nil {
		return err
	}
	return nil
}

// Shutdown is a method to turn off receiving.
func (r *stefReceiver) Shutdown(ctx context.Context) error {
	var err error

	if r.serverGRPC != nil {
		r.serverGRPC.GracefulStop()
	}

	r.shutdownWG.Wait()
	return err
}

func (r *stefReceiver) onStream(grpcReader stefgrpc.GrpcReader, ackFunc func(sequenceId uint64) error) error {
	r.settings.Logger.Info("Incoming STEF/gRPC connection.")

	reader, err := oteltef.NewMetricsReader(grpcReader)
	if err != nil {
		r.settings.Logger.Error("Cannot decode data on incoming STEF/gRPC connection", zap.Error(err))
		return err
	}

	done := make(chan struct{})
	defer close(done)

	type BadData struct {
		from, to uint64
	}
	badDataCh := make(chan BadData)
	defer close(badDataCh)

	var lastReadRecord atomic.Uint64
	go func() {
		t := time.NewTicker(10 * time.Millisecond)
		var acksSent uint64
		var lastAcked uint64
		for {
			select {
			case <-badDataCh: // TODO add ability to report bad data.

			case <-t.C:
				readRecordCount := lastReadRecord.Load()
				if readRecordCount > lastAcked {
					lastAcked = readRecordCount
					err = ackFunc(lastAcked)
					if err != nil {
						r.settings.Logger.Error("Error acking STEF gRPC connection", zap.Error(err))
						return
					}
					acksSent++
				}
				//stats := grpcReader.Stats()
				//fmt.Printf(
				//	"Records: %v, Messages: %v, Bytes: %v, Bytes/point: %.2f, Acks: %v, Last ACKID: %v  \r",
				//	readRecordCount,
				//	stats.MessagesReceived,
				//	stats.BytesReceived,
				//	float64(stats.BytesReceived)/float64(readRecordCount),
				//	acksSent,
				//	lastAcked,
				//)
			case <-done:
				return
			}
		}
	}()

	converter := stefpdatametrics.STEFToOTLPUnsorted{}

	for {
		fromRecordID := reader.RecordCount()
		mdata, err := converter.ConvertTillAvailable(reader)
		if err != nil {
			r.settings.Logger.Error("Cannot read from STEF/gRPC connection", zap.Error(err))
			return err
		}

		toRecordID := reader.RecordCount()
		if err := r.nextMetrics.ConsumeMetrics(context.Background(), mdata); err != nil {

			r.settings.Logger.Error(
				"Error pushing data to consumer",
				zap.Error(err),
				zap.Uint64("fromID", fromRecordID),
				zap.Uint64("toID", toRecordID),
			)

			badDataCh <- BadData{
				from: fromRecordID,
				to:   toRecordID,
			}
		} else {
			lastReadRecord.Store(toRecordID)
		}
	}
}
