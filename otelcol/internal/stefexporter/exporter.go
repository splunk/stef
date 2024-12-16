// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package stefexporter // import "github.com/tigrannajaryan/stef/otelcol/internal/stefexporter"

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/open-telemetry/otel-arrow/pkg/config"
	"github.com/open-telemetry/otel-arrow/pkg/otel/arrow_record"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/tigrannajaryan/stef/stef-go/pkg"

	tef_grpc "github.com/tigrannajaryan/stef/stef-gogrpc"
	"github.com/tigrannajaryan/stef/stef-gogrpc/stef_proto"
	tef_grpctypes "github.com/tigrannajaryan/stef/stef-gogrpc/types"
	"github.com/tigrannajaryan/stef/stef-otel/oteltef"
	otlpconvert "github.com/tigrannajaryan/stef/stef-pdata/metrics"
	"github.com/tigrannajaryan/stef/stef-pdata/metrics/sortedbymetric"
)

type stefExporter struct {
	logger *zap.Logger

	cfg *Config

	arrowFile     *os.File
	arrowProducer *arrow_record.Producer

	grpcConn     *grpc.ClientConn
	writeMutex   sync.Mutex
	remoteWriter *oteltef.MetricsWriter

	lastSent time.Time

	// Ack fields
	ackMutex          sync.Mutex
	lastSentRecordId  uint64
	lastAckedRecordId uint64
	sentPendingAck    map[uint64]*sortedbymetric.SortedTree
	stopped           chan struct{}
}

type wrapLogger struct {
	logger *zap.Logger
}

func (w *wrapLogger) Debugf(ctx context.Context, format string, v ...interface{}) {
	w.logger.Debug(fmt.Sprintf(format, v...))
}

func (w *wrapLogger) Errorf(ctx context.Context, format string, v ...interface{}) {
	w.logger.Error(fmt.Sprintf(format, v...))
}

var _ tef_grpctypes.Logger = (*wrapLogger)(nil)

func newStefExporter(logger *zap.Logger, cfg *Config) *stefExporter {
	return &stefExporter{
		logger:         logger,
		cfg:            cfg,
		sentPendingAck: map[uint64]*sortedbymetric.SortedTree{},
		stopped:        make(chan struct{}),
	}
}

func (s *stefExporter) Start(ctx context.Context, host component.Host) error {
	var err error
	if s.cfg.ArrowPath != "" {
		s.arrowFile, err = os.Create(s.cfg.ArrowPath)
		if err != nil {
			return err
		}
	}

	compression := pkg.CompressionNone
	if s.cfg.Compression == "zstd" {
		compression = pkg.CompressionZstd
	}

	if err := s.startGrpcClient(compression); err != nil {
		return err
	}

	var opts []config.Option
	if s.cfg.Compression == "zstd" {
		opts = append(opts, config.WithZstd())
	} else {
		opts = append(opts, config.WithNoZstd())
	}
	s.arrowProducer = arrow_record.NewProducerWithOptions(opts...)

	return nil
}

func (s *stefExporter) Shutdown(ctx context.Context) error {
	close(s.stopped)
	if s.arrowFile != nil {
		s.arrowFile.Close()
	}
	if s.grpcConn != nil {
		return s.grpcConn.Close()
	}
	return nil
}

func (s *stefExporter) pushMetrics(_ context.Context, md pmetric.Metrics) error {
	s.logger.Info(
		"MetricsExporter",
		zap.Int("resource metrics", md.ResourceMetrics().Len()),
		zap.Int("metrics", md.MetricCount()),
		zap.Int("data points", md.DataPointCount()),
	)

	converter := otlpconvert.NewOtlpToSortedTree()
	sorted, err := converter.FromOtlp(md.ResourceMetrics())
	if err != nil {
		return err
	}

	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	if s.arrowFile != nil {
		records, err := s.arrowProducer.BatchArrowRecordsFromMetrics(md)
		if err != nil {
			return err
		}

		arrowBytes, err := proto.Marshal(records)
		_, err = s.arrowFile.Write(arrowBytes)
	}

	if s.remoteWriter != nil {
		err := sorted.ToTef(s.remoteWriter)
		if err != nil {
			return err
		}

		s.ackMutex.Lock()
		s.lastSentRecordId = s.remoteWriter.RecordCount()
		if s.lastAckedRecordId >= s.lastSentRecordId {
			// Already received by destination and acknowledged. Can happen if onGrpcAck() is called
			// before we even manage to add it to the map here. No need to add it to the map anymore.
			// It is acknowledged and safe.
		} else {
			s.sentPendingAck[s.lastSentRecordId] = sorted
		}
		s.ackMutex.Unlock()
	}

	return nil
}

func (s *stefExporter) flusher() {
	timer := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-timer.C:
			s.writeMutex.Lock()
			err := s.remoteWriter.Flush()
			s.writeMutex.Unlock()
			if err != nil {
				log.Printf("Cannot send TEF data: %v\n", err)
			}
		case <-s.stopped:
			return
		}
	}
}

func (s *stefExporter) onGrpcAck(ackId uint64) error {
	s.ackMutex.Lock()
	defer s.ackMutex.Unlock()
	for ; s.lastAckedRecordId < ackId; s.lastAckedRecordId++ {
		delete(s.sentPendingAck, s.lastAckedRecordId)
	}
	return nil
}

func (s *stefExporter) startGrpcClient(compression pkg.Compression) error {
	// Connect to the server.
	var err error
	s.grpcConn, err = grpc.NewClient(
		fmt.Sprintf(s.cfg.Endpoint),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	// Open a stream to the server.
	grpcClient := stef_proto.NewSTEFDestinationClient(s.grpcConn)

	schema, err := oteltef.MetricsWireSchema()
	if err != nil {
		return err
	}

	settings := tef_grpc.ClientSettings{
		Logger:       &wrapLogger{s.logger},
		GrpcClient:   grpcClient,
		ClientSchema: schema,
		Callbacks: tef_grpc.ClientCallbacks{
			OnAck: s.onGrpcAck,
		},
	}
	client := tef_grpc.NewClient(settings)

	grpcWriter, opts, err := client.Connect(context.Background())

	opts.Compression = compression

	// Create a byte writer over the stream.
	//grpcWriter, err := tef_grpc.NewGrpcWriter(context.Background(), grpcClient, s.onGrpcAck)
	if err != nil {
		return err
	}

	// Create record writer.
	s.remoteWriter, err = oteltef.NewMetricsWriter(grpcWriter, opts)
	if err != nil {
		return err
	}
	go s.flusher()

	return err
}

func clearUnsupportedMetricTypes(md pmetric.Metrics) {
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sms := rm.ScopeMetrics().At(j)
			sms.Metrics().RemoveIf(
				func(metric pmetric.Metric) bool {
					switch metric.Type() {
					case pmetric.MetricTypeGauge:
						return false
					case pmetric.MetricTypeSum:
						return false
					default:
						return true
					}
				},
			)
		}
	}

}
