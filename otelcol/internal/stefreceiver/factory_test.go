// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package stefreceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
}

func TestCreateMetric(t *testing.T) {
	factory := NewFactory()
	defaultGRPCSettings := &configgrpc.ServerConfig{
		NetAddr: confignet.AddrConfig{
			Endpoint:  "127.0.0.1:0",
			Transport: confignet.TransportTypeTCP,
		},
	}

	tests := []struct {
		name         string
		cfg          *Config
		wantStartErr bool
		wantErr      bool
		sink         consumer.Metrics
	}{
		{
			name: "default",
			cfg: &Config{
				ServerConfig: *defaultGRPCSettings,
			},
			sink: consumertest.NewNop(),
		},
		{
			name: "invalid_grpc_address",
			cfg: &Config{
				ServerConfig: configgrpc.ServerConfig{
					NetAddr: confignet.AddrConfig{
						Endpoint:  "327.0.0.1:1122",
						Transport: confignet.TransportTypeTCP,
					},
				},
			},
			wantStartErr: true,
			sink:         consumertest.NewNop(),
		},
	}
	ctx := context.Background()
	creationSet := receivertest.NewNopSettings()
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				mr, err := factory.CreateMetrics(ctx, creationSet, tt.cfg, tt.sink)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				require.NoError(t, err)
				if tt.wantStartErr {
					assert.Error(t, mr.Start(context.Background(), componenttest.NewNopHost()))
				} else {
					require.NoError(t, mr.Start(context.Background(), componenttest.NewNopHost()))
					assert.NoError(t, mr.Shutdown(context.Background()))
				}
			},
		)
	}
}
