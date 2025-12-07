// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package stefexporter // import "github.com/splunk/stef/otelcol/internal/stefexporter"

import (
	"errors"

	"go.opentelemetry.io/collector/component"
)

// Config defines configuration for logging exporter.
type Config struct {
	Endpoint    string `mapstructure:"endpoint"`
	Compression string `mapstructure:"compression"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the exporter configuration is valid
func (cfg *Config) Validate() error {
	if cfg.Endpoint == "" {
		return errors.New("endpoint must be non-empty")
	}
	switch cfg.Compression {
	case "":
	case "zstd":
	default:
		return errors.New("invalid compression, only 'zstd' supported")
	}
	return nil
}
