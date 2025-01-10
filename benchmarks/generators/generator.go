package generators

import (
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type Generator interface {
	GetName() string
	Generate() pmetric.Metrics
}
