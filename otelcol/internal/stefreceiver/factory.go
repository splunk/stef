// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package stefreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

// NewFactory creates a new OTLP receiver factory.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType("stef"),
		createDefaultConfig,
		receiver.WithMetrics(createMetrics, component.StabilityLevelAlpha),
	)
}

// createDefaultConfig creates the default configuration for receiver.
func createDefaultConfig() component.Config {
	grpcCfg := configgrpc.NewDefaultServerConfig()
	grpcCfg.NetAddr = confignet.NewDefaultAddrConfig()
	grpcCfg.NetAddr.Endpoint = "localhost:4320"
	grpcCfg.NetAddr.Transport = confignet.TransportTypeTCP
	grpcCfg.ReadBufferSize = 512 * 1024

	return &Config{
		ServerConfig: *grpcCfg,
	}
}

// createMetrics creates a metrics receiver based on provided config.
func createMetrics(
	_ context.Context,
	set receiver.Settings,
	cfg component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	return newStefReceiver(cfg.(*Config), &set, consumer)
}
