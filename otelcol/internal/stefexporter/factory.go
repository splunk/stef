// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package stefexporter // import "github.com/splunk/stef/otelcol/internal/stefexporter"

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/splunk/stef/otelcol/internal/stefexporter/internal/metadata"
)

// The value of "type" key in configuration.
var componentType = component.MustNewType("stef")

// NewFactory creates a factory for Debug exporter
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		componentType,
		createDefaultConfig,
		exporter.WithMetrics(createMetricsExporter, metadata.MetricsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createMetricsExporter(ctx context.Context, set exporter.Settings, config component.Config) (
	exporter.Metrics, error,
) {
	cfg := config.(*Config)
	exporterLogger := createLogger(cfg, set.TelemetrySettings.Logger)
	stefexporter := newStefExporter(exporterLogger, cfg)
	return exporterhelper.NewMetricsExporter(
		ctx, set, config,
		stefexporter.pushMetrics,
		exporterhelper.WithStart(stefexporter.Start),
		exporterhelper.WithShutdown(stefexporter.Shutdown),
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithTimeout(exporterhelper.TimeoutSettings{Timeout: 0}),
	)
}

func createLogger(cfg *Config, logger *zap.Logger) *zap.Logger {
	core := zapcore.NewSamplerWithOptions(
		logger.Core(),
		1*time.Second,
		10,
		5,
	)

	return zap.New(core)
}
