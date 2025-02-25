// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package stefreceiver

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configgrpc"
)

// Config defines configuration for STEF receiver.
type Config struct {
	configgrpc.ServerConfig `mapstructure:",squash"`
}

var (
	_ component.Config = (*Config)(nil)
)

// Validate checks the receiver configuration is valid
func (cfg *Config) Validate() error {
	return nil
}
