// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package stefreceiver // import "go.opentelemetry.io/collector/receiver/otlpreceiver"

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componentstatus"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	stefgrpc "github.com/splunk/stef/go/grpc"
	"github.com/splunk/stef/go/grpc/stef_proto"
	"github.com/splunk/stef/go/otel/oteltef"
	stefpdatametrics "github.com/splunk/stef/go/pdata/metrics"
	"github.com/splunk/stef/otelcol/internal/stefreceiver/internal"
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
		Callbacks:    stefgrpc.Callbacks{OnStream: r.onStream},
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

// Start runs the STEF gRPC receiver.
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
		// TODO: Graceful shutdown.
		r.serverGRPC.Stop()
	}

	r.shutdownWG.Wait()
	return err
}

func (r *stefReceiver) onStream(grpcReader stefgrpc.GrpcReader, stream stefgrpc.STEFStream) error {
	r.settings.Logger.Info("Incoming STEF/gRPC connection.")

	reader, err := oteltef.NewMetricsReader(grpcReader)
	if err != nil {
		r.settings.Logger.Error("Cannot decode data on incoming STEF/gRPC connection", zap.Error(err))
		return err
	}

	// Create a responder for this stream and run it in a separate goroutine.
	resp := internal.NewResponder(r.settings.Logger, stream)
	defer resp.Stop()
	go resp.Run()

	converter := stefpdatametrics.STEFToOTLPUnsorted{}

	// Read, decode, convert the incoming data and push it to the next consumer.
	for {
		respError := resp.LastError()
		if respError != nil {
			// We had problem sending responses. Can't continue using this connection since
			// responding is essential for operation.
			r.settings.Logger.Error(
				"Closing STEF/gRPC connection since responding failed",
				zap.Error(respError),
			)
			return respError
		}

		// Mark the start of the converted batch.
		fromRecordID := reader.RecordCount()

		// Read and convert records. We use ConvertTillEndOfFrame to make sure we are not
		// blocked in the middle of a batch indefinitely, with lingering data in memory,
		// neither pushed to pipeline, nor acked.
		mdata, err := converter.ConvertTillEndOfFrame(reader)
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.Canceled {
				// A regular disconnection case. The client closed the connection.
				r.settings.Logger.Debug("STEF/gRPC connection closed", zap.Error(err))
			} else {
				r.settings.Logger.Error("Cannot read from STEF/gRPC connection", zap.Error(err))
			}
			return err
		}

		toRecordID := reader.RecordCount()

		// Push converted data to the next consumer.
		if err := r.nextMetrics.ConsumeMetrics(context.Background(), mdata); err != nil {
			r.settings.Logger.Error(
				"Error pushing data to consumer",
				zap.Error(err),
				zap.Uint64("fromID", fromRecordID),
				zap.Uint64("toID", toRecordID),
			)

			if consumererror.IsPermanent(err) {
				// This is a permanent error. Let the client know and continue receiving data.
				resp.ScheduleBadDataResponse(
					internal.BadData{
						FromID: fromRecordID,
						ToID:   toRecordID,
					},
				)
			} else {
				// The next consumer is temporarily unable to process the data.
				// Close the stream and indicate to client to try again later.
				return status.New(codes.Unavailable, "try again later").Err()
			}
		} else {
			// Successfully received and consumed. Acknowledge it.
			resp.ScheduleAck(toRecordID)
		}
	}
}
